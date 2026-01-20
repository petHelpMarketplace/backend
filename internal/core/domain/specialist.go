package domain

import (
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Specialist represents the 'specialists' table in the database.
type Specialist struct {
	ID             int64          `json:"id" db:"id"`
	Name           sql.NullString `json:"name" db:"name"`
	FamilyName     sql.NullString `json:"family_name" db:"family_name"`
	Phone          sql.NullString `json:"phone" db:"phone"`
	Email          string         `json:"email" db:"email"`
	PasswordHash   string         `json:"-" db:"password_hash"`
	Bio            sql.NullString `json:"bio" db:"bio"`
	Avatar         sql.NullString `json:"avatar" db:"avatar"`
	AddressID      sql.NullInt32  `json:"address_id" db:"address_id"`
	OrganisationID sql.NullInt32  `json:"organisation_id" db:"organisation_id"`
	BranchID       sql.NullInt32  `json:"branch_id" db:"branch_id"`
	Position       sql.NullString `json:"position" db:"position"`
	Experience     sql.NullInt32  `json:"experience" db:"experience"`
	Description    sql.NullString `json:"description" db:"description"`
	ImageID        []pgtype.Text  `json:"image_id" db:"image_id"`
	IsBanned       bool           `json:"is_banned" db:"is_banned"`
	IsDeleted      bool           `json:"is_deleted" db:"is_deleted"`
	IsActive       bool           `json:"is_active" db:"is_active"`
	IsVerified     bool           `json:"is_verified" db:"is_verified"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

// SpecialistProfDTO represents the public profile data of a specialist.
// @Description Specialist profile data returned to clients.
type SpecialistProfDTO struct {
	// Unique identifier of the specialist.
	// example: 123
	ID int64 `json:"id"`

	// Given name of the specialist.
	// example: Kateryna
	Name *string `json:"name"`

	// Family name (surname) of the specialist.
	// example: Walls
	FamilyName *string `json:"family_name"`

	// Phone number of the specialist in E.164 format.
	// example: +380961234567
	Phone *string `json:"phone"`

	// Email address of the specialist.
	// format: email
	// example: kateryna.walls@example.com
	Email string `json:"email"`

	// Short biography or summary of the specialist.
	// example: Experienced veterinarian specializing in small animal care.
	Bio *string `json:"bio"`

	// URL to the specialist's avatar image.
	// format: uri
	// example: https://your-cdn.com/avatars/kateryna_avatar.jpg
	AvatarURL *string `json:"avatar_url"`

	// URLs to the specialist's portfolio images.
	// format: array
	// example: ["https://your-cdn.com/portfolio/image1.jpg", "https://your-cdn.com/portfolio/video1.mp4"]
	PortfolioURLs []*string `json:"portfolio_urls"`

	// Years of professional experience.
	// minimum: 0
	// example: 7
	Experience *int32 `json:"experience"`

	// Professional position or title (e.g., "Veterinarian", "Dog Groomer").
	// example: Veterinarian
	Position *string `json:"position"`

	// Detailed description of services offered or qualifications.
	// example: Provides comprehensive veterinary services including diagnostics, surgery, and preventive medicine for cats and dogs.
	Description *string `json:"description"`

	// Indicates if the specialist's profile is currently active.
	// example: true
	IsActive bool `json:"is_active"`

	// Indicates if the specialist's credentials have been verified.
	// example: true
	IsVerified bool `json:"is_verified"`
}

// RegistrationRequest represents the request body for user registration.
// @Description User registration request payload
type RegistrationRequest struct {
	// Name of the user.
	// Allows Unicode letters, spaces, hyphens, and apostrophes.
	// required: true
	// minLength: 2
	// maxLength: 100
	// pattern: "^[\\p{L}\\s\\-'\\u2019]+$"
	// example: John Doe
	Name string `json:"name" validate:"required,min=2,max=100,custom_name" example:"John"`

	// Phone number of the user in a flexible E.123-like international format.
	// Must start with '+' followed by country code (1-3 digits).
	// Allows spaces, parentheses, and hyphens as separators.
	// required: true
	// minLength: 13 // Minimum length for +38 (093) 987-65-32
	// pattern: "^\\+\\d{1,3}(?:[()\\s-]*\\d+)*$"
	// example: "+38 (096) 123-45-67"
	Phone string `json:"phone" validate:"required,e123,min=13" example:"+38 (093) 987-65-32"`

	// Email address of the user.
	// required: true
	// format: email
	// maxLength: 255
	// example: john.doe@example.com
	Email string `json:"email" validate:"required,email,max=255" example:"john.doe@example.com"`

	// Password for the user account.
	// Must be at least 12 characters long.
	// Must contain at least one uppercase letter, one lowercase letter, one number, and one special character from @$!%*?&.
	// required: true
	// minLength: 12
	// maxLength: 255 // Arbitrary max length, adjust as needed
	// pattern: "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]{12,255}$"
	// example: Str0ngP@ssw0rd!
	Password string `json:"password" validate:"required,min=12" example:"Str0ngP@ssw0rd!"`

	// Password confirmation. Must match the Password field.
	// required: true
	// example: Str0ngP@ssw0rd!
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password" example:"Str0ngP@ssw0rd!"`
}

// SpecialistProfUpdateReq represents the request body for updating specialist profile information.
// @Description Specialist profile update request payload
type SpecialistProfUpdateReq struct {
	// Name of the specialist.
	// Allows Unicode letters, spaces, hyphens, and apostrophes.
	// minLength: 2
	// maxLength: 100
	// pattern: "^[\\p{L}\\s\\-'\\u2019]+$"
	// example: Kateryna
	Name *string `json:"name" validate:"omitempty,min=2,max=100,custom_name"`

	// Family name (surname) of the specialist.
	// minLength: 2
	// maxLength: 100
	// pattern: "^[\\p{L}\\s\\-'\\u2019]+$"
	// example: Walls
	FamilyName *string `json:"family_name" validate:"omitempty,min=2,max=100,custom_name"`

	// Phone number of the specialist in a flexible E.123-like international format.
	// Must start with '+' followed by country code (1-3 digits).
	// Allows spaces, parentheses, and hyphens as separators.
	// minLength: 13
	// pattern: "^\\+\\d{1,3}(?:[()\\s-]*\\d+)*$"
	// example: "+38 (096) 123-45-67"
	Phone *string `json:"phone" validate:"omitempty,e123,min=13"`

	// Years of professional experience.
	// minimum: 0
	// example: 7
	Experience *int32 `json:"experience_years" validate:"omitempty,min=0"`

	// Short biography or summary of the specialist.
	// maxLength: 1000
	// example: Experienced veterinarian specializing in small animal care.
	Bio *string `json:"bio" validate:"omitempty,max=1000"`
}


type SearchSpecialistParams struct {
	Animal     	    int64          `form:"animal_id" json:"animal_id,omitempty" db:"animal_id"`
	AnimalSize      int64          `form:"animal_size_id" json:"animal_size_id,omitempty" db:"animal_size_id"`
	Service         int64          `form:"service_id" json:"service_id,omitempty" db:"service_id"`
	City            int64          `form:"city_id" json:"city_id,omitempty" db:"city_id"`
	Area            int64          `form:"area_id" json:"area_id,omitempty" db:"area_id"`
}

type SearchSpecialistUriParams struct {
    AnimalCategory int64 `uri:"animalCategory" binding:"required"`
    AnimalSize     int64 `uri:"animalSize" binding:"required"`
    ServiceID      int64 `uri:"serviceID" binding:"required"`
    DistrictID     int64 `uri:"districtID" binding:"required"`
}

type SpecialistProfileSearchResponseDTO struct {
	// Unique identifier of the specialist.
	// example: 123
	ID int64 `json:"id,omitempty"`

	// Given name of the specialist.
	// example: Kateryna
	Name string `json:"name,omitempty"`

	// Family name (surname) of the specialist.
	// example: Walls
	FamilyName string `json:"family_name,omitempty"`

	// URL to the specialist's avatar image.
	// format: uri
	// example: https://your-cdn.com/avatars/kateryna_avatar.jpg
	AvatarURL string `json:"avatar_url,omitempty"`

	// Years of professional experience.
	// minimum: 0
	// example: 7
	Experience int32 `json:"experience,omitempty"`

	// Detailed description of services offered or qualifications.
	// example: Provides comprehensive veterinary services including diagnostics, surgery, and preventive medicine for cats and dogs.
	Description string `json:"description,omitempty"`

	// Indicates if the specialist's profile is currently active.
	// example: true
	IsActive bool `json:"is_active,omitempty"`

	// Indicates if the specialist's credentials have been verified.
	// example: true
	IsVerified bool `json:"is_verified,omitempty"`
}

type ServicePrice struct {
	Service      sql.NullString `json:"service" db:"service_name"`
    PricePerHour float64        `json:"price_per_hour" db:"price_per_hour"`
    PricePerDay  float64        `json:"price_per_day" db:"price_per_day"`

}

type ServicePriceDTO struct {
	Service      string  `json:"service,omitempty"`
    PricePerHour float64 `json:"price_per_hour,omitempty"`
    PricePerDay  float64 `json:"price_per_day,omitempty"`
}

type SpecialistDetails struct {
	Specialist
	ServicePrices []ServicePrice `json:"services"` 
}

type SpecialistDetailsDTO struct {
	SpecialistProfDTO
	ServicePrices []ServicePriceDTO `json:"services"` 
}


