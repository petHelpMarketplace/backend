package ports

import "pethelp-backend/internal/core/domain"

type SpecialistValidator interface {
	ValidateRegistrationReq(domain.RegistrationRequest) []domain.FieldError
	ValidateChangePasswordReq(domain.ChangePassReq) []domain.FieldError
}

type UnauthAppointmentValidator interface {
	ValidateUnauthAppointmentRequest(domain.SaveUnauthAppointmentnRequest) []domain.FieldError
}