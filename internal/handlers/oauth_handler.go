package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"go.uber.org/zap"
)

const operationOAuthName = "oauth_handler:"

// OAuthHandlers OAuth2.0 services contains
type OAuthHandlersImpl struct {
	oauthService      ports.OAuthService
	specialistService ports.SpecialistService
	tokenService      ports.TokenService
	logger            *zap.Logger
}

var _ ports.OAuthHandlers = (*OAuthHandlersImpl)(nil)

// NewOAuthHandlers create new OAuthHandlers
func NewOAuthHandlers(oauth ports.OAuthService, specialist ports.SpecialistService, token ports.TokenService, logger *zap.Logger) *OAuthHandlersImpl {
	return &OAuthHandlersImpl{
		oauthService:      oauth,
		specialistService: specialist,
		tokenService:      token,
		logger:            logger,
	}
}

// @Summary OAuth2.0 Provider Callback
// @Description Handles the callback from an OAuth2.0 provider after user authentication.
// @Tags OAuth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth2.0 provider name (e.g., google, github)"
// @Success 201 {object} domain.TokensPair "Successfully authenticated and generated tokens"
// @Failure 404 {object} domain.ErrorResponse "Account with this email not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error or failed to complete OAuth2.0 authentication"
// @Router /oauth/{provider} [get]
// SignInWithProvider redirect to provider login page
func (oh *OAuthHandlersImpl) SignInWithProvider(c *gin.Context) {
	provider := c.Param("provider")
	fmt.Println(provider)
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func (oh *OAuthHandlersImpl) ProviderCallback(c *gin.Context) {
	provider := c.Param("provider")

	user, err := oh.oauthService.InitOAuth(c.Request.Context(), provider, c.Writer, c.Request)
	if err != nil {
		oauthErr := fmt.Errorf("%s failed to complete OAuth2.0 authentication: %w", operationOAuthName, err)
		oh.logger.Error("", zap.Error(oauthErr))

		oauthMessage := domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "failed to complete OAuth2.0 authentication",
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, oauthMessage)
		return
	}

	specialist, err := oh.specialistService.ShowByEmail(c.Request.Context(), user.Email)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, domain.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "account with this email not found",
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		})
		return
	}

	tokens, err := oh.tokenService.GenerateTokenPair(c.Request.Context(), &specialist)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	// — Success response
	c.JSON(http.StatusCreated, tokens)

}
