package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
)

const (
	operationSpHandler = "specialist_handler: "
)

type SpecialistHandlerImpl struct {
	validator         ports.SpecialistValidator
	specialistService ports.SpecialistService
	tokenService      ports.TokenService
	cookieManager     ports.CookieManager
	logger            *zap.Logger
}

var _ ports.SpecialistHandlers = (*SpecialistHandlerImpl)(nil)

func NewSpecialistHandler(specialistSrv ports.SpecialistService, tokenSrv ports.TokenService, validator ports.SpecialistValidator, cookieMngr ports.CookieManager, logger *zap.Logger) *SpecialistHandlerImpl {
	return &SpecialistHandlerImpl{
		validator:         validator,
		specialistService: specialistSrv,
		tokenService:      tokenSrv,
		cookieManager:     cookieMngr,
		logger:            logger,
	}
}

type successRegistration struct {
	ID      int64  `json:"id" example:"100"`
	Message string `json:"message" default:"Registration successful"`
}

// @Summary Registration
// @Description New specialist registration
// @Tags Specialist
// @Accept       json
// @Produce      json
// @Param request body domain.RegistrationRequest true "Registration request body"
// @Success 201 {object} successRegistration "Sign-up succeeded"
// @Failure      400  {object}  domain.BadRequestError "Invalid request payload or validation failed"
// @Failure      409  {object}  domain.ConflictError "Account with this email already exists"
// @Failure      500  {object}  domain.InternalServerError "Internal server error"
// @Router /specialist/register [post]
func (sh *SpecialistHandlerImpl) Registration(c *gin.Context) {

	req := domain.RegistrationRequest{}

	if err := c.ShouldBindJSON(&req); err != nil {

		var fieldErrors []domain.FieldError
		message := "Invalid request payload"

		var jsonErr *json.UnmarshalTypeError
		var syntaxErr *json.SyntaxError

		bindErr := fmt.Errorf("%s invalid registration payload: %w", operationSpHandler, err)

		if errors.As(err, &jsonErr) {
			message = "The request contains invalid data types."
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:   jsonErr.Field,
				Message: fmt.Sprintf("Expected type '%s' for field.", jsonErr.Type),
			})
		} else if errors.As(err, &syntaxErr) {
			message = "The request body is not valid JSON."
		} else if err == io.EOF {
			message = "Request body cannot be empty."
		}

		sh.logger.Error("bind JSON failed", zap.Error(bindErr), zap.Any("details", fieldErrors))
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: message,
			Details: fieldErrors,
		})
		return
	}

	if validationErrors := sh.validator.ValidateRegistrationReq(req); len(validationErrors) > 0 {
		sh.logger.Error("validation failed", zap.Any("errors", validationErrors))
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: "validation failed",
			Details: validationErrors,
		})
		return
	}

	id, err := sh.specialistService.Registration(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrAccountAlreadyExists) {
			c.JSON(http.StatusConflict, domain.ConflictError{
				Code:    http.StatusConflict,
				Message: fmt.Sprintf("specialist email %s already used", req.Email),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		})
		return
	}

	sessionID := ulid.MustNew(ulid.Timestamp(time.Now()), ulid.DefaultEntropy()).String()

	// Write session
	sh.cookieManager.Set(c, "session_id", sessionID)
	sh.cookieManager.Set(c, "request_id", requestid.Get(c))
	err = sh.cookieManager.Save(c)
	if err != nil {
		sh.logger.Error("failed to save registration cookie ", zap.Error(err))
	}

	// — Success response
	c.JSON(http.StatusCreated, successRegistration{
		ID:      id,
		Message: "Registration successful",
	})

}

// @Summary Login
// @Description Login specialist
// @Tags Specialist
// @Accept       json
// @Produce      json
// @Param request body domain.LoginReq true "Login request body"
// @Success 200 {object} domain.TokensPair "Login succeeded"
// @Failure      400  {object}  domain.BadRequestError "Invalid request payload"
// @Failure      401  {object}  domain.UnauthorizedError "Invalid credentials"
// @Failure      404  {object}  domain.NotFoundError "Account not found"
// @Failure      500  {object}  domain.InternalServerError "Internal server error"
// @Router /specialist/login [post]
func (sh *SpecialistHandlerImpl) Login(c *gin.Context) {
	var loginData domain.LoginReq

	//bind and validate JSON payload
	if err := c.ShouldBindJSON(&loginData); err != nil {
		var fieldErrors []domain.FieldError
		message := "Invalid request payload"

		var jsonErr *json.UnmarshalTypeError
		var syntaxErr *json.SyntaxError

		bindErr := fmt.Errorf("%s invalid login request payload: %w", operationSpHandler, err)

		if errors.As(err, &jsonErr) {
			message = "The request contains invalid data types."
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:   jsonErr.Field,
				Message: fmt.Sprintf("Expected type '%s' for field.", jsonErr.Type),
			})
		} else if errors.As(err, &syntaxErr) {
			message = "The request body is not valid JSON."
		} else if err == io.EOF {
			message = "Request body cannot be empty."
		}

		sh.logger.Error("bindJSON failed", zap.Error(bindErr), zap.Any("details", fieldErrors))
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: message,
			Details: fieldErrors,
		})
		return
	}

	//perform authentication and token issuance
	spec, err := sh.specialistService.Login(c.Request.Context(), loginData.Email, loginData.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, domain.UnauthorizedError{
				Code:    http.StatusUnauthorized,
				Message: "Invalid credentials",
			})
		} else if errors.Is(err, domain.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, domain.NotFoundError{
				Code:    http.StatusNotFound,
				Message: "Authorization email not found",
			})

		} else {
			c.JSON(http.StatusInternalServerError, domain.InternalServerError{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
			})
		}
		return
	}

	tokens, jti, err := sh.tokenService.GenerateTokenPair(c.Request.Context(), &spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	sessionID := ulid.MustNew(ulid.Timestamp(time.Now()), ulid.DefaultEntropy()).String()
	sessionValues := map[string]interface{}{
		"session_id":    sessionID,
		"request_id":    requestid.Get(c),
		"user_id":       spec.ID,
		"jti":           jti,
		"refresh_token": tokens.Refresh,
	}

	// Write session
	sh.cookieManager.BulkSet(c, sessionValues)
	err = sh.cookieManager.Save(c)
	if err != nil {
		sh.logger.Error("failed to save login cookie ", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	sh.logger.Info("Session cookie set with",
		zap.String("session_id", sessionID),
		zap.Int64("user_id", spec.ID),
		zap.String("jti", jti))

	//return both tokens in the response body
	c.JSON(http.StatusOK, tokens)
}

// Me godoc
// @Summary      Get current specialist
// @Description  Get information about the currently authenticated specialist. Requires a valid Bearer token.
// @Tags         Specialist
// @Produce      json
// @Success      200  {object}  domain.SpecialistProfDTO "Successfully retrieved specialist profile"
// @Failure      401  {object}  domain.UnauthorizedError "Unauthorized: User is not authenticated"
// @Failure      404  {object}  domain.NotFoundError "Specialist account associated with the token not found"
// @Failure      500  {object}  domain.InternalServerError "Internal server error"
// @Router       /specialist/me [get]
// @Security 	 BearerAuth
func (sh *SpecialistHandlerImpl) Me(c *gin.Context) {

	userID, ok := getUserIDFromContext(c, sh.logger)
	if !ok {
		return
	}

	specialist, err := sh.specialistService.ShowByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			sh.logger.Warn("specialist not found for ID from token", zap.Int64("userID", userID))
			c.JSON(http.StatusNotFound, domain.NotFoundError{
				Code:    http.StatusNotFound,
				Message: "Specialist account not found",
			})
			return
		}
		sh.logger.Error("failed to get specialist by ID", zap.Int64("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, specialist)
}

// ChangePassword
// @Summary      Change specialist password
// @Description  Allows an authenticated specialist to change their password. All active sessions will be terminated upon successful password change.
// @Tags         Specialist
// @Accept       json
// @Produce      json
// @Param        request body domain.ChangePassReq true "Change password request"
// @Success      200  {object}  domain.SuccessResponse "Password updated successfully"
// @Failure      400  {object}  domain.BadRequestError "Invalid request payload or validation failed"
// @Failure      401  {object}  domain.UnauthorizedError "Unauthorized or invalid old password"
// @Failure      404  {object}  domain.NotFoundError "Specialist account not found"
// @Failure      409  {object}  domain.ConflictError "Conflict: New password is the same as the old one"
// @Failure      500  {object}  domain.InternalServerError "Internal server error"
// @Router       /specialist/change-password [patch]
// @Security 	 BearerAuth
func (sh *SpecialistHandlerImpl) ChangePassword(c *gin.Context) {

	userID, ok := getUserIDFromContext(c, sh.logger)
	if !ok {
		return
	}

	var reqData domain.ChangePassReq

	//bind and validate JSON payload
	if err := c.ShouldBindJSON(&reqData); err != nil {
		var fieldErrors []domain.FieldError
		message := "Invalid request payload"

		var jsonErr *json.UnmarshalTypeError
		var syntaxErr *json.SyntaxError

		bindErr := fmt.Errorf("%s invalid change password request payload: %w", operationSpHandler, err)

		if errors.As(err, &jsonErr) {
			message = "The request contains invalid data types."
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:   jsonErr.Field,
				Message: fmt.Sprintf("Expected type '%s' for field.", jsonErr.Type),
			})
		} else if errors.As(err, &syntaxErr) {
			message = "The request body is not valid JSON."
		} else if err == io.EOF {
			message = "Request body cannot be empty."
		}

		sh.logger.Error("bindJSON failed", zap.Error(bindErr), zap.Any("details", fieldErrors))
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: message,
			Details: fieldErrors,
		})
		return
	}

	if validationErrors := sh.validator.ValidateChangePasswordReq(reqData); len(validationErrors) > 0 {
		sh.logger.Error("validation failed", zap.Any("errors", validationErrors))
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: "validation failed",
			Details: validationErrors,
		})
		return
	}

	err := sh.specialistService.ChangePassword(c.Request.Context(), userID, reqData.CurrentPass, reqData.NewPass)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, domain.UnauthorizedError{
				Code:    http.StatusUnauthorized,
				Message: "Invalid old password",
			})
			return
		}
		if errors.Is(err, domain.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, domain.NotFoundError{
				Code:    http.StatusNotFound,
				Message: "Specialist account not found",
			})
			return
		}
		if errors.Is(err, domain.ErrPasswordReuse) {
			c.JSON(http.StatusConflict, domain.ConflictError{
				Code:    http.StatusConflict,
				Message: "New password cannot be the same as the old password.",
			})
			return
		}

		sh.logger.Error("failed to update password", zap.Int64("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	err = sh.tokenService.RevokeAllUserSessions(c.Request.Context(), strconv.FormatInt(userID, 10))
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Code:    http.StatusOK,
		Message: "Password changed successfully.",
	})

}

// Logout godoc
// @Summary      Logout specialist
// @Description  Logs out the specialist by blacklisting the current access token, revoking the refresh token and clearing the session cookie.
// @Tags         Specialist
// @Accept       json
// @Produce      json
// @Success      200  {object}  domain.SuccessResponse "Logout successful"
// @Failure      401  {object}  domain.UnauthorizedError "Unauthorized"
// @Failure      500  {object}  domain.InternalServerError "Internal server error during logout process"
// @Router       /specialist/logout [post]
// @Security 	 BearerAuth
func (sh *SpecialistHandlerImpl) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	authHeader := c.GetHeader("Authorization")
	if accessToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer ")); accessToken != "" {
		if err := sh.tokenService.BlacklistAccessToken(ctx, accessToken); err != nil {
			// Log the error but don't fail the request. The token will expire naturally.
			sh.logger.Error("failed to blacklist access token during logout", zap.Error(err))
		}
	}

	cookieRefreshToken, err := sh.cookieManager.Get(c, "refresh_token")
	if err == nil {
		if refreshToken, ok := cookieRefreshToken.(string); ok && refreshToken != "" {
			if err := sh.tokenService.RevokeRefreshToken(ctx, refreshToken); err != nil {
				sh.logger.Error("failed to revoke refresh token during logout", zap.Error(err))
			} else {
				sh.logger.Info("refresh token revoked successfully")
			}
		}
	} else {
		sh.logger.Warn("could not retrieve refresh token cookie during logout", zap.Error(err))
	}

	//Clear the session cookie on the client side
	if err := sh.cookieManager.Clear(c); err != nil {
		sh.logger.Error("failed to clear session cookie during logout", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to clear session. Please clear your browser cookies.",
		})
		return
	}

	userID, _ := c.Get("userID")
	sh.logger.Info("user logged out successfully", zap.Any("userID", userID))
	c.JSON(http.StatusOK, domain.SuccessResponse{
		Code:    http.StatusOK,
		Message: "Logout successful.",
	})
}

// updateProfileSuccessResponse defines the successful response for the profile update endpoint.
type updateProfileSuccessResponse struct {
	Code    int                      `json:"code" example:"200"`
	Message string                   `json:"message" example:"Profile updated successfully."`
	Data    domain.SpecialistProfDTO `json:"data"`
}

// UpdateProfile
// @Summary      Update specialist profile
// @Description  Allows an authenticated specialist to update their profile information (name, family_name, phone, experience_years, bio).
// @Tags         Specialist
// @Accept       json
// @Produce      json
// @Param        request body domain.SpecialistProfUpdateReq true "Specialist profile update request"
// @Success      200  {object}  updateProfileSuccessResponse "Profile updated successfully"
// @Failure      400  {object}  domain.BadRequestError "Invalid request payload or validation failed"
// @Failure      401  {object}  domain.UnauthorizedError "Unauthorized"
// @Failure      404  {object}  domain.NotFoundError "Specialist account not found"
// @Failure      500  {object}  domain.InternalServerError "Internal server error"
// @Router       /specialist/profile [patch]
// @Security 	 BearerAuth
func (sh *SpecialistHandlerImpl) UpdateProfile(c *gin.Context) {
	userID, ok := getUserIDFromContext(c, sh.logger)
	if !ok {
		return // getUserIDFromContext already handled the error response
	}

	var reqData domain.SpecialistProfUpdateReq

	if err := c.ShouldBindJSON(&reqData); err != nil {
		var fieldErrors []domain.FieldError
		message := "Invalid request payload"

		var jsonErr *json.UnmarshalTypeError
		var syntaxErr *json.SyntaxError

		bindErr := fmt.Errorf("%s invalid specialist profile update payload: %w", operationSpHandler, err)

		if errors.As(err, &jsonErr) {
			message = "The request contains invalid data types."
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:   jsonErr.Field,
				Message: fmt.Sprintf("Expected type '%s' for field.", jsonErr.Type),
			})
		} else if errors.As(err, &syntaxErr) {
			message = "The request body is not valid JSON."
		} else if err == io.EOF {
			message = "Request body cannot be empty."
		}

		sh.logger.Error("bind JSON failed", zap.Error(bindErr), zap.Any("details", fieldErrors))
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: message,
			Details: fieldErrors,
		})
		return
	}

	if validationErrors := sh.validator.ValidateSpecialistProfileUpdateReq(reqData); len(validationErrors) > 0 {
		sh.logger.Error("validation failed for specialist profile update", zap.Any("errors", validationErrors))
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: "validation failed",
			Details: validationErrors,
		})
		return
	}

	specProf, err := sh.specialistService.UpdateProfile(c.Request.Context(), userID, reqData)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, domain.NotFoundError{
				Code:    http.StatusNotFound,
				Message: "Specialist account not found",
			})
			return
		}
		sh.logger.Error("failed to update specialist profile", zap.Int64("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, updateProfileSuccessResponse{
		Code:    http.StatusOK,
		Message: "Profile updated successfully.",
		Data:    specProf,
	})
}

// DeactivateProfile
// @Summary      Deactivate or activate specialist profile
// @Description  Allows an authenticated specialist to deactivate or activate their profile.
// @Tags         Specialist
// @Accept       json
// @Produce      json
// @Param        request body object{is_active=bool} true "Profile activation/deactivation request"
// @Success      200  {object}  domain.SuccessResponse "Profile status updated successfully"
// @Failure      400  {object}  domain.BadRequestError "Invalid status parameter"
// @Failure      401  {object}  domain.UnauthorizedError "Unauthorized"
// @Failure      404  {object}  domain.NotFoundError "Specialist account not found"
// @Failure      500  {object}  domain.InternalServerError "Internal server error"
// @Router       /specialist/me/status [patch]
// @Security 	 BearerAuth
func (sh *SpecialistHandlerImpl) DeactivateProfile(c *gin.Context) {

	userID, ok := getUserIDFromContext(c, sh.logger)
	if !ok {
		return // getUserIDFromContext already handled the error response
	}

	var req struct {
		IsActive *bool `json:"is_active" binding:"required" example:"true"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.IsActive == nil {
		sh.logger.Error("Request body must include boolean field 'is_active'")
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: "Request body must include boolean field 'is_active'",
		})
		return
	}
	isActive := *req.IsActive

	err := sh.specialistService.DeactivateProfile(c.Request.Context(), userID, isActive)
	if err != nil {
		sh.logger.Error("failed to update specialist profile active status", zap.Int64("userID", userID), zap.Bool("status", isActive), zap.Error(err))
		if errors.Is(err, domain.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, domain.NotFoundError{
				Code:    http.StatusNotFound,
				Message: "Specialist account not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	message := "Profile deactivated successfully."
	if isActive {
		message = "Profile activated successfully."
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Code:    http.StatusOK,
		Message: message,
	})
}
