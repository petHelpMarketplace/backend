package config

import (
	"errors"
	"os"

	"go.uber.org/zap"
)

const (
	dsnEnvName = "PG_DSN"
)

var (
	ErrMissingRequiredEnvVar = errors.New("missing required environment variable")
)

var _ PostgresConfig = (*postgresConfig)(nil)

type postgresConfig struct {
	dsn string
}

// NewPostgresConfig creates a new configuration for PostgresSQL using environment variables.
func NewPostgresConfig(logger *zap.Logger) (PostgresConfig, error) {
	dsn := os.Getenv(dsnEnvName)
	if dsn == "" {
		logger.Error("Postgres DSN environment variable is not set", zap.String("env_name", dsnEnvName))
		return nil, ErrMissingRequiredEnvVar
	}

	logger.Info("Postgres DSN loaded successfully")
	return &postgresConfig{
		dsn: dsn,
	}, nil
}

func (cfg *postgresConfig) DSN() string {
	return cfg.dsn
}
