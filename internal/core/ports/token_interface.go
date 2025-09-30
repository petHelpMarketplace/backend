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
	GenerateTokenPair(ctx context.Context, sp *domain.SpecialistProfileDTO) (*domain.TokensPair, string, error)
	ValidateToken(ctx context.Context, token string, isAccess bool) (jti string, userID string, err error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error
	BlacklistAccessToken(ctx context.Context, tokenString string) error
	IsAccessTokenBlacklisted(ctx context.Context, jti string) (bool, error)
}

type TokenRepository interface {
	SaveRefreshTokenState(ctx context.Context, jti string, userID string, expiry time.Time) error
	IsRefreshTokenValid(ctx context.Context, jti string, userID string) (bool, error)
	RevokeRefreshToken(ctx context.Context, jti string, userID string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID string) error
	BlacklistAccessToken(ctx context.Context, jti string, expiresAt time.Time) error
	IsAccessTokenBlacklisted(ctx context.Context, jti string) (bool, error)
}
