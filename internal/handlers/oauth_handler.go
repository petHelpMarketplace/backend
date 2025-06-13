package handlers

import (
	"fmt"
	"net/http"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"go.uber.org/zap"
)

const operationName = "oauth_handler:"

// OAuthHandlers OAuth2.0 services contains
type OAuthHandlersImpl struct {
	OAuthService ports.OAuthService
	Logger       *zap.Logger
}

var _ ports.OAuthHandlers = (*OAuthHandlersImpl)(nil)

// NewOAuthHandlers create new OAuthHandlers
func NewOAuthHandlers(service ports.OAuthService, logger *zap.Logger) *OAuthHandlersImpl {
	return &OAuthHandlersImpl{OAuthService: service, Logger: logger}
}

// SignInWithProvider redirect to provider login page
func (h *OAuthHandlersImpl) SignInWithProvider(c *gin.Context) {
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// CallbackHandler process provider callback and save user data
func (h *OAuthHandlersImpl) ProviderCallback(c *gin.Context) {
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		messErr := fmt.Errorf("%s failed to complete OAuth2.0 authentication: %w", operationName, err)
		h.Logger.Error("", zap.Error(messErr))

		errMessage := domain.RequestResponse{
			Code:    http.StatusInternalServerError,
			Type:    "OAuth error",
			Message: messErr.Error(),
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errMessage)
		return
	}

	err = h.OAuthService.InitAuth(c.Request.Context(), &user)
	if err != nil {
		messErr := fmt.Errorf("%s failed to save auth data: %w", operationName, err)
		h.Logger.Error("", zap.Error(messErr))

		errMessage := domain.RequestResponse{
			Code:    http.StatusInternalServerError,
			Type:    "DB error",
			Message: messErr.Error(),
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, errMessage)
		return
	}

	tokenData := domain.TokensPair{
		Access:  user.AccessToken,
		Refresh: user.RefreshToken,
	}

	// — Success response
	c.JSON(http.StatusCreated, tokenData)

}
