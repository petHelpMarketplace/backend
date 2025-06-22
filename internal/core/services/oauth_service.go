package services

import (
	"context"
	"fmt"
	"net/http"

	"pethelp-backend/internal/config"
	"pethelp-backend/internal/core/ports"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"go.uber.org/zap"
)

const operationName = "oauth_token_service:"

type OAuthServiceImpl struct {
	tokenRepo     ports.TokenRepository
	oauthProvider goth.Provider
	logger        *zap.Logger
}

var _ ports.OAuthService = (*OAuthServiceImpl)(nil)

func NewOAuthService(repo ports.TokenRepository, googleOAuthConf *config.GoogleOAuthConfig, logger *zap.Logger) *OAuthServiceImpl {

	scopes := []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"}

	googleProvider := google.New(googleOAuthConf.ClientID, googleOAuthConf.ClientSecret, googleOAuthConf.ClientCallbackURL, scopes...)
	googleProvider.SetAccessType("offline")
	googleProvider.SetPrompt("consent")
	googleProvider.SetName("google")

	goth.UseProviders(googleProvider)

	return &OAuthServiceImpl{tokenRepo: repo, oauthProvider: googleProvider, logger: logger}
}

// InitAuth method save user token data to Redis DB
func (os *OAuthServiceImpl) InitOAuth(ctx context.Context, provider string, wr http.ResponseWriter, req *http.Request) (*goth.User, error) {

	q := req.URL.Query()
	q.Add("provider", provider)
	req.URL.RawQuery = q.Encode()

	user, err := gothic.CompleteUserAuth(wr, req)
	if err != nil {
		messErr := fmt.Errorf("%s failed to complete OAuth2.0 authentication: %w", operationName, err)
		os.logger.Error("OAuth error", zap.Error(messErr))

		return nil, err
	}

	return &user, nil
}

// VerifyAuth method get token data from Redis DB
func (s *OAuthServiceImpl) VerifyAuth(ctx context.Context, identifier string) (*goth.User, error) {
	// user, err := s.oauthRepo.Get(ctx, identifier)
	// if err != nil {
	// 	return nil, fmt.Errorf("%s failed to verify token: %w", operationName, err)
	// }
	return &goth.User{}, nil
}

func (s *OAuthServiceImpl) RefreshAuth(ctx context.Context, token string) (string, error) {
	if s.oauthProvider.RefreshTokenAvailable() {
		token, err := s.oauthProvider.RefreshToken(token)
		if err != nil {
			return "", fmt.Errorf("%s failed to refresh token: %w", operationName, err)
		}
		// err = s.oauthRepo.SetToken(ctx, user)
		// if err != nil {
		// 	return "", fmt.Errorf("%s failed to save/update token: %w", operationName, err)
		// }
		return token.AccessToken, nil
	} else {
		return "", fmt.Errorf("%s refresh token not available for this provider", operationName)
	}
}

func (s *OAuthServiceImpl) RevokeAuth(ctx context.Context, c *gin.Context, identifier string) error {
	// err := gothic.Logout(c.Writer, c.Request)
	// if err != nil {
	// 	return fmt.Errorf("%s failed to revoke token: %w", operationName, err)
	// }
	// err = s.oauthRepo.Del(ctx, identifier)
	// if err != nil {
	// 	return fmt.Errorf("%s failed to delete token: %w", operationName, err)
	// }
	return nil
}
