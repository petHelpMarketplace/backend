package ports

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
)

type OAuthHandlers interface {
	SignInWithProvider(c *gin.Context)
	ProviderCallback(c *gin.Context)
}

type OAuthService interface {
	InitAuth(ctx context.Context, user *goth.User) error
	VerifyAuth(ctx context.Context, identifier string) (*goth.User, error)
	RefreshAuth(ctx context.Context, identifier string) (string, error)
	RevokeAuth(ctx context.Context, c *gin.Context, identifier string) error
}
