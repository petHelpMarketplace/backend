package redis

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

var (
	RedisClient *redis.Client
	Ctx         context.Context
)

func InitRedis() error {
	Ctx = context.Background()
	
	redisURI := os.Getenv("REDIS_URI")
	if redisURI == "" {
		return fmt.Errorf("REDIS_URI environment variable is not set")
	}

	opt, err := redis.ParseURL(redisURI)
	if err != nil {
		return fmt.Errorf("failed to parse Redis URI: %v", err)
	}

	RedisClient = redis.NewClient(opt)

	// Test the connection
	_, err = RedisClient.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	fmt.Println("Successfully connected to Redis!")
	return nil
}

// GetRedisClient returns the Redis client instance
func GetRedisClient() *redis.Client {
	return RedisClient
}

// GetContext returns the Redis context
func GetContext() context.Context {
	return Ctx
}