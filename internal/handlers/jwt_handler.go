package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
)

type TokenHandlerImpl struct {
	specialistService ports.SpecialistService
	tokenService      ports.TokenService
	cookieManager     ports.CookieManager
	logger            *zap.Logger
}

var _ ports.TokenHandlers = (*TokenHandlerImpl)(nil)

func NewTokenHandler(specialistSrv ports.SpecialistService, tokenSrv ports.TokenService, cookieMngr ports.CookieManager, logger *zap.Logger) *TokenHandlerImpl {
	return &TokenHandlerImpl{
		specialistService: specialistSrv,
		tokenService:      tokenSrv,
		cookieManager:     cookieMngr,
		logger:            logger,
	}
}

// RefreshToken godoc
// @Summary      Update access and refresh tokens
// @Description  Exchanges a valid refresh token (from an HTTP-only cookie) for a new access token and a new refresh token. The used refresh token is revoked. This endpoint does not accept a request body.
// @Tags         Token
// @Produce      json
// @Success      200  {object}  domain.TokensPair "Successfully generated new token pair"
// @Failure      400  {object}  domain.ErrorResponse "Invalid request payload or malformed refresh token"
// @Failure      401  {object}  domain.ErrorResponse "Unauthorized: Invalid, expired, or malformed refresh token"
// @Failure      403  {object}  domain.ErrorResponse "Forbidden: Refresh token has been revoked or is otherwise not allowed"
// @Failure      500  {object}  domain.ErrorResponse "Internal server error"
// @Router       /token/refresh [post]
func (th *TokenHandlerImpl) RefreshToken(c *gin.Context) {

	cookieRefreshToken, err := th.cookieManager.Get(c, "refresh_token")
	if err != nil {
		th.logger.Error("failed to get refresh token from cookie", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Get cookie error.",
		})
		return
	}
	var refreshToken string
	if cookieRefreshToken != nil {
		token, ok := cookieRefreshToken.(string)
		if !ok {
			refreshToken = ""
		}
		refreshToken = token
	}

	// var dto RefreshReq
	// if err := c.ShouldBindJSON(&dto); err != nil {
	// 	th.logger.Warn("failed to bind refresh token request", zap.Error(err))
	// 	c.JSON(http.StatusBadRequest, domain.ErrorResponse{
	// 		Code:    http.StatusBadRequest,
	// 		Message: "invalid request: refresh_token is required and must be a valid JWT.",
	// 	})
	// 	return
	// }

	if refreshToken == "" {
		c.JSON(http.StatusForbidden, domain.ErrorResponse{
			Code:    http.StatusForbidden,
			Message: "Refresh token cookie not found or invalid.",
		})
		return
	}

	_, id, err := th.tokenService.ValidateToken(c.Request.Context(), refreshToken, false)
	if err != nil {
		var code int
		var message string
		logLevel := zap.WarnLevel
		switch {
		case errors.Is(err, domain.ErrTokenMalformed),
			errors.Is(err, domain.ErrTokenSignatureInvalid),
			errors.Is(err, domain.ErrTokenExpired),
			errors.Is(err, domain.ErrRefreshTokenNotValid),
			errors.Is(err, domain.ErrTokenInvalid):
			code = http.StatusUnauthorized
			message = "refresh token is invalid or has expired."
		case errors.Is(err, domain.ErrTokenRevoked),
			errors.Is(err, domain.ErrForbidden),
			errors.Is(err, domain.ErrUserIDMismatch):
			code = http.StatusForbidden
			message = "refresh token has been revoked or is not permitted."
		default:
			code = http.StatusInternalServerError
			message = "internal server error."
			logLevel = zap.ErrorLevel
		}
		if ce := th.logger.Check(logLevel, "refresh token validation failed"); ce != nil {
			ce.Write(zap.Error(err), zap.String("reason", message))
		}
		c.JSON(code, domain.ErrorResponse{Code: code, Message: message})
		return
	}

	userID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		th.logger.Error("failed to parse userID from token claims", zap.String("userID_claim", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "internal server error.",
		})
		return
	}

	cookieUserID, err := th.cookieManager.Get(c, "user_id")
	if err != nil {
		th.logger.Error("failed to get user_id from cookie", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Get cookie error.",
		})
		return
	}
	if cookieUserID != nil {
		usr, ok := cookieUserID.(int64)
		if !ok {
			th.logger.Error("error cast interface type",
				zap.Any("type", fmt.Sprintf("%T", cookieUserID)))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Code:    http.StatusForbidden,
				Message: "Internal server error.",
			})
			return
		}
		if usr != userID {
			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Code:    http.StatusForbidden,
				Message: "UserID cookie mismatch.",
			})
			return
		}
	}

	specialistDTO, err := th.specialistService.ShowByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			th.logger.Warn("user account not found for valid refresh token", zap.Int64("userID", userID))
			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Code:    http.StatusForbidden,
				Message: "user associated with token no longer exists.",
			})
			return
		}
		th.logger.Error("failed to retrieve specialist for token refresh", zap.Int64("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "internal server error.",
		})
		return
	}

	tokens, jti, err := th.tokenService.GenerateTokenPair(c.Request.Context(), &specialistDTO)
	if err != nil {
		th.logger.Error("failed to generate new token pair", zap.Int64("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "internal server error.",
		})
		return
	}

	if err := th.tokenService.RevokeRefreshToken(c.Request.Context(), refreshToken); err != nil {
		th.logger.Error("failed to revoke used refresh token after issuing a new pair",
			zap.Int64("userID", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "failed to finalize token refresh. Please log in again.",
		})
		return
	}

	var sessionID string
	cookieSessionID, err := th.cookieManager.Get(c, "session_id")
	if err != nil {
		th.logger.Warn("failed to get session_id from cookie", zap.Error(err))
	} else {
		if cookieSessionID != nil {
			sID, ok := cookieSessionID.(string)
			if !ok {
				th.logger.Warn("session_id in cookie is not a string", zap.Any("type", fmt.Sprintf("%T", cookieSessionID)))
			}
			sessionID = sID
		}
	}

	th.cookieManager.UpdateOptions(c)
	sessionValues := map[string]interface{}{
		"jti":           jti,
		"refresh_token": tokens.Refresh,
	}
	// Write session
	th.cookieManager.BulkSet(c, sessionValues)
	th.cookieManager.Save(c)
	if err != nil {
		th.logger.Error("failed to save refresh token cookie ", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	th.logger.Info("Session cookie updated",
		zap.String("session_id", sessionID),
		zap.Int64("user_id", userID),
		zap.String("jti", jti),
		zap.String("refresh_token", tokens.Refresh))

	c.JSON(http.StatusOK, tokens)
}
