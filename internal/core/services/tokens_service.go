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

const (
	operationTokenServ = "token_service: "
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

func (ts *TokenServiceImpl) GenerateTokenPair(ctx context.Context, s *domain.Specialist) (*domain.TokensPair, error) {

	tokens := &domain.TokensPair{}

	id := strconv.FormatInt(s.ID, 10)
	roles := []string{"specialist"}
	tenant := "petHelp"

	accessToken, _, err := genJWT.GenerateAccessToken(id, roles, tenant, ts.jwtSecret, ts.accessExpMin)
	if err != nil {
		accessErr := fmt.Errorf("%s failed to generate access token: %w", operationTokenServ, err)
		ts.logger.Error(domain.ErrFailedToHashPassword.Error(), zap.Error(accessErr))
		return tokens, accessErr
	}
	tokens.Access = accessToken

	refreshToken, refreshTokenID, expired, err := genJWT.GenerateRefreshToken(id, ts.jwtSecret, ts.refreshExpDays)
	if err != nil {
		refreshErr := fmt.Errorf("%s failed to generate access token: %w", operationTokenServ, err)
		ts.logger.Error(domain.ErrFailedToHashPassword.Error(), zap.Error(refreshErr))
		return tokens, refreshErr
	}
	tokens.Refresh = refreshToken

	// set freshly minted refresh token to valid list
	if err = ts.tokenRepo.SaveRefreshTokenState(ctx, refreshTokenID, id, expired); err != nil {
		saveRefreshErr := fmt.Errorf("%s failed to store refresh token: %w", operationTokenServ, err)
		ts.logger.Error(domain.ErrFailedToHashPassword.Error(), zap.Error(saveRefreshErr))
		return tokens, saveRefreshErr
	}

	return tokens, nil
}

func (ts *TokenServiceImpl) ParseAccessToken(context.Context, string) (*domain.AccessTokenClaims, error) {
	panic("unimplemented")
}

func (ts *TokenServiceImpl) ParseRefreshToken(context.Context, string) (*domain.RefreshTokenClaims, error) {
	panic("unimplemented")
}

func (ts *TokenServiceImpl) ValidateToken(context.Context, *domain.AccessTokenClaims) (*domain.Specialist, error) {
	panic("unimplemented")
}

func (ts *TokenServiceImpl) RevokeToken(context.Context, string) error {
	panic("unimplemented")
}
