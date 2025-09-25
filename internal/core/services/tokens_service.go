package services

import (
	"context"
	"errors"
	"fmt"
	"pethelp-backend/internal/config"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	genJWT "pethelp-backend/pkg/jwtoken"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type TokenServiceImpl struct {
	tokenRepo      ports.TokenRepository
	accessExpMin   time.Duration
	refreshExpDays time.Duration
	jwtSecret      []byte
	logger         *zap.Logger
}

var _ ports.TokenService = (*TokenServiceImpl)(nil)

func NewTokenService(repo ports.TokenRepository, cfg config.AuthConfig, logger *zap.Logger) *TokenServiceImpl {
	return &TokenServiceImpl{
		tokenRepo:      repo,
		accessExpMin:   cfg.AccessTTL,
		refreshExpDays: cfg.RefreshTTL,
		jwtSecret:      cfg.JWTSecret,
		logger:         logger,
	}
}

func (ts *TokenServiceImpl) GenerateTokenPair(ctx context.Context, s *domain.SpecialistProfDTO) (*domain.TokensPair, string, error) {

	tokens := &domain.TokensPair{}

	id := strconv.FormatInt(s.ID, 10)
	roles := []string{"specialist"}
	tenant := "petHelp"

	accessToken, accessJTI, err := genJWT.GenerateAccessToken(id, roles, tenant, ts.jwtSecret, ts.accessExpMin)
	if err != nil {
		ts.logger.Error("failed to generate access token",
			zap.String("accessTokenID", accessJTI),
			zap.String("userID", id),
			zap.Error(err))
		return tokens, accessJTI, domain.ErrInternalServer
	}
	tokens.Access = accessToken

	refreshToken, refreshJTI, expired, err := genJWT.GenerateRefreshToken(id, ts.jwtSecret, ts.refreshExpDays)
	if err != nil {
		ts.logger.Error("failed to generate refresh token",
			zap.String("userID", id),
			zap.Error(err))
		return tokens, accessJTI, domain.ErrInternalServer
	}
	tokens.Refresh = refreshToken

	if err = ts.tokenRepo.SaveRefreshTokenState(ctx, refreshJTI, id, expired); err != nil {
		ts.logger.Error("failed to save refresh token state to repository",
			zap.String("refreshTokenID", refreshJTI),
			zap.String("userID", id),
			zap.Error(err))
		return tokens, accessJTI, domain.ErrInternalServer
	}

	return tokens, accessJTI, nil
}

func (ts *TokenServiceImpl) ValidateToken(ctx context.Context, token string, isAccess bool) (string, string, error) {
	genClaims, err := genJWT.ValidateTokens(token, ts.jwtSecret, isAccess)
	if err != nil {
		ts.logger.Error("token validation failed",
			zap.Bool("isAccess", isAccess),
			zap.Error(err))
		return "", "", err
	}

	var userID, jti string

	if accessor, ok := genClaims.(domain.ClaimDataAccessor); ok {
		userID = accessor.GetUserID()
		jti = accessor.GetJTI()
		ts.logger.Info("extracted token data", zap.String("userID", userID), zap.String("jti", jti))
	} else {
		ts.logger.Error("claims type does not implement ClaimDataAccessor", zap.String("type", fmt.Sprintf("%T", genClaims)))
		return "", "", domain.ErrInternalServer
	}

	if !isAccess {
		isValid, err := ts.tokenRepo.IsRefreshTokenValid(ctx, jti, userID)
		if err != nil {
			if errors.Is(err, domain.ErrRefreshTokenNotFound) {
				ts.logger.Warn("refresh token not found",
					zap.String("jti", jti),
					zap.String("userID", userID),
					zap.Error(err))
				return "", "", domain.ErrRefreshTokenNotValid
			} else if errors.Is(err, domain.ErrUserIDMismatch) {
				ts.logger.Warn("user ID mismatch for refresh token",
					zap.String("jti", jti),
					zap.String("userID", userID),
					zap.Error(err))
				return "", "", domain.ErrForbidden
			} else if errors.Is(err, domain.ErrTokenRevoked) {
				ts.logger.Warn("refresh token is explicitly revoked",
					zap.String("jti", jti),
					zap.String("userID", userID),
					zap.Error(err))
				return "", "", domain.ErrTokenRevoked
			} else if errors.Is(err, domain.ErrJTIInUserSessionsNotFound) {
				ts.logger.Warn("refresh token JTI not found in user sessions set",
					zap.String("jti", jti),
					zap.String("userID", userID),
					zap.Error(err))
				return "", "", domain.ErrRefreshTokenNotValid
			} else if errors.Is(err, domain.ErrRevokedStatusParseFail) {
				ts.logger.Error("failed to parse 'revoked' status from Redis for refresh token",
					zap.String("jti", jti),
					zap.String("userID", userID),
					zap.Error(err))
				return "", "", domain.ErrInternalServer
			} else {
				ts.logger.Error("unexpected error during refresh token validation",
					zap.String("jti", jti),
					zap.String("userID", userID),
					zap.Error(err))
				return "", "", domain.ErrInternalServer
			}
		}

		if !isValid {
			ts.logger.Warn("refresh token is considered invalid by repository logic without specific error",
				zap.String("jti", jti),
				zap.String("userID", userID))
			return "", "", domain.ErrRefreshTokenNotValid
		}

	}

	ts.logger.Info("refresh token successfully validated",
		zap.String("jti", jti),
		zap.String("userID", userID))

	return jti, userID, nil

}

func (ts *TokenServiceImpl) RevokeRefreshToken(ctx context.Context, token string) error {
	genClaims, err := genJWT.ParseRefreshToken(token, ts.jwtSecret)
	if err != nil {
		ts.logger.Error("failed to parse refresh token for revocation",
			zap.String("token", token),
			zap.Error(err))
		return domain.ErrTokenInvalid
	}

	err = ts.tokenRepo.RevokeRefreshToken(ctx, genClaims.ID, genClaims.UserID)
	if err != nil {
		ts.logger.Error("failed to revoke refresh token",
			zap.String("jti", genClaims.ID),
			zap.String("userID", genClaims.UserID),
			zap.Error(err))
		return domain.ErrInternalServer
	}

	return nil
}

func (ts *TokenServiceImpl) RevokeAllUserSessions(ctx context.Context, userID string) error {
	if err := ts.tokenRepo.RevokeAllUserRefreshTokens(ctx, userID); err != nil {
		ts.logger.Error("failed to revoke all user sessions",
			zap.String("userID", userID),
			zap.Error(err))
		return domain.ErrInternalServer
	}

	ts.logger.Info("all sessions revoked for user", zap.String("userID", userID))

	return nil
}

// BlacklistAccessToken parses an access token, extracts its JTI and expiry,
// and adds the JTI to the blacklist in the repository.
// If the token is already expired, it is a no-op.
func (ts *TokenServiceImpl) BlacklistAccessToken(ctx context.Context, tokenString string) error {
	claims, err := genJWT.ParseAccessToken(tokenString, ts.jwtSecret)
	if err != nil {
		if !errors.Is(err, domain.ErrTokenExpired) {
			ts.logger.Warn("attempted to blacklist an invalid access token", zap.Error(err))
			return domain.ErrTokenInvalid
		}
		// If the token is already expired, there's no need to blacklist it.
		ts.logger.Info("attempted to blacklist an already expired token, operation skipped", zap.Error(err))
		return nil
	}

	jti := claims.ID
	if claims.ExpiresAt == nil {
		ts.logger.Warn("access token missing exp; blacklist skipped", zap.String("jti", jti))
		return domain.ErrTokenInvalid
	}
	expiresAt := claims.ExpiresAt.Time
	if time.Until(expiresAt) <= 0 {
		ts.logger.Info("access token already expired; blacklist skipped", zap.String("jti", jti))
		return domain.ErrTokenExpired
	}

	if err := ts.tokenRepo.BlacklistAccessToken(ctx, jti, expiresAt); err != nil {
		ts.logger.Error("failed to blacklist access token in repository",
			zap.String("jti", jti),
			zap.Error(err))
		return domain.ErrInternalServer
	}

	ts.logger.Info("access token successfully blacklisted", zap.String("jti", jti))
	return nil
}

// IsAccessTokenBlacklisted checks if an access token's JTI is in the blacklist.
func (ts *TokenServiceImpl) IsAccessTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	isBlacklisted, err := ts.tokenRepo.IsAccessTokenBlacklisted(ctx, jti)
	if err != nil {
		ts.logger.Error("failed to check access token blacklist status", zap.String("jti", jti), zap.Error(err))
		return false, domain.ErrInternalServer
	}
	return isBlacklisted, nil
}
