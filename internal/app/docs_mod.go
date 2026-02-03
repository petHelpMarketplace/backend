package app

import (
	"github.com/gin-gonic/gin"
	swaggoFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "pethelp-backend/docs"

	"go.uber.org/fx"
)

//go:generate swag init -g cmd/api/main.go --parseDependency --parseInternal


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
	docsGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggoFiles.Handler, ginSwagger.URL("/api/v1/swagger/doc.json")))
}
