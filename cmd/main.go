package main

import (
	"log"
	"pethelp_backend/internal/infrastructure/db"
	"pethelp_backend/internal/infrastructure/redis"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
		log.Println("Continuing with existing environment variables...")
	}

	// Initialize database connection
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.GetDB().Close()

	// Initialize Redis connection
	if err := redis.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer redis.GetRedisClient().Close()

	// Initialize Gin server
	server := gin.Default()

	// Add middleware for database and redis access
	server.Use(func(c *gin.Context) {
		c.Set("db", db.GetDB())
		c.Set("redis", redis.GetRedisClient())
		c.Set("redisCtx", redis.GetContext())
		c.Next()
	})

	// Health check endpoint
	server.GET("/api/v1", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "Server is running",
		})
	})

	// Test connection to Redis
	server.GET("/api/v1/cache", func(c *gin.Context) {
		val, err := redis.RedisClient.Get(redis.Ctx, "ping").Result()
		if err != nil {
			log.Println("Redis error:", err)
		}
		c.JSON(200, gin.H{"status": "Redis is working", "message": val})
	})

	// Start the server
	if err := server.Run(":8000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
