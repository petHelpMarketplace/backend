package jwtokens

import (
	"errors"
	"fmt"
	"pethelp-backend/internal/core/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

// GenerateAccessToken creates a new access token.
func GenerateAccessToken(userID string, roles []string, tenantID string, secretKey []byte, expiry time.Duration) (string, string, error) {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		locErr := fmt.Errorf("failed to time load location: %w", err)
		return "", "", locErr
	}
	timeNow := time.Now().In(loc)
	timeExp := timeNow.Add(expiry)

	jti := ulid.MustNew(ulid.Timestamp(timeNow), ulid.DefaultEntropy()).String()

	claims := domain.AccessTokenClaims{
		UserID:   userID,
		Roles:    roles,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(timeExp),
			IssuedAt:  jwt.NewNumericDate(timeNow),
			NotBefore: jwt.NewNumericDate(timeNow),
			Issuer:    "petHelp",
			Subject:   userID,
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}
	return tokenString, jti, nil
}

// GenerateRefreshToken creates a new refresh token.
// It also generates a unique JTI for server-side revocation.
func GenerateRefreshToken(userID string, secretKey []byte, expiry time.Duration) (string, string, time.Time, error) {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		locErr := fmt.Errorf("failed to time load location: %w", err)
		return "", "", time.Time{}, locErr
	}
	timeNow := time.Now().In(loc)
	timeExp := timeNow.Add(expiry)

	jti := ulid.MustNew(ulid.Timestamp(timeNow), ulid.DefaultEntropy()).String()

	claims := domain.RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(timeExp),
			IssuedAt:  jwt.NewNumericDate(timeNow),
			NotBefore: jwt.NewNumericDate(timeNow),
			Issuer:    "petHelp",
			Subject:   userID,
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return tokenString, jti, timeExp, nil
}

// ParseAccessToken parses and validates an access token.
func ParseAccessToken(tokenString string, secretKey []byte) (*domain.AccessTokenClaims, error) {
	claims := &domain.AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	switch {
	case token.Valid:
		fmt.Println("You look nice today")
	case errors.Is(err, jwt.ErrTokenMalformed):
		fmt.Println("That's not even a token")
		return nil, fmt.Errorf("access token parsing failed: %w", err)
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		// Invalid signature
		fmt.Println("Invalid signature")
		return nil, domain.ErrTokenInvalid
	case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
		// Token is either expired or not active yet
		fmt.Println("Timing is everything")
		return nil, domain.ErrTokenInvalid
	default:
		fmt.Println("Couldn't handle this token:", err)
		return nil, fmt.Errorf("access token parsing failed: %w", err)
	}

	// if err != nil {
	// 	return nil, fmt.Errorf("access token parsing failed: %w", err)
	// }

	// if !token.Valid {
	// 	return nil, domain.ErrTokenInvalid
	// }

	return claims, nil
}

// ParseRefreshToken parses and validates a refresh token.
func ParseRefreshToken(tokenString string, secretKey []byte) (*domain.RefreshTokenClaims, error) {
	claims := &domain.RefreshTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("refresh token parsing failed: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("refresh token is invalid")
	}

	if claims.ID == "" {
		return nil, errors.New("refresh token missing JTI claim")
	}
	// Important: Here you would typically check the JTI against a database
	// or cache (e.g., Redis) to see if this specific refresh token has been
	// revoked (e.g., due to logout or suspicious activity).
	// For example: if isRevoked(claims.ID) { return nil, errors.New("refresh token revoked") }

	return claims, nil
}
