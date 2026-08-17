package main

import (
	"github.com/doanvuduy170921/quick-link/boot"
	"github.com/doanvuduy170921/quick-link/configs"
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

	boot.RunServer(cfg, cnt)

}
