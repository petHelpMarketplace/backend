package ports

import (
	"context"
	"time"
)

type EmailService interface {
	SendAppointmentConfirmationEmail(ctx context.Context, specialistID int64, clientEmail string, date, startTime, endTime time.Time) error
	GetSpecialistConfirmationEmail(ctx context.Context, specialistID int64) (string, error)
}

