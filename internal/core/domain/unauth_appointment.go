package domain

import "time"

type Appointment struct {
	ID           int64     `json:"appointment_id" db:"appointment_id"`
	Category     int64     `json:"service_id" db:"service_id"`
	City         int64     `json:"city_id" db:"city_id"`
	District     int64     `json:"area_id" db:"area_id"`
	LocationType int64     `json:"location_type" db:"location_type"`
	AnimalSize   int64     `json:"animal_size_id" db:"location_type_id"`
	Address      int64     `json:"address_type_id" db:"address_type_id"`
	Description  string    `json:"description" db:"description"`
	Date         time.Time `json:"appointment_date" db:"appointment_date"`
	StartTime    time.Time `json:"start_time" db:"start_time"`
	EndTime      time.Time `json:"end_time" db:"end_time"`
	Amount       float32   `json:"amount" db:"amount"`
	UserEmail    string    `json:"user_id" db:"user_id"`
	SpecialistID int64     `json:"specialist_id" db:"specialist_id"`
	// Status          string         `json:"appointment_status" db:"appointment_status"`

}

// SaveUnauthAppointmentRequest represents the request body for saving anauthorized user' appointment request.
// @Description Anauthorized user' appointment request
type SaveUnauthAppointmentRequest struct {
	ServiceId    int       `json:"service_id" validate:"required,gt=0" example:"3"`
	CityId       int       `json:"city_id" validate:"required,gt=0" example:"1"`
	DistrictId   int       `json:"area_id" validate:"required,gt=0" example:"1"`
	Street       string    `json:"street" validate:"required" example:"вул. Володимирська"`
	LocationType string    `json:"location_type" validate:"required"`
	Unit         string    `json:"unit" validate:"required" example:"14"`
	Apt          string    `json:"apt" validate:"required" example:"14"`
	AnimalSizeId int       `json:"animal_size_id" validate:"required,gt=0" example:"1"`
	Description  string    `json:"description" validate:"required" example:"Пудель, потрібен грумінг перед виставкою."`
	Date         time.Time `json:"appointment_date" validate:"required" example:"2025-07-22"`
	StartTime    time.Time `json:"start_time" validate:"required" example:"08:00"`
	EndTime      time.Time `json:"end_time" validate:"required" example:"10:00"`
	Amount       float32   `json:"amount" validate:"required,gt=0" example:"500.00"`
	//unauth user email
	Email        string `json:"email" validate:"required,email,max=255" example:"john.doe@example.com"`
	SpecialistId int    `json:"specialist_id" validate:"required,gt=0" example:"3"`
	Status       string `json:"status" validate:"required" example:"pending"`
}
