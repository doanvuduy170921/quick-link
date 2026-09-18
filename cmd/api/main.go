package main

import (
	"github.com/doanvuduy170921/quick-link/configs"
	_ "github.com/doanvuduy170921/quick-link/docs"
	"github.com/doanvuduy170921/quick-link/internal/boot"
	"github.com/doanvuduy170921/quick-link/internal/container"
	"github.com/joho/godotenv"
	"log"
)

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

// @title           Quick-Link API
// @version         1.0
// @description     URL Shortener service with JWT auth, click analytics (goroutine/channel worker pool), Redis caching and rate limiting.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Doan Vu Duy
// @contact.url    https://github.com/doanvuduy170921

// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// load configs
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}
	// init container
	cnt, err := container.New(cfg)
	if err != nil {
		log.Fatalf(" Failed to create container: %v", err)
	}

	defer cnt.Close()

	if err := boot.RunServer(cfg, cnt); err != nil {
		log.Fatalf(" Failed to start server: %v", err)
	}

}
