package services

import (
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"

	"github.com/go-playground/validator/v10"
)

type UnauthAppointmentValidatorImpl struct {
	validator *validator.Validate
}

var _ ports.UnauthAppointmentValidator = (*UnauthAppointmentValidatorImpl)(nil)

func NewUnauthAppointmentValidator() *UnauthAppointmentValidatorImpl {
	v := validator.New()
	v.RegisterValidation("e123", isValidE123)
	v.RegisterValidation("custom_name", isValidName)

	return &UnauthAppointmentValidatorImpl{validator: v}
}

func (sv *UnauthAppointmentValidatorImpl) ValidateUnauthAppointmentRequest(data domain.SaveUnauthAppointmentRequest) []domain.FieldError {
	var validationErrors []domain.FieldError

	if err := sv.validator.Struct(data); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			validationErrors = append(validationErrors, domain.FieldError{Field: "general", Message: "Invalid validation object"})
			return validationErrors
		}

		for _, err := range err.(validator.ValidationErrors) {
			var fe domain.FieldError
			fe.Field = err.Field()
			switch err.Field() {
			case "ServiceId":
				fe.Message = "Invalid service id."
			case "CityId":
				fe.Message = "Invalid city id."
			case "DistrictId":
				fe.Message = "Invalid district id."
			case "Street":
				fe.Message = "Invalid street"
			case "LocationType":
				fe.Message = "Invalid location type"
			case "Unit":
				fe.Message = "Invalid unit"
			case "Apt":
				fe.Message = "Invalid appartment"
			case "AnimalSizeId":
				fe.Message = "Invalid animal size"
			case "Description":
				fe.Message = "Invalid description"
			case "Date":
				fe.Message = "Invalid appointment date"
			case "StartTime":
				fe.Message = "Invalid appointment start time"
			case "EndTime":
				fe.Message = "Invalid appointment end time"
			case "Amount":
				fe.Message = "Invalid ammount"
			case "Email":
				fe.Message = "Invalid email"
			case "SpecialistId":
				fe.Message = "Invalid specialist id"


			}
			
			validationErrors = append(validationErrors, fe)
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}