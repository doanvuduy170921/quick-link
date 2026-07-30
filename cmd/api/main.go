package main

import (
	"context"
	"fmt"
	"github.com/doanvuduy170921/quick-link/configs"
	"github.com/doanvuduy170921/quick-link/internal/infrastructure"
	"github.com/gin-gonic/gin"
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

	r := gin.Default()

	r.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"Message": "Hello World",
		})
	})

	r.Run(":" + cfg.Port)

}
