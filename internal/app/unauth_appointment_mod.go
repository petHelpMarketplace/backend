package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"pethelp-backend/internal/core/ports"
	"pethelp-backend/internal/core/services"
	"pethelp-backend/internal/handlers"
	"pethelp-backend/internal/repositories"
	"pethelp-backend/pkg/database/postgres"
	redisCache "pethelp-backend/pkg/database/redis"
)

const UnauthAppointmentRoutePath = "/api/v1/public-appointment-request"

// ModuleParams holds common dependencies for auth modules.
// It supplies the Gin router, Postgres pool, Logger, and Redis client.
type UnauthAppointmentModuleParams struct {      
	fx.In
	Router         *gin.Engine
	DB             *postgres.DB
	Cache          *redisCache.DB
	Logger         *zap.Logger
}

var UnauthAppointmentModule = fx.Module("unauth_appointment",
	fx.Provide(
		fx.Annotate(
			repositories.NewUnauthAppointmentRepository,
			fx.As(new(ports.UnauthAppointmentRepository)),
		),

		fx.Annotate(
			services.NewUnauthAppointmentService,
			fx.As(new(ports.UnauthAppointmentService)),
		),

		fx.Annotate(
			services.NewUnauthAppointmentValidator,
			fx.As(new(ports.UnauthAppointmentValidator)),
		),

		fx.Annotate(
			handlers.NewUnauthAppointmentHandler,
			fx.As(new(ports.UnauthAppointmentHandler)),
		),

		// middleware.NewAuthMiddleware,
	),
	fx.Invoke(
		func(mp UnauthAppointmentModuleParams, handler ports.UnauthAppointmentHandler) {
			specRouterGroup := mp.Router.Group(UnauthAppointmentRoutePath)

			specRouterGroup.POST("", handler.Book)


			mp.Logger.Info("Unauth appointment routes",
				zap.String("base_path", UnauthAppointmentRoutePath),
				zap.String("book_unauth_app_endpoint", "/public-appointment-request"),
				zap.String("method", "POST"))
		},
	),
)
