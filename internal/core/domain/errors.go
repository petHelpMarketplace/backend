package domain

import "errors"

var (
	ErrInternalServer              = errors.New("internal server error")
	ErrInvalidCredentials          = errors.New("invalid credentials")
	ErrTokenExpired                = errors.New("token expired")
	ErrTokenRevoked                = errors.New("token status revoked")
	ErrTokenInvalid                = errors.New("token invalid")
	ErrAuthFailed                  = errors.New("authentication failed")
	ErrInvalidEmail                = errors.New("invalid email")
	ErrInvalidPhone                = errors.New("invalid phone")
	ErrInvalidName                 = errors.New("invalid name")
	ErrInvalidFamilyName           = errors.New("invalid family name")
	ErrInvalidPasswordConfirmation = errors.New("invalid password confirmation")
	ErrEmailAlreadyInUse           = errors.New("email already in use")
	ErrPhoneAlreadyInUse           = errors.New("phone already in use")
	ErrAccountNotFound             = errors.New("account not found")
	ErrAccountAlreadyExists        = errors.New("account already exists")
	ErrFailedToHashPassword        = errors.New("failed to hash password")
	ErrInvalidInput                = errors.New("invalid input")
	ErrPasswordTooShort            = errors.New("password must be at least 12 characters long")
	ErrNoUppercase                 = errors.New("password must contain at least one uppercase letter")
	ErrNoLowercase                 = errors.New("password must contain at least one lowercase letter")
	ErrNoNumber                    = errors.New("password must contain at least one number")
	ErrNoSpecialChar               = errors.New("password must contain at least one special character")
)

type RequestResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
