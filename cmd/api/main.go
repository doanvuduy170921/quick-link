package main

import (
	"context"
	"fmt"
	"github.com/doanvuduy170921/quick-link/configs"
	"github.com/doanvuduy170921/quick-link/internal/api"
	"github.com/doanvuduy170921/quick-link/internal/infrastructure"
	"log"
)

func main() {
	
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}
	ctx := context.Background()
	redis := infrastructure.NewRedisClient(cfg.Redis)

	if err := redis.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ Connected to redis successfully")

	server := api.NewServer(cfg, redis)

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}

}
