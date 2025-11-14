package ports

import (
	"context"
	"pethelp-backend/internal/core/domain"
	"time"
)

type EmailService interface {
	SendAppointmentConfirmationEmail(ctx context.Context, specialistID int64, clientEmail string, date, startTime, endTime time.Time) error
	GetSpecialistConfirmationEmail(ctx context.Context, specialistID int64) (string, error)
	SendAppointmentExpirationNotification(ctx context.Context, clientEmail string, date, startTime, endTime time.Time) error 
}

type EmailRepository interface{
	MarkEmailSent(ctx context.Context, ID int64) (domain.Appointment, error)
}

