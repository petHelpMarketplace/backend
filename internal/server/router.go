package server

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
)

func NewGinServer(lc fx.Lifecycle, logger *zap.Logger, server *Server) *gin.Engine {
	router := gin.Default()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// lc.Append(fx.Hook{
	// 	OnStart: func(ctx context.Context) error {
	// 		go func() {
	// 			logger.Info("Listening on", zap.String("address", server.config.Address))
	// 			if err := server.ListenAndServe(router); err != nil {
	// 				logger.Fatal("Failed to start server", zap.Error(err))
	// 			}
	// 		}()
	// 		return nil
	// 	},
	// 	OnStop: func(ctx context.Context) error {
	// 		logger.Info("Stopping server...")
	// 		return server.httpServer.Shutdown(ctx)
	// 	},
	// })

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Listening on", zap.String("address", server.config.Address))
	
			env := os.Getenv("ENV")
			if env == "production" {
				return server.ListenAndServe(router)
			} else {
				go func() {
					if err := server.ListenAndServe(router); err != nil && err != http.ErrServerClosed {
						logger.Fatal("Failed to start server", zap.Error(err))
					}
				}()
				return nil
			}
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping server...")
			return server.httpServer.Shutdown(ctx)
		},
	})
	
	
	return router
}