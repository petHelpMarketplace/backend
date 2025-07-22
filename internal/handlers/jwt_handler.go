package handlers

import (
	"errors"
	"net/http"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TokenHandlerImpl struct {
	specialistService ports.SpecialistService
	tokenService      ports.TokenService
	logger            *zap.Logger
}

var _ ports.TokenHandlers = (*TokenHandlerImpl)(nil)

func NewTokenHandler(specialistSrv ports.SpecialistService, tokenSrv ports.TokenService, logger *zap.Logger) *TokenHandlerImpl {
	return &TokenHandlerImpl{
		specialistService: specialistSrv,
		tokenService:      tokenSrv,
		logger:            logger,
	}
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required" validate:"jwt"`
}

// RefreshToken godoc
// @Summary      Update access and refresh tokens
// @Description  Exchanges a valid refresh token for a new access token and a new refresh token. The used refresh token is revoked.
// @Tags         Token
// @Accept       json
// @Produce      json
// @Param        request body RefreshReq true "Refresh token request"
// @Success      200  {object}  domain.TokensPair "Successfully generated new token pair"
// @Failure      400  {object}  domain.ErrorResponse "Invalid request payload or malformed refresh token"
// @Failure      401  {object}  domain.ErrorResponse "Unauthorized: Invalid, expired, or malformed refresh token"
// @Failure      403  {object}  domain.ErrorResponse "Forbidden: Refresh token has been revoked or is otherwise not allowed"
// @Failure      500  {object}  domain.ErrorResponse "Internal server error"
// @Router       /token/refresh [post]
func (th *TokenHandlerImpl) RefreshToken(c *gin.Context) {
	var dto RefreshReq
	if err := c.ShouldBindJSON(&dto); err != nil {
		th.logger.Warn("failed to bind refresh token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid request: refresh_token is required and must be a valid JWT.",
		})
		return
	}

	_, id, err := th.tokenService.ValidateToken(c.Request.Context(), dto.RefreshToken, false)
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

	tokens, err := th.tokenService.GenerateTokenPair(c.Request.Context(), &specialistDTO)
	if err != nil {
		th.logger.Error("failed to generate new token pair", zap.Int64("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "internal server error.",
		})
		return
	}

	if err := th.tokenService.RevokeRefreshToken(c.Request.Context(), dto.RefreshToken); err != nil {
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

	c.JSON(http.StatusOK, tokens)
}
