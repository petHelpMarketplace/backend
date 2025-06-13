package handlers

import (
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

type RefreshDTO struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshHandler returns a Gin handler that processes refresh token requests.
// It reads the old refresh token and returns new pair.
func (th *TokenHandlerImpl) RefreshToken(c *gin.Context) {

	var dto RefreshDTO
	// Bind and validate JSON payload
	if err := c.ShouldBindJSON(&dto); err != nil {
		bindErr := fmt.Errorf("%s invalid refresh payload %w", operationTokenHandler, err)
		th.logger.Error("bindJSON failed", zap.Error(bindErr))
		c.JSON(http.StatusBadRequest, domain.RequestResponse{
			Code:    http.StatusBadRequest,
			Type:    "input error",
			Message: "invalid registration payload",
		})
		return
	}

	refreshClaims, err := th.tokenService.ParseRefreshToken(c.Request.Context(), dto.RefreshToken)
	if err != nil {
		bindErr := fmt.Errorf("%s invalid or expired refresh token %w", operationTokenHandler, err)
		th.logger.Error("token failed", zap.Error(bindErr))
		c.JSON(http.StatusUnauthorized, domain.RequestResponse{
			Code:    http.StatusUnauthorized,
			Type:    "token error",
			Message: "invalid or expired refresh token",
		})
		return
	}

	userID, err := strconv.ParseInt(refreshClaims.UserID, 10, 64)
	if err != nil {
		parseErr := fmt.Errorf("%s parse int failed %w", operationTokenHandler, err)
		th.logger.Error("parse failed", zap.Error(parseErr))
		c.JSON(http.StatusInternalServerError, domain.RequestResponse{
			Code:    http.StatusInternalServerError,
			Type:    "server error",
			Message: " parse int failed",
		})
		return
	}

	specialist, err := th.specialistService.Show(c.Request.Context(), userID)
	if err != nil {
		showErr := fmt.Errorf("%s failed to show specialist from DB %w", operationTokenHandler, err)
		th.logger.Error("show failed", zap.Error(showErr))
		c.JSON(http.StatusInternalServerError, domain.RequestResponse{
			Code:    http.StatusInternalServerError,
			Type:    "db error",
			Message: "Internal server error",
		})
		return
	}

	tokens, err := th.tokenService.GenerateTokenPair(c.Request.Context(), &specialist)
	if err != nil {
		tokenErr := fmt.Errorf("%s failed to generate tokens %w", operationTokenHandler, err)
		th.logger.Error("generate failed", zap.Error(tokenErr))
		c.JSON(http.StatusInternalServerError, domain.RequestResponse{
			Code:    http.StatusInternalServerError,
			Type:    "token error",
			Message: "Internal server error",
		})
		return
	}

	// Return new tokens
	c.JSON(http.StatusOK, tokens)
}
