package infrastructure

import (
	"context"
	"github.com/doanvuduy170921/quick-link/configs"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Ping(ctx context.Context) error
	Close() error
}

type redisClient struct {
	client *redis.Client
}

func NewRedisClient(cfg configs.RedisConfig) RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &redisClient{
		client: rdb,
	}
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *redisClient) Close() error {
	return r.client.Close()
}
