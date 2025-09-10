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
	rtDetailsPrefix            = "rt_details:"   // Used for refresh token details (Hash)
	userActiveRTSessionsPrefix = "user_rts:"     // Used for user's active refresh token JTIs (Set)
	atBlacklistPrefix          = "at_blacklist:" // New prefix for blacklisted access token JTIs (String/Set)
	operationToken             = "token_repo:"
)

type TokenRepoImpl struct {
	redis *redisDB.DB
}

func NewTokenRepository(db *redisDB.DB) *TokenRepoImpl {
	return &TokenRepoImpl{redis: db}
}

// SaveRefreshTokenState saves the state of a refresh token to Redis.
// Stores token details in a HASH and adds its JTI to a user's SADD set.
func (r *TokenRepoImpl) SaveRefreshTokenState(ctx context.Context, jti string, userID string, expiry time.Time) error {
	rtKey := rtDetailsPrefix + jti
	userSessionsKey := userActiveRTSessionsPrefix + userID

	pipe := r.redis.Client().Pipeline()

	// Store refresh token details in a Redis Hash
	pipe.HSet(ctx, rtKey,
		"user_id", userID,
		"expires_at", expiry.Unix(), // Store Unix timestamp
		"revoked", false, // Initial state is not revoked
	)
	// Set TTL for the refresh token details key matching its expiry
	pipe.ExpireAt(ctx, rtKey, expiry)

	// Add JTI to the user's active sessions Set
	pipe.SAdd(ctx, userSessionsKey, jti)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("%s SaveRefreshTokenState: failed to execute Redis pipeline: %w", operationToken, err)
	}
	return nil
}

// IsRefreshTokenValid checks if a refresh token is valid and not revoked in Redis.
func (r *TokenRepoImpl) IsRefreshTokenValid(ctx context.Context, jti string, userID string) (bool, error) {

	tokenKey := rtDetailsPrefix + jti

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

	sessionKey := userActiveRTSessionsPrefix + userID
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
	tokenKey := rtDetailsPrefix + jti
	sessionKey := userActiveRTSessionsPrefix + userID

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
	sessionKey := userActiveRTSessionsPrefix + userID

	// Get all JTIs for the user
	jtis, err := r.redis.Client().SMembers(ctx, sessionKey).Result()
	if err != nil {
		return fmt.Errorf("%s failed to get user JTIs for revocation: %w", operationToken, err)
	}

	pipe := r.redis.Client().Pipeline()
	for _, jti := range jtis {
		pipe.HSet(ctx, rtDetailsPrefix+jti, "revoked", true)
	}
	pipe.Del(ctx, sessionKey)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to revoke all user tokens: %w", operationToken, err)
	}
	return nil
}

// BlacklistAccessToken adds an access token's JTI to a blacklist in Redis.
// The entry expires when the access token itself would naturally expire.
func (r *TokenRepoImpl) BlacklistAccessToken(ctx context.Context, jti string, expiresAt time.Time) error {

	blacklistKey := atBlacklistPrefix + jti

	// Store the JTI in Redis with an expiry that matches the token's expiry
	pipe := r.redis.Client().Pipeline()
	pipe.Set(ctx, blacklistKey, "revoked", 0)
	pipe.ExpireAt(ctx, blacklistKey, expiresAt)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to blacklist access token %s: %w", operationToken, jti, err)
	}
	return nil
}

// IsAccessTokenBlacklisted checks if an access token's JTI exists in the blacklist.
func (r *TokenRepoImpl) IsAccessTokenBlacklisted(ctx context.Context, jti string) (bool, error) {

	blacklistKey := atBlacklistPrefix + jti

	// Check if the key exists in Redis
	exists, err := r.redis.Client().Exists(ctx, blacklistKey).Result()
	if err != nil {
		return false, fmt.Errorf("%s failed to check access token blacklist status for %s: %w", operationToken, jti, err)
	}
	return exists > 0, nil
}
