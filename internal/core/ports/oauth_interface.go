package ports

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
)

type OAuthHandlers interface {
	SignInWithProvider(c *gin.Context)
	ProviderCallback(c *gin.Context)
}

type OAuthService interface {
	InitOAuth(ctx context.Context, provider string, wr http.ResponseWriter, req *http.Request) (*goth.User, error)
}
