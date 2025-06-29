package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	operationTokenHandler = "token_handler: "
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
// @Description  Exchanges a valid refresh token for a new access token and a new refresh token.
// @Tags         Token
// @Accept       json
// @Produce      json
// @Param        request body RefreshReq true "Refresh token request"
// @Success      200  {object}  domain.TokensPair "Successfully generated new token pair"
// @Failure      400  {object}  domain.ErrorResponse "Invalid request payload or malformed refresh token"
// @Failure      401  {object}  domain.ErrorResponse "Unauthorized: Invalid refresh token signature or expired"
// @Failure      403  {object}  domain.ErrorResponse "Forbidden: Refresh token has been revoked"
// @Failure      500  {object}  domain.ErrorResponse "Internal server error due to token validation, database lookup, or token generation/revocation issues"
// @Router       /token/refresh [post]
func (th *TokenHandlerImpl) RefreshToken(c *gin.Context) {

	var dto RefreshReq
	// Bind and validate JSON payload
	if err := c.ShouldBindJSON(&dto); err != nil {
		bindErr := fmt.Errorf("%s invalid refresh payload %w", operationTokenHandler, err)
		th.logger.Error("bindJSON failed", zap.Error(bindErr))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid registration payload",
		})
		return
	}

	_, id, err := th.tokenService.ValidateToken(c.Request.Context(), dto.RefreshToken, false)
	if err != nil {
		if errors.Is(err, domain.ErrTokenInvalid) {
			validateErr := fmt.Errorf("%s invalid refresh token %w", operationTokenHandler, err)
			th.logger.Error("token failed", zap.Error(validateErr))
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "invalid refresh token",
			})
			return
		} else if errors.Is(err, domain.ErrTokenRevoked) {

			validateErr := fmt.Errorf("%s revoked refresh token %w", operationTokenHandler, err)
			th.logger.Error("token failed", zap.Error(validateErr))
			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Code:    http.StatusForbidden,
				Message: "revoked refresh token",
			})
			return

		} else {
			validateErr := fmt.Errorf("%s internal refresh token error %w", operationTokenHandler, err)
			th.logger.Error("token failed", zap.Error(validateErr))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "refresh token validation error",
			})
			return
		}
	}

	userID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		parseErr := fmt.Errorf("%s parse int failed %w", operationTokenHandler, err)
		th.logger.Error("parse failed", zap.Error(parseErr))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		})
		return
	}

	specialist, err := th.specialistService.ShowByID(c.Request.Context(), userID)
	if err != nil {
		showErr := fmt.Errorf("%s failed to show specialist from DB %w", operationTokenHandler, err)
		th.logger.Error("show failed", zap.Error(showErr))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		})
		return
	}

	tokens, err := th.tokenService.GenerateTokenPair(c.Request.Context(), &specialist)
	if err != nil {
		tokenErr := fmt.Errorf("%s failed to generate tokens %w", operationTokenHandler, err)
		th.logger.Error("generate failed", zap.Error(tokenErr))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		})
		return
	}

	err = th.tokenService.RevokeToken(c.Request.Context(), dto.RefreshToken)
	if err != nil {
		revokeErr := fmt.Errorf("%s failed to revoke tokens %w", operationTokenHandler, err)
		th.logger.Error("generate failed", zap.Error(revokeErr))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "revoke old token failed",
		})
		return
	}

	// Return new tokens
	c.JSON(http.StatusOK, tokens)
}
