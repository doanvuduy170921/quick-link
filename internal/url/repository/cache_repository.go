package repository

import (
	"context"
	"github.com/doanvuduy170921/quick-link/internal/infrastructure"
	"time"
)

type UrlRepository interface {
	GetURL(ctx context.Context, key string) (string, error)
	SaveURL(ctx context.Context, key string, value string, expiration time.Duration) error
}

func NewUrlRepository(redis infrastructure.RedisClient) UrlRepository {
	return &urlRepository{
		redis: redis,
	}
}

type urlRepository struct {
	redis infrastructure.RedisClient
}

func (u *urlRepository) GetURL(ctx context.Context, key string) (string, error) {
	return u.redis.Get(ctx, key)

}

func (u *urlRepository) SaveURL(ctx context.Context, key string, value string, expiration time.Duration) error {
	return u.redis.Set(ctx, key, value, expiration)
}
