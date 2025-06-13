package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

const (
	healthRoutePath = "/health"
)

var HealthModule = fx.Module("health",
	fx.Invoke(registerRoutes))

func registerRoutes(route *gin.Engine) {
	healthGroup := route.Group(healthRoutePath)
	{
		healthGroup.GET("", Check())
	}
}

// Check returns http handler for server health check
func Check() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	}
}
