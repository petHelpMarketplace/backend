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
	Name                 string `json:"name" validate:"required,min=2,max=100,custom_name" example:"John"`
	Phone                string `json:"phone" validate:"required,e123,min=13" example:"+38 (XXX) XXX-XX-XX)"`
	Email                string `json:"email" validate:"required,email,max=255" example:"john.doe@example.com"`
	Password             string `json:"password" validate:"required,min=12" example:"Str0ngP@ssw0rd!"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password" example:"Str0ngP@ssw0rd!"`
}
