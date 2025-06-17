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

func (ts *TokenServiceImpl) ValidateToken(ctx context.Context, token string, isAccess bool) (string, string, error) {

	genClaims, err := genJWT.ValidateTokens(token, ts.jwtSecret, isAccess)
	if err != nil {
		validateErr := fmt.Errorf("%s failed validate token: %w", operationTokenServ, err)
		ts.logger.Error("validate failed", zap.Error(validateErr))
		return "", "", domain.ErrTokenInvalid
	}

	var userID, jti string

	// Assert to the common interface.
	if accessor, ok := genClaims.(domain.ClaimDataAccessor); ok {
		userID = accessor.GetUserID()
		jti = accessor.GetJTI()
		ts.logger.Info("Extracted token data:", zap.String("userID", userID), zap.String("jti", jti))
	} else {
		ts.logger.Error("Error: Claims type does not implement ClaimDataAccessor", zap.String("type", fmt.Sprintf("%T", genClaims)))

		return "", "", domain.ErrInternalServer
	}

	revoked, err := ts.tokenRepo.IsRefreshTokenRevoked(ctx, jti, userID)
	if err != nil {
		revokedErr := fmt.Errorf("%s refresh token revoked check failed: %w", operationTokenServ, err)
		ts.logger.Error("check revoked error", zap.Error(revokedErr))
		return "", "", err
	}

	if !revoked {
		validErr := fmt.Errorf("%s refresh token revoked: %w", operationTokenServ, err)
		ts.logger.Error("token revoked", zap.Error(validErr))
		return "", "", domain.ErrTokenRevoked
	}

	ts.logger.Info("token valid")

	return jti, userID, nil

}

func (ts *TokenServiceImpl) RevokeToken(ctx context.Context, token string) error {
	genClaims, err := genJWT.ParseRefreshToken(token, ts.jwtSecret)
	if err != nil {
		parseErr := fmt.Errorf("%s failed parse refresh token: %w", operationTokenServ, err)
		ts.logger.Error("parse failed", zap.Error(parseErr))
		return domain.ErrTokenInvalid
	}

	err = ts.tokenRepo.RevokeRefreshToken(ctx, genClaims.ID, genClaims.UserID)
	if err != nil {
		revokeErr := fmt.Errorf("%s failed revoke refresh token: %w", operationTokenServ, err)
		ts.logger.Error("revoke failed", zap.Error(revokeErr))
		return domain.ErrTokenRevoked
	}

	return nil
}
