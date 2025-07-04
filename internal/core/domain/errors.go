package domain

import "errors"

var (
	ErrInternalServer        = errors.New("internal server error")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrTokenExpired          = errors.New("token expired")
	ErrTokenRevoked          = errors.New("token status revoked")
	ErrTokenInvalid          = errors.New("token invalid")
	ErrTokenMalformed        = errors.New("token malformed")
	ErrTokenSignatureInvalid = errors.New("token signature invalid")
	ErrAuthFailed            = errors.New("authentication failed")
	ErrAccountNotFound       = errors.New("account not found")
	ErrAccountAlreadyExists  = errors.New("account already exists")
	ErrNoUppercase           = errors.New("password must contain at least one uppercase letter")
	ErrNoLowercase           = errors.New("password must contain at least one lowercase letter")
	ErrNoNumber              = errors.New("password must contain at least one number")
	ErrNoSpecialChar         = errors.New("password must contain at least one special character")
)

// FieldError contains validation error details for a specific field.
type FieldError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"This field is required."`
}

// ErrorResponse is the standard structure for API error responses.
type ErrorResponse struct {
	Code    int          `json:"code" example:"400"`
	Message string       `json:"message" example:"Validation failed"`
	Details []FieldError `json:"details,omitempty"`
}
