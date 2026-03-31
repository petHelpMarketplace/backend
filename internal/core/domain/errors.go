package domain

import (
	"errors"
)

var (
	ErrInternalServer            = errors.New("internal server error")
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrTokenExpired              = errors.New("token expired")
	ErrTokenInvalid              = errors.New("token invalid")
	ErrTokenMalformed            = errors.New("token malformed")
	ErrTokenSignatureInvalid     = errors.New("token signature invalid")
	ErrAuthFailed                = errors.New("authentication failed")
	ErrAccountNotFound           = errors.New("account not found")
	ErrDistrictNotFound          = errors.New("district not found")
	ErrAccountAlreadyExists      = errors.New("account already exists")
	ErrNoUppercase               = errors.New("password must contain at least one uppercase letter")
	ErrNoLowercase               = errors.New("password must contain at least one lowercase letter")
	ErrNoNumber                  = errors.New("password must contain at least one number")
	ErrNoSpecialChar             = errors.New("password must contain at least one special character")
	ErrRefreshTokenNotFound      = errors.New("refresh token not found or expired")
	ErrUserIDMismatch            = errors.New("user ID mismatch")
	ErrRevokedStatusParseFail    = errors.New("failed to parse revoked status")
	ErrTokenRevoked              = errors.New("refresh token is revoked")
	ErrSessionMembershipFail     = errors.New("failed to check session membership")
	ErrJTIInUserSessionsNotFound = errors.New("JTI not found in user sessions")
	ErrRefreshTokenNotValid      = errors.New("refresh token not valid")
	ErrUnauthorized              = errors.New("unauthorized")
	ErrForbidden                 = errors.New("forbidden")
	ErrSessionTerminated         = errors.New("session terminated")
	ErrPasswordReuse             = errors.New("password reuse")
	ErrTimeUnavailable           = errors.New("time is unavailable")
	ErrInvalidTimeWindow         = errors.New("invalid time window")
	ErrNotFound                  = errors.New("results not found")
	ErrSpecialistsNotFound       = errors.New("specialists not found")	
	ErrInvalidParameter          = errors.New("invalid parameter")	
)

// FieldError contains validation error details for a specific field.
// type FieldError struct {
// 	Field   string `json:"field" example:"email"`
// 	Message string `json:"message" example:"This field is required."`
// }

// ErrorResponse is the standard structure for API error responses.
// type ErrorResponse struct {
// 	Code    int          `json:"code"`
// 	Message string       `json:"message" example:"Validation failed"`
// 	Details []FieldError `json:"details,omitempty"`
// }

// FieldError provides details about a single validation error.
type FieldError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"must be a valid email address"`
}

// BadRequestError is the response for 400 Bad Request errors, including validation.
type BadRequestError struct {
	Code    int          `json:"code" example:"400"`
	Message string       `json:"message" example:"Input validation failed"`
	Details []FieldError `json:"details,omitempty"`
}

// UnauthorizedError is the response for 401 Unauthorized errors.
type UnauthorizedError struct {
	Code    int          `json:"code" example:"401"`
	Message string       `json:"message" example:"Bearer token is missing or invalid"`
	Details []FieldError `json:"details,omitempty"`
}

// ForbiddenError is the response for 403 Forbidden errors.
// This is used when an authenticated user does not have the necessary permissions to perform an action.
type ForbiddenError struct {
	Code    int          `json:"code" example:"403"`
	Message string       `json:"message" example:"You do not have permission to access this resource"`
	Details []FieldError `json:"details,omitempty"`
}

// NotFoundError is the response for 404 Not Found errors.
type NotFoundError struct {
	Code    int          `json:"code" example:"404"`
	Message string       `json:"message" example:"Specialist account not found"`
	Details []FieldError `json:"details,omitempty"`
}

// ConflictError is the response for 409 Conflict errors.
// This typically occurs when trying to create a resource that already exists (e.g., duplicate email).
type ConflictError struct {
	Code    int          `json:"code" example:"409"`
	Message string       `json:"message" example:"A resource with this identifier already exists"`
	Details []FieldError `json:"details,omitempty"`
}

// PayloadTooLargeError is the response for 413 Payload Too Large errors.
// This is used when the client sends a request body (like a file upload) that exceeds the server's configured size limit.
type PayloadTooLargeError struct {
	Code    int          `json:"code" example:"413"`
	Message string       `json:"message" example:"The request payload is too large"`
	Details []FieldError `json:"details,omitempty"`
}

// UnsupportedMediaTypeError is the response for 415 Unsupported Media Type errors.
// This occurs when the server rejects the request because the `Content-Type` header (e.g., 'application/xml') is not supported by the endpoint.
type UnsupportedMediaTypeError struct {
	Code    int          `json:"code" example:"415"`
	Message string       `json:"message" example:"The provided media type is not supported"`
	Details []FieldError `json:"details,omitempty"`
}

// InternalServerError is the response for 500 Internal Server errors.
type InternalServerError struct {
	Code    int          `json:"code" example:"500"`
	Message string       `json:"message" example:"An unexpected error occurred on the server"`
	Details []FieldError `json:"details,omitempty"`
}
