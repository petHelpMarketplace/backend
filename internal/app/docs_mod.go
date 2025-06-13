package app

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"

	_ "pethelp-backend/docs"
)

const (
	docsRoutePath = "api/v1"
)

var DocsModule = fx.Module("docs",
	fx.Invoke(registerDocsRoutes))

func registerDocsRoutes(route *gin.Engine) {

	docsGroup := route.Group(docsRoutePath)
	docsGroup.Use(gin.BasicAuth(gin.Accounts{
		"admin": "sysadmin",
	}))

	docsGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
