package ports

import "pethelp-backend/internal/core/domain"

type SpecialistValidator interface {
	Validate(*domain.RegistrationRequest) []domain.FieldError
}
