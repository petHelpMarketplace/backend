package main

import (
	"pethelp-backend/internal/app"
	"pethelp-backend/docs"
	_ "github.com/lib/pq"
)


// you can move this function to a better location in terms of structure
func initSwagger() {
    docs.SwaggerInfo.Title = "PetHelp API"
    docs.SwaggerInfo.Version = "1.0"
    docs.SwaggerInfo.BasePath = "/"
    docs.SwaggerInfo.Description = "API for PetHelp backend"
}

func main() {
    initSwagger()
	app.NewApp().Run()
}
