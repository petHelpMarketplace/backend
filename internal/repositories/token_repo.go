package repositories

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"pethelp-backend/internal/core/domain"
	redisDB "pethelp-backend/pkg/database/redis"
)

const (
	tokenKeyPrefix        = "token:"
	userSessionsKeyPrefix = "user:sessions:"
	operationToken        = "token_repo:"
)

type TokenRepoImpl struct {
	redis *redisDB.DB
}

func NewTokenRepository(db *redisDB.DB) *TokenRepoImpl {
	return &TokenRepoImpl{redis: db}
}

// SaveRefreshTokenState saves the state of a refresh token to Redis.
func (r *TokenRepoImpl) SaveRefreshTokenState(ctx context.Context, jti string, userID string, expiry time.Time) error {
	tokenKey := tokenKeyPrefix + jti
	sessionKey := userSessionsKeyPrefix + userID

	pipe := r.redis.Client().Pipeline()

	// Store token details as a Hash
	pipe.HSet(ctx, tokenKey,
		"user_id", userID,
		"expiry", expiry.Unix(), // Store timestamp
		"revoked", false,
	)
	pipe.ExpireAt(ctx, tokenKey, expiry) // Set TTL for the token key

	// Add JTI to the user's active sessions set
	pipe.SAdd(ctx, sessionKey, jti)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to save refresh token state: %w", operationToken, err)
	}
	return nil
}

// IsRefreshTokenValid checks if a refresh token is valid and not revoked in Redis.
func (r *TokenRepoImpl) IsRefreshTokenValid(ctx context.Context, jti string, userID string) (bool, error) {

	tokenKey := tokenKeyPrefix + jti

	tokenState, err := r.redis.Client().HGetAll(ctx, tokenKey).Result()
	if err != nil {
		return false, fmt.Errorf("%s failed to retrieve token state: %w", operationToken, err)
	}

	if len(tokenState) == 0 {
		return false, domain.ErrRefreshTokenNotFound
	}

	storedUserID := tokenState["user_id"]
	if storedUserID != userID {
		return false, fmt.Errorf("%w for token JTI %s: expected %s, got %s", domain.ErrUserIDMismatch, jti, userID, storedUserID)
	}

	revokedStr, ok := tokenState["revoked"]
	if !ok {
		return false, fmt.Errorf("%w: 'revoked' status field missing for token JTI %s", domain.ErrRevokedStatusParseFail, jti)
	}

	isRevoked, err := strconv.ParseBool(revokedStr)
	if err != nil {
		return false, fmt.Errorf("%w: failed to parse revoked status for JTI %s: %v", domain.ErrRevokedStatusParseFail, jti, err)
	}
	if isRevoked {
		return false, domain.ErrTokenRevoked
	}

	sessionKey := userSessionsKeyPrefix + userID
	isMember, err := r.redis.Client().SIsMember(ctx, sessionKey, jti).Result()
	if err != nil {
		return false, fmt.Errorf("%w: failed to check session membership: %v", domain.ErrSessionMembershipFail, err)
	}
	if !isMember {
		return false, domain.ErrJTIInUserSessionsNotFound
	}

	return true, nil
}

// RevokeRefreshToken marks a specific refresh token as revoked.
func (r *TokenRepoImpl) RevokeRefreshToken(ctx context.Context, jti string, userID string) error {
	tokenKey := tokenKeyPrefix + jti
	sessionKey := userSessionsKeyPrefix + userID

	pipe := r.redis.Client().Pipeline()
	pipe.HSet(ctx, tokenKey, "revoked", true) // Mark as revoked
	pipe.SRem(ctx, sessionKey, jti)           // Remove from user's active sessions
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to revoke refresh token: %w", operationToken, err)
	}
	return nil
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a given user.
func (r *TokenRepoImpl) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	sessionKey := userSessionsKeyPrefix + userID

	// Get all JTIs for the user
	jtis, err := r.redis.Client().SMembers(ctx, sessionKey).Result()
	if err != nil {
		return fmt.Errorf("%s failed to get user JTIs for revocation: %w", operationToken, err)
	}

	pipe := r.redis.Client().Pipeline()
	for _, jti := range jtis {
		pipe.HSet(ctx, tokenKeyPrefix+jti, "revoked", true)
	}
	pipe.Del(ctx, sessionKey)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to revoke all user tokens: %w", operationToken, err)
	}
	return nil
}
