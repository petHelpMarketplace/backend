package domain

import "github.com/golang-jwt/jwt/v5"

// AccessTokenClaims represents the claims for an access token.
// It embeds jwt.RegisteredClaims for standard JWT claims.
type AccessTokenClaims struct {
	UserID   string   `json:"user_id"`
	Roles    []string `json:"roles,omitempty"` // Omit if empty/not present
	TenantID string   `json:"tenant_id,omitempty"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims represents the claims for a refresh token.
type RefreshTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type TokensPair struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

// ClaimDataAccessor defines methods to access common data from any claims type.
type ClaimDataAccessor interface {
	GetUserID() string
	GetJTI() string
}

func (c *AccessTokenClaims) GetUserID() string {
	return c.UserID
}

func (c *AccessTokenClaims) GetJTI() string {
	return c.ID
}

func (c *RefreshTokenClaims) GetUserID() string {
	return c.UserID
}

func (c *RefreshTokenClaims) GetJTI() string {
	return c.ID
}
