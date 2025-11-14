package worker

import (
	"context"
	"pethelp-backend/internal/core/ports"
	"time"

	"go.uber.org/zap"
)


func StartExpirationChecker(apptService ports.UnauthAppointmentService, logger *zap.Logger) {
	ticker := time.NewTicker(120 * time.Minute)

	//stop the ticker when the server shuts down
	go func() {
		defer ticker.Stop()
		logger.Info("Appointment Expiration Checker started. Running every 2 hours.")

		apptService.CheckAndNotifyExpiredAppointments(context.Background())

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 80*time.Minute)
			apptService.CheckAndNotifyExpiredAppointments(ctx)
			cancel()
		}
	}()
	
}