package services

import (
	"context"
	"fmt"

	"pethelp-backend/internal/config"
	"pethelp-backend/internal/core/ports"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

const operationName = "oauth_token_service:"

type OAuthServiceImpl struct {
	// oauthRepo     ports.TokenRepository
	oauthProvider goth.Provider
}

var _ ports.OAuthService = (*OAuthServiceImpl)(nil)

func NewOAuthService(repo ports.TokenRepository, googleOAuthConf *config.GoogleOAuthConfig) *OAuthServiceImpl {

	scopes := []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"}

	googleProvider := google.New(googleOAuthConf.ClientID, googleOAuthConf.ClientSecret, googleOAuthConf.ClientCallbackURL, scopes...)
	googleProvider.SetAccessType("offline")
	googleProvider.SetPrompt("consent")
	googleProvider.SetName("google")

	goth.UseProviders(googleProvider)

	return &OAuthServiceImpl{oauthProvider: googleProvider}
}

// InitAuth method save user token data to Redis DB
func (s *OAuthServiceImpl) InitAuth(ctx context.Context, user *goth.User) error {
	// Save or overwrite the user data
	// err := s.oauthRepo.Set(ctx, user)
	// if err != nil {
	// 	return fmt.Errorf("%s failed to save/update token: %w", operationName, err)
	// }
	return nil
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
