package domain

import (
	"time"
)

type Specialist struct {
	ID           int64     `json:"id,omitempty"`
	Name         string    `json:"name,omitempty"`
	FamilyName   string    `json:"family_name,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	Email        string    `json:"email,omitempty"`
	PasswordHash string    `json:"-"` // store hashed password; omit from JSON responses
	IsBanned     bool      `json:"is_banned,omitempty"`
	IsDeleted    bool      `json:"is_deleted,omitempty"`
	IsActive     bool      `json:"is_active,omitempty"`
	IsVerified   bool      `json:"is_verified,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type RegistrationRequest struct {
	Name                 string `json:"name" validate:"required,min=2,max=12,custom_name"`
	FamilyName           string `json:"family_name" validate:"required,min=2,custom_name"`
	Phone                string `json:"phone" validate:"required,e123"`
	Email                string `json:"email" validate:"required,email"`
	Password             string `json:"password" validate:"required,min=12"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
}
