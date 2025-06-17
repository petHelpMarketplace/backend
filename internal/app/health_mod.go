package app

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

const (
	healthRoutePath = "/api/v1/health"
)

var HealthModule = fx.Module("health",
	fx.Invoke(registerRoutes))

func registerRoutes(route *gin.Engine) {
	healthGroup := route.Group(healthRoutePath)
	{
		healthGroup.GET("", Check)
	}
}

// HealthCheckResponse defines the structure of the health check response.
type HealthCheckResponse struct {
	Status    string `json:"status" example:"healthy"`
	Timestamp string `json:"timestamp" example:"2023-10-27T10:00:00Z"`
}

// healthCheck godoc
// @Summary      Performs a health check on the application
// @Description  Checks if the application is running and responsive.
// @Tags         System
// @Accept       json
// @Produce      json
// @Success      200  {object}  HealthCheckResponse "Application is healthy"
// @Failure      503  {object}  map[string]string   "Application is unhealthy or service unavailable"
// @Router       /health [get]
func Check(c *gin.Context) {
	response := HealthCheckResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	c.JSON(http.StatusOK, response)
}
