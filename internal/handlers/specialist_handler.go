package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	operationSpHandler = "specialist_handler: "
)

type SpecialistHandlerImpl struct {
	validator         ports.SpecialistValidator
	specialistService ports.SpecialistService
	tokenService      ports.TokenService
	logger            *zap.Logger
}

var _ ports.SpecialistHandlers = (*SpecialistHandlerImpl)(nil)

func NewSpecialistHandler(specialistSrv ports.SpecialistService, tokenSrv ports.TokenService, validator ports.SpecialistValidator, logger *zap.Logger) *SpecialistHandlerImpl {
	return &SpecialistHandlerImpl{
		validator:         validator,
		specialistService: specialistSrv,
		tokenService:      tokenSrv,
		logger:            logger,
	}
}

type successRegistration struct {
	ID      string `json:"id" example:"1"`
	Message string `json:"message" default:"Registration successful"`
}

// @Summary Registration
// @Description New specialist registration
// @Tags Specialist
// @Accept       json
// @Produce      json
// @Param request body domain.RegistrationRequest true "Registration request body"
// @Success 201 {object} successRegistration "Sign-up succeeded"
// @Failure      400,409,500 {object} domain.ErrorResponse
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
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: message,
			Details: fieldErrors,
		})
		return
	}
	// — Run field‐validator on it
	if validationErrors := sh.validator.Validate(req); len(validationErrors) > 0 {
		sh.logger.Error("validation failed", zap.Any("errors", validationErrors))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "validation failed",
			Details: validationErrors,
		})
		return
	}

	//email uniqueness
	id, err := sh.specialistService.Registration(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrAccountAlreadyExists) {
			c.JSON(http.StatusConflict, domain.ErrorResponse{
				Code:    http.StatusConflict,
				Message: fmt.Sprintf("specialist email %s already used", req.Email),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		})
		return
	}

	// — Success response
	c.JSON(http.StatusCreated, successRegistration{
		ID:      strconv.FormatInt(id, 10),
		Message: "Registration successful",
	})

}

// LoginReq represents the request body for user login.
// @Description User login request payload
type LoginReq struct {
	// Email address of the user.
	// required: true
	// format: email
	// example: user@example.com
	Email string `json:"email" binding:"required,email" example:"user@example.com"`

	// Password for the user account.
	// required: true
	// example: MySecretPassword123!
	Password string `json:"password" binding:"required" example:"MySecretPassword123!"`
}

// @Summary Login
// @Description Login specialist
// @Tags Specialist
// @Accept       json
// @Produce      json
// @Param request body LoginReq true "Login request body"
// @Success 200 {object} domain.TokensPair "Login succeeded"
// @Failure      400,401,500 {object} domain.ErrorResponse
// @Router /specialist/login [post]
func (sh *SpecialistHandlerImpl) Login(c *gin.Context) {
	var loginData LoginReq

	//bind and validate JSON payload
	if err := c.ShouldBindJSON(&loginData); err != nil {
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

		sh.logger.Error("bindJSON failed", zap.Error(bindErr), zap.Any("details", fieldErrors))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
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
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "Invalid credentials",
			})
		} else if errors.Is(err, domain.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "Authorization email not found",
			})

		} else {
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
			})
		}
		return
	}

	tokens, err := sh.tokenService.GenerateTokenPair(c.Request.Context(), &spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	//return both tokens in the response body
	c.JSON(http.StatusOK, tokens)
}

// Me godoc
// @Summary      Get current specialist
// @Description  Get information about the currently authenticated specialist. Requires a valid Bearer token.
// @Tags         Specialist
// @Produce      json
// @Success      200  {object}  domain.SpecialistProfileDTO "Successfully retrieved specialist data"
// @Failure      401  {object}  domain.ErrorResponse "Unauthorized. The user is not authenticated."
// @Failure      404  {object}  domain.ErrorResponse "Specialist account associated with the token not found."
// @Failure      500  {object}  domain.ErrorResponse "Internal server error."
// @Router       /specialist/me [get]
// @Security 	 BearerAuth
func (sh *SpecialistHandlerImpl) Me(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		sh.logger.Warn("userID not found in context, middleware might not have run or failed")
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		sh.logger.Error("userID in context is not a string", zap.Any("type", fmt.Sprintf("%T", userIDRaw)))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		sh.logger.Error("failed to parse userID from context", zap.String("userID", userIDStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	specialist, err := sh.specialistService.ShowByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			sh.logger.Warn("specialist not found for ID from token", zap.Int64("userID", userID))
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Code: http.StatusNotFound, Message: "Specialist account not found"})
			return
		}
		sh.logger.Error("failed to get specialist by ID", zap.Int64("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Code: http.StatusInternalServerError, Message: "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, specialist)
}
