package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"pethelp-backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const operationName = "NewPGPool"

// DB is the PostgreSQL database connection pool.
type DB struct {
	pool *pgxpool.Pool
}

// Module provides the PostgreSQL database connection pool to the FX container.
var Module = fx.Options(
	fx.Provide(NewPGPool),
)

// NewPGPool creates a new PostgreSQL connection pool using pgxpool.
// It configures the pool based on environment variables and the provided server configuration.
func NewPGPool(lc fx.Lifecycle, dbConf *config.Servers, logger *zap.Logger) (*DB, error) {

	connString := os.Getenv("PG_DSN")
	if connString == "" {
		getEnvErr := fmt.Errorf("%s: PG_DSN environment variable not set", operationName)
		logger.Error("failed env", zap.Error(getEnvErr))
		return nil, getEnvErr
	}

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		getEnvErr := fmt.Errorf("%s: failed to parse connection string: %w", operationName, err)
		logger.Error("failed config", zap.Error(getEnvErr))
		return nil, getEnvErr
	}

	poolConfig.MaxConns = dbConf.Postgres.MaxPoolSize
	poolConfig.MinConns = dbConf.Postgres.MinPoolSize
	poolConfig.MaxConnLifetime = dbConf.Postgres.MaxLifetime
	poolConfig.MaxConnIdleTime = dbConf.Postgres.IdleTimeout
	poolConfig.ConnConfig.ConnectTimeout = dbConf.Postgres.ConnectionTimeout

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //For connect
	defer cancel()

	dbpool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		poolErr := fmt.Errorf("%s: failed to connect to database: %w", operationName, err)
		logger.Error("failed connect", zap.Error(poolErr))
		return nil, poolErr
	}

	// Add a lifecycle hook to close the connection pool.
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing database connection pool")
			dbpool.Close()
			return nil
		},
	})

	// Test the connection pool
	if err := dbpool.Ping(ctx); err != nil {
		pingErr := fmt.Errorf("%s: failed to connect to database: %w", operationName, err)
		logger.Error("failed test ping", zap.Error(pingErr))
		return nil, pingErr

	}

	logger.Info("Connected to PostgreSQL pool")
	return &DB{dbpool}, nil

}

// Pool returns the underlying pgxpool.Pool instance.
func (s *DB) Pool() *pgxpool.Pool {
	return s.pool
}
