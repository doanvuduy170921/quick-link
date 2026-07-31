package usecase

import (
	"context"
	"github.com/doanvuduy170921/quick-link/internal/url/repository"
	"github.com/doanvuduy170921/quick-link/pkg/utils"
	"time"
)

type UseCase interface {
	GenerateKey(ctx context.Context, value string, expiration int) (string, error)
}

type useCase struct {
	repo repository.UrlRepository
}

func NewUseCase(repo repository.UrlRepository) UseCase {
	return &useCase{
		repo: repo,
	}
}

func (u *useCase) GenerateKey(ctx context.Context, value string, expiration int) (string, error) {
	key, err := utils.GenerateBase62Key(7)
	if err != nil {
		return "", err
	}
	exp := time.Duration(expiration) * time.Second
	if err := u.repo.SaveURL(ctx, key, value, exp); err != nil {
		return "", err
	}
	return key, err
}
