package app

import (
	"pethelp-backend/internal/core/ports"
	"pethelp-backend/internal/core/services"
	"pethelp-backend/internal/handlers"
	"pethelp-backend/internal/repositories"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ModuleParams holds common dependencies for auth modules.
// It supplies the Gin router, Postgres pool, Logger, and Redis client.
type FileUploadModuleParams struct {
	fx.In
	Router         *gin.Engine
	Logger         *zap.Logger
	AuthMiddleware gin.HandlerFunc
}

var FileUploadModule = fx.Module("file_storage",
	fx.Provide(
		fx.Annotate(
			repositories.NewS3Repository,
			fx.As(new(ports.FileRepository)),
		),
		fx.Annotate(
			services.NewFileUploadService,
			fx.As(new(ports.FileUploadService)),
		),
		fx.Annotate(
			handlers.NewFileHandler,
			fx.As(new(ports.FileHandlers)),
		),
	),
)
