package middleware

import (
	"context"
	"fmt"
	"github.com/doanvuduy170921/quick-link/internal/infrastructure"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type RateLimiter struct {
	redis infrastructure.RedisClient
	ttl   time.Duration
	limit int64
}

func NewRateLimiter(redis infrastructure.RedisClient, ttl time.Duration, limit int64) *RateLimiter {
	return &RateLimiter{
		redis: redis,
		ttl:   ttl,
		limit: limit,
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("%s_%s", c.Request.Host, clientIP)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()

		count, err := rl.redis.Incr(ctx, key)
		if err != nil {
			c.Next()
			return
		}
		if count == 1 {
			_ = rl.redis.Expire(ctx, key, rl.ttl)

		}
		if count > rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
