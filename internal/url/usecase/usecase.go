package usecase

import (
	"context"
	"github.com/doanvuduy170921/quick-link/internal/infrastructure"
	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
	"github.com/doanvuduy170921/quick-link/internal/url/repository"
	"github.com/doanvuduy170921/quick-link/pkg/utils"
	"time"
)

const (
	RedisExp = time.Duration(24 * time.Hour)
)

type UseCase interface {
	GenerateKey(ctx context.Context, URL string) (string, error)
	Redirect(ctx context.Context, code string) (string, error)
}

type useCase struct {
	repo  repository.UrlRepository
	redis infrastructure.RedisClient
}

func NewUseCase(repo repository.UrlRepository, redis infrastructure.RedisClient) UseCase {
	return &useCase{
		repo:  repo,
		redis: redis,
	}
}

func (u *useCase) GenerateKey(ctx context.Context, URL string) (string, error) {
	// gen short code
	key, err := utils.GenerateBase62Key(7)
	if err != nil {
		return "", err
	}
	// save into a database
	url, err := u.repo.CreateURL(ctx, db.CreateURLParams{
		ShortCode:   key,
		OriginalUrl: URL,
	})
	if err != nil {
		return "", err
	}
	// save into redis
	if err := u.redis.Set(ctx, key, url.OriginalUrl, RedisExp); err != nil {
		return "", err
	}
	return key, nil

}

func (u *useCase) Redirect(ctx context.Context, code string) (string, error) {
	return "", nil
}
