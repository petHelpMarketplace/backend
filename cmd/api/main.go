package main

import (
	"pethelp-backend/internal/app"

	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a PetHelp project backend
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host   petbackend-a2vg.onrender.com
// @BasePath  /api/v1/

// @schemes https http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	fx.New(app.NewApp()).Run()
}
