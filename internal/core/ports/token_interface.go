package ports

import (
	"context"
	"pethelp-backend/internal/core/domain"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
)

type TokenHandlers interface {
	RefreshToken(c *gin.Context)
}

type TokenService interface {
	GenerateTokenPair(ctx context.Context, sp *domain.SpecialistProfileDTO) (*domain.TokensPair, error)
	ValidateToken(ctx context.Context, token string, isAccess bool) (jti string, userID string, err error)
	RevokeToken(ctx context.Context, token string) error

	// ExtractIDFromToken(requestToken string, entitysecret string) (string, error)
}

type TokenRepository interface {
	SaveRefreshTokenState(ctx context.Context, jti string, userID string, expiry time.Time) error
	IsRefreshTokenValid(ctx context.Context, jti string, userID string) (bool, error)
	RevokeRefreshToken(ctx context.Context, jti string, userID string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID string) error
	Set(ctx context.Context, user *goth.User) error
}
