package app

import (
	"context"
	"errors"
	"net/http"
	"pethelp-backend/internal/core/ports"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewGinServer(lc fx.Lifecycle, logger *zap.Logger, server *Server, cookieMngr ports.CookieManager) *gin.Engine {
	router := gin.Default()

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	// Add a unique ID to each request for tracing and logging.
	router.Use(requestid.New())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	router.Use(gin.Recovery())

	router.Use(cors.New(cors.Config{
		// AllowAllOrigins: true,
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:5174", "https://petbackend-a2vg.onrender.com", "https://accounts.google.com/*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           900 * time.Second,
	}))

	// Cookie middleware initialize session cookie
	router.Use(cookieMngr.Middleware())

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				err := server.ListenAndServe(router)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Fatal("Failed to start server", zap.Error(err))
				}
				logger.Info("Server stopped serving connections (could be shutdown or error)", zap.Error(err))
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := server.Shutdown(ctx); err != nil {
				return err
			}
			return nil
		},
	})

	return router
}
