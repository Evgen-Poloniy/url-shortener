package main

import (
	"flag"

	"github.com/Evgen-Poloniy/url-shortener/internal/app"

	// Register pgx driver for sqlx.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// @title           URL Shortener API
// @version         1.0
// @description     URL Shortener service backend API.
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	var storageType string
	flag.StringVar(&storageType, "storage-type", "", "Storage type (postgres, memory)")
	flag.Parse()

	app.Run(storageType)
}
