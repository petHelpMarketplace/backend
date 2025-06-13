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

		fx.StartTimeout(10*time.Second),
	)
}
