package app

import (
	"log"
	"os"
	"time"

	"pethelp-backend/pkg/logger"

	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"pethelp-backend/internal/config"

	"pethelp-backend/pkg/database/postgres"
	redisDB "pethelp-backend/pkg/database/redis"
)

// NewApp returns an fx.Option that configures and wires application components for startup.
// 
// It loads a local ".env" file when APP_STAGE is empty or "local" (if the file exists),
// then initializes a logger from the resolved stage and LOG_LEVEL (fatal on logger init error).
// The returned option supplies the logger, registers configuration and server constructors,
// includes database modules (Postgres, Redis) and API modules (Health, Specialist, OAuth, Docs,
// Token, UnauthAppointment), and sets a 10-second start timeout for the Fx application.
// 
// Note: failures during .env loading or logger creation call log.Fatal and terminate the process.
func NewApp() fx.Option {

	envFileName := ".env"
	stage := os.Getenv("APP_STAGE")
	if stage == "" || stage == "local" {
		if _, err := os.Stat(envFileName); err == nil {
			if err := godotenv.Load(envFileName); err != nil {
				log.Fatalf("Failed to load .env file: %s - %s", envFileName, err.Error())

			}
		}

		log.Printf("Loaded .env file path: %s", envFileName)
	}

	logger, err := logger.New(logger.Stage(stage), os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.Fatal(err)
	}

	return fx.Options(
		fx.Supply(logger),
		// Core services
		fx.Provide(
			config.NewPostgresConfig,
			config.NewRedisConfig,
			config.NewServersConfig,
			config.LoadOAuthConf,
			config.LoadAuthConfig,
			NewHTTPServer,
			NewGinServer,
		),

		// Database modules
		postgres.Module,
		redisDB.Module,

		// API modules
		HealthModule,
		SpecialistModule,
		OauthModule,
		DocsModule,
		TokenModule,
		UnauthAppointmentModule,

		fx.StartTimeout(10*time.Second),
	)
}
