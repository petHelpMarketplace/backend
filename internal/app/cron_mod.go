package app

import (
	"context"
	"time"

	"pethelp-backend/internal/core/ports"

	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// CronModule defines the module for background tasks.
var CronModule = fx.Module("cron",
	fx.Invoke(RegisterCronJobs),
)

// RegisterCronJobs initializes the cron scheduler and registers the daily cleanup job.
func RegisterCronJobs(
	lc fx.Lifecycle,
	logger *zap.Logger,
	specialistService ports.SpecialistService,
) {
	// Create a new cron scheduler
	scheduler := cron.New()

	// Register the job to run at 02:00 AM every day.
	// Cron expression: "0 2 * * *" (Minute 0, Hour 2, Every Day, Every Month, Every Day of Week)
	_, err := scheduler.AddFunc("0 2 * * *", func() {
		logger.Info("starting scheduled daily cleanup of expired accounts")

		ctx := context.Background()

		// Execute the cleanup logic
		if err := specialistService.DeleteExpiredAccounts(ctx); err != nil {
			logger.Error("scheduled daily cleanup failed", zap.Error(err))
		} else {
			logger.Info("scheduled daily cleanup completed successfully")
		}
	})

	if err != nil {
		logger.Fatal("failed to add cron job", zap.Error(err))
	}

	// Hook into the application lifecycle
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting cron scheduler")
			scheduler.Start() // Start the scheduler in its own goroutine
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping cron scheduler")
			jobCtx := scheduler.Stop() // Stop the scheduler gracefully

			select {
			case <-jobCtx.Done():
				return nil
			case <-time.After(5 * time.Second):
				logger.Warn("cron scheduler stop timed out")
				return nil
			}
		},
	})
}
