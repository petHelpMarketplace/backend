package ports

import (
	"context"
	"time"
)

type EmailSender interface {
	SendAppointmentConfirmationEmail(ctx context.Context, clientEmai string, date, startTime, endTime time.Time) error
}
