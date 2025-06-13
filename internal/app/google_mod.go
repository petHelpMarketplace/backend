package app

import (
	"pethelp-backend/internal/core/ports"
	"pethelp-backend/internal/core/services"
	"pethelp-backend/internal/handlers"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const oauthCoreRoutePath = "/api/v1/oauth"

type OauthModuleParams struct {
	fx.In

	Router *gin.Engine
	Logger *zap.Logger
}

var OauthModule = fx.Module("google_oauth",
	fx.Provide(
		// fx.Annotate(
		// 	repositories.NewOAuthTokenRepo,
		// 	fx.As(new(ports.TokenRepository)),
		// ),

		fx.Annotate(
			services.NewOAuthService,
			fx.As(new(ports.OAuthService)),
		),
		fx.Annotate(
			handlers.NewOAuthHandlers,
			fx.As(new(ports.OAuthHandlers)),
		),
	),
	fx.Invoke(
		func(mp OauthModuleParams, handler ports.OAuthHandlers) {
			oauthGroup := mp.Router.Group(oauthCoreRoutePath)
			{
				oauthGroup.GET("/google", handler.SignInWithProvider)
				oauthGroup.GET("/google/callback", handler.ProviderCallback)

				mp.Logger.Info("Registered google OAuth routes",
					zap.String("base_path", oauthCoreRoutePath),
					zap.String("register_endpoint", "/google"),
					zap.String("method", "GET"))
			}
		},
	),
)
