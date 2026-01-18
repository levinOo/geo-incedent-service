package main

import (
	"os"

	"github.com/levinOo/geo-incedent-service/internal/app"
)

// @title Geo Incedent Service API
// @version 1.0
// @description
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
