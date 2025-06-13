package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const operationName = "redis_client"

// DB is the Redis database connection.
type DB struct {
	client *redis.Client
}

// Module provides the Redis database connection to the FX container.
var Module = fx.Options(
	fx.Provide(NewClient),
)

// NewClient creates a new Redis client instance.
// It parses the REDIS_URI environment variable for connection details and pings the server to ensure connectivity.
func NewClient(lc fx.Lifecycle, logger *zap.Logger) (*DB, error) {

	connString := os.Getenv("REDIS_URI") // Or from a config struct
	if connString == "" {
		getEnvErr := fmt.Errorf("%s: REDIS_URI environment variable not set", operationName)
		logger.Error("failed env", zap.Error(getEnvErr))
		return nil, getEnvErr
	}

	opt, err := redis.ParseURL(connString)
	if err != nil {
		parseURLErr := fmt.Errorf("%s: failed parse URL: %w", operationName, err)
		logger.Error("failed URL", zap.Error(parseURLErr))
		return nil, parseURLErr
	}
	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //For connect
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		pingErr := fmt.Errorf("%s: failed ping Redis database: %w", operationName, err)
		logger.Error("failed test ping", zap.Error(pingErr))
		return nil, pingErr
	}

	logger.Info("Redis connection created successfully")
	redisDB := &DB{client: client}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing Redis connection")
			if err := redisDB.client.Close(); err != nil {
				logger.Error("Error closing Redis connection", zap.Error(err))
				return err
			}
			return nil
		},
	})

	return redisDB, nil
}

// Client returns the underlying redis.Client instance.
func (s *DB) Client() *redis.Client {
	return s.client
}
