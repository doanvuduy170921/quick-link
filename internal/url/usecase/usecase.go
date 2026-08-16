package usecase

import (
	"context"
	"errors"
	"log"

	"github.com/doanvuduy170921/quick-link/internal/infrastructure"
	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
	urlErr "github.com/doanvuduy170921/quick-link/internal/url/error"
	"github.com/doanvuduy170921/quick-link/internal/url/repository"
	"github.com/doanvuduy170921/quick-link/pkg/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

const (
	RedisExp            = 24 * time.Hour
	maxShortCodeRetries = 3
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
	if URL == "" {
		return "", urlErr.NewValidationError("url must not be empty")
	}

	expiresAt := time.Now().Add(RedisExp)

	var (
		key string
		url db.Url
	)
	for range maxShortCodeRetries {
		k, err := utils.GenerateBase62Key(7)
		if err != nil {
			return "", urlErr.NewInternalError("generate key error", err)
		}

		url, err = u.repo.CreateURL(ctx, db.CreateURLParams{
			ShortCode:   k,
			OriginalUrl: URL,
			ExpiresAt:   &expiresAt,
		})
		if err == nil {
			key = k
			break
		}
		if isUniqueViolation(err) {
			continue
		}
		return "", urlErr.NewInternalError("create url error", err)
	}
	if key == "" {
		return "", urlErr.NewInternalError("create url error", errors.New("failed to generate unique short code"))
	}

	ttl := time.Until(expiresAt)
	if url.ExpiresAt != nil {
		ttl = time.Until(*url.ExpiresAt)
	}
	if err := u.redis.Set(ctx, key, url.OriginalUrl, ttl); err != nil {
		log.Printf("redis set warning: short_code=%s: %v", key, err)
	}
	return key, nil
}

func (u *useCase) Redirect(ctx context.Context, code string) (string, error) {
	if code == "" {
		return "", urlErr.NewValidationError("code must not be empty")
	}
	// in redis have code duplicate with code param input --> return value(url)
	url, err := u.redis.Get(ctx, code)
	if err == nil && url != "" {
		return url, nil
	}
	// redis haven't code , query in db
	urlRe, err := u.repo.GetOriginalUrlByShortCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", urlErr.NewNotFoundError("short code not found", err)
		}
		return "", urlErr.NewInternalError("get original url error", err)
	}

	ttl := RedisExp
	if urlRe.ExpiresAt != nil {
		ttl = time.Until(*urlRe.ExpiresAt)
		if ttl <= 0 {
			return "", urlErr.NewNotFoundError("link expired", nil)
		}
	}

	if err := u.redis.Set(ctx, code, urlRe.OriginalUrl, ttl); err != nil {
		log.Printf("redis set warning: short_code=%s: %v", code, err)
	}
	return urlRe.OriginalUrl, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
