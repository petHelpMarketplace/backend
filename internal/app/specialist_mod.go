package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"pethelp-backend/internal/app/middleware"
	"pethelp-backend/internal/config"
	"pethelp-backend/internal/core/ports"
	"pethelp-backend/internal/core/services"
	"pethelp-backend/internal/handlers"
	"pethelp-backend/internal/repositories"
	"pethelp-backend/pkg/database/postgres"
	redisCache "pethelp-backend/pkg/database/redis"
)

const SpecialistRoutePath = "/api/v1/specialist"

// ModuleParams holds common dependencies for auth modules.
// It supplies the Gin router, Postgres pool, Logger, and Redis client.
type SpecialistModuleParams struct {
	fx.In
	Router         *gin.Engine
	DB             *postgres.DB
	Cache          *redisCache.DB
	Logger         *zap.Logger
	AuthConfig     config.AuthConfig
	AuthMiddleware gin.HandlerFunc
	CookieManager  ports.CookieManager
	FileHandler    ports.FileHandlers
}

var SpecialistModule = fx.Module("specialist",
	fx.Provide(
		fx.Annotate(
			repositories.NewSpecialistRepository,
			fx.As(new(ports.SpecialistRepository)),
		),

		fx.Annotate(
			services.NewSpecialistService,
			fx.As(new(ports.SpecialistService)),
		),

		fx.Annotate(
			repositories.NewTokenRepository,
			fx.As(new(ports.TokenRepository)),
		),

		fx.Annotate(
			services.NewTokenService,
			fx.As(new(ports.TokenService)),
		),

		fx.Annotate(
			services.NewCustomValidator,
			fx.As(new(ports.SpecialistValidator)),
		),

		fx.Annotate(
			handlers.NewSpecialistHandler,
			fx.As(new(ports.SpecialistHandlers)),
		),

		fx.Annotate(services.NewCookieManager,
			fx.As(new(ports.CookieManager))),

		middleware.NewAuthMiddleware,
	),
	fx.Invoke(
		func(mp SpecialistModuleParams, handler ports.SpecialistHandlers) {
			specRouterGroup := mp.Router.Group(SpecialistRoutePath)

			specRouterGroup.POST("/register", handler.Registration)
			specRouterGroup.POST("/login", handler.Login)
			specRouterGroup.GET("/specialists/search/:animal_id/:animal_size_id/:service_id/:area_id", handler.SearchSpecialistByServicePetArea)
			specRouterGroup.GET("/specialists/:id", handler.GetSpecialistDetailsById)

			protected := specRouterGroup.Use(mp.AuthMiddleware)
			protected.GET("/me", handler.Me)
			protected.PATCH("/change-password", handler.ChangePassword)
			protected.POST("/logout", handler.Logout)
			protected.PATCH("/profile", handler.UpdateProfile)
			protected.POST("/avatar", mp.FileHandler.UploadAvatar)
			protected.PATCH("/me/status", handler.DeactivateProfile)
			protected.POST("/portfolio", mp.FileHandler.UploadPortfolio)
			protected.DELETE("/portfolio/image", mp.FileHandler.DeletePortfolioImage)

			mp.Logger.Info("Registered specialist routes",
				zap.String("base_path", SpecialistRoutePath),
				zap.String("register_endpoint", "/specialist"),
				zap.String("method", "POST"))
		},
	),
)
