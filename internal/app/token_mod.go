package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"pethelp-backend/internal/core/ports"
	"pethelp-backend/internal/handlers"
)

const TokenRoutePath = "/api/v1/token"

// ModuleParams holds common dependencies for token modules.
// It supplies the Gin router, Logger.
type TokenModuleParams struct {
	fx.In
	Router *gin.Engine
	Logger *zap.Logger
}

var TokenModule = fx.Module("token",
	fx.Provide(
		fx.Annotate(
			handlers.NewTokenHandler,
			fx.As(new(ports.TokenHandlers)),
		),
	),
	fx.Invoke(
		func(mp TokenModuleParams, handler ports.TokenHandlers) {
			tokenRouterGroup := mp.Router.Group(TokenRoutePath)

			tokenRouterGroup.POST("/refresh", handler.RefreshToken)

			mp.Logger.Info("Registered token routes",
				zap.String("base_path", TokenRoutePath),
				zap.String("register_endpoint", "/token"),
				zap.String("method", "POST"))
		},
	),
)
