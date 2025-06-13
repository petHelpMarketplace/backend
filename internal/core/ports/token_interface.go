package ports

import (
	"context"
	"pethelp-backend/internal/core/domain"
	"time"

	"github.com/gin-gonic/gin"
)

type TokenHandlers interface {
	RefreshToken(c *gin.Context)
}

type TokenService interface {
	GenerateTokenPair(context.Context, *domain.Specialist) (*domain.TokensPair, error)
	ParseAccessToken(context.Context, string) (*domain.AccessTokenClaims, error)
	ParseRefreshToken(context.Context, string) (*domain.RefreshTokenClaims, error)
	RevokeToken(context.Context, string) error

	// ExtractIDFromToken(requestToken string, entitysecret string) (string, error)
}

type TokenRepository interface {
	SaveRefreshTokenState(ctx context.Context, jti string, userID string, expiry time.Time) error
	IsRefreshTokenValid(ctx context.Context, jti string, userID string) (bool, error)
	RevokeRefreshToken(ctx context.Context, jti string, userID string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID string) error
}
