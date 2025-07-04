package services

import (
	"context"
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

func (ts *TokenServiceImpl) GenerateTokenPair(ctx context.Context, s *domain.SpecialistProfileDTO) (*domain.TokensPair, error) {

	tokens := &domain.TokensPair{}

	id := strconv.FormatInt(s.ID, 10)
	roles := []string{"specialist"}
	tenant := "petHelp"

	accessToken, _, err := genJWT.GenerateAccessToken(id, roles, tenant, ts.jwtSecret, ts.accessExpMin)
	if err != nil {
		ts.logger.Error("Failed to generate access token",
			zap.String("userID", id),
			zap.Error(err))
		return tokens, domain.ErrInternalServer
	}
	tokens.Access = accessToken

	refreshToken, refreshTokenID, expired, err := genJWT.GenerateRefreshToken(id, ts.jwtSecret, ts.refreshExpDays)
	if err != nil {
		ts.logger.Error("Failed to generate refresh token",
			zap.String("userID", id),
			zap.Error(err))
		return tokens, domain.ErrInternalServer
	}
	tokens.Refresh = refreshToken

	if err = ts.tokenRepo.SaveRefreshTokenState(ctx, refreshTokenID, id, expired); err != nil {
		ts.logger.Error("Failed to save refresh token state to repository",
			zap.String("refreshTokenID", refreshTokenID),
			zap.String("userID", id),
			zap.Error(err))
		return tokens, domain.ErrInternalServer
	}

	return tokens, nil
}

func (ts *TokenServiceImpl) ValidateToken(ctx context.Context, token string, isAccess bool) (string, string, error) {
	genClaims, err := genJWT.ValidateTokens(token, ts.jwtSecret, isAccess)
	if err != nil {
		ts.logger.Error("Token validation failed",
			zap.Bool("isAccess", isAccess),
			zap.Error(err))
		return "", "", err
	}

	var userID, jti string

	if accessor, ok := genClaims.(domain.ClaimDataAccessor); ok {
		userID = accessor.GetUserID()
		jti = accessor.GetJTI()
		ts.logger.Info("Extracted token data:", zap.String("userID", userID), zap.String("jti", jti))
	} else {
		ts.logger.Error("Error: Claims type does not implement ClaimDataAccessor", zap.String("type", fmt.Sprintf("%T", genClaims)))
		return "", "", domain.ErrInternalServer
	}

	if !isAccess {
		revoked, err := ts.tokenRepo.IsRefreshTokenRevoked(ctx, jti, userID)
		if err != nil {
			ts.logger.Error("Failed to check refresh token revocation status in repository",
				zap.String("jti", jti),
				zap.String("userID", userID),
				zap.Error(err))
			return "", "", domain.ErrInternalServer
		}

		if !revoked {
			ts.logger.Warn("Refresh token is revoked",
				zap.String("jti", jti),
				zap.String("userID", userID))
			return "", "", domain.ErrTokenRevoked
		}
	}

	ts.logger.Info("token valid", zap.String("userID", userID), zap.String("jti", jti))

	return jti, userID, nil

}

func (ts *TokenServiceImpl) RevokeToken(ctx context.Context, token string) error {
	genClaims, err := genJWT.ParseRefreshToken(token, ts.jwtSecret)
	if err != nil {
		ts.logger.Error("Failed to parse refresh token for revocation",
			zap.String("token", token),
			zap.Error(err))
		return domain.ErrTokenInvalid
	}

	err = ts.tokenRepo.RevokeRefreshToken(ctx, genClaims.ID, genClaims.UserID)
	if err != nil {
		ts.logger.Error("Failed to revoke refresh token in repository",
			zap.String("jti", genClaims.ID),
			zap.String("userID", genClaims.UserID),
			zap.Error(err))
		return domain.ErrInternalServer
	}

	return nil
}
