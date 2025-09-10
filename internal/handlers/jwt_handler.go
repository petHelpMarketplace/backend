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

	var refreshToken string
	cookieRefreshToken, err := th.cookieManager.Get(c, "refresh_token")
	if err == nil && cookieRefreshToken != nil {
		switch v := cookieRefreshToken.(type) {
		case string:
			refreshToken = v
		case []byte:
			refreshToken = string(v)
		default:
			th.logger.Warn("refresh_token cookie has unexpected type", zap.String("type", fmt.Sprintf("%T", cookieRefreshToken)))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Get cookie error.",
			})
			return
		}
	} else if err != nil {
		th.logger.Debug("refresh_token cookie not present", zap.Error(err))
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Code:    http.StatusUnauthorized,
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
		th.logger.Warn("user_id cookie not present", zap.Error(err))
	} else if cookieUserID != nil {
		var usrFromCookie int64
		switch v := cookieUserID.(type) {
		case int64:
			usrFromCookie = v
		case string:
			if parsed, perr := strconv.ParseInt(v, 10, 64); perr != nil {
				th.logger.Error("failed to parse user_id cookie", zap.String("value", v), zap.Error(perr))
				c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error.",
				})
				return
			} else {
				usrFromCookie = parsed
			}
		case []byte:
			if parsed, perr := strconv.ParseInt(string(v), 10, 64); perr != nil {
				th.logger.Error("failed to parse user_id cookie", zap.String("value", string(v)), zap.Error(perr))
				c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error.",
				})
				return
			} else {
				usrFromCookie = parsed
			}
		default:
			th.logger.Error("unexpected user_id cookie type", zap.String("type", fmt.Sprintf("%T", cookieUserID)))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error.",
			})
			return
		}
		if usrFromCookie != userID {
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

	th.cookieManager.ResetOptions(c)
	sessionValues := map[string]interface{}{
		"jti":           jti,
		"refresh_token": tokens.Refresh,
	}
	// Write session
	th.cookieManager.BulkSet(c, sessionValues)
	if err := th.cookieManager.Save(c); err != nil {
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
		zap.String("jti", jti))

	c.JSON(http.StatusOK, tokens)
}
