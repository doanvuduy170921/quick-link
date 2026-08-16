package repository

import (
	"context"
	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
)

type UrlRepository interface {
	CreateURL(ctx context.Context, arg db.CreateURLParams) (db.Url, error)
	GetOriginalUrlByShortCode(ctx context.Context, shortCode string) (db.GetOriginalUrlByShortCodeRow, error)
}

func NewUrlRepository(db db.Querier) UrlRepository {
	return &urlRepository{
		db: db,
	}
}

type urlRepository struct {
	db db.Querier
}

func (u *urlRepository) CreateURL(ctx context.Context, arg db.CreateURLParams) (db.Url, error) {
	return u.db.CreateURL(ctx, arg)
}

func (u *urlRepository) GetOriginalUrlByShortCode(ctx context.Context, shortCode string) (db.GetOriginalUrlByShortCodeRow, error) {
	return u.db.GetOriginalUrlByShortCode(ctx, shortCode)
}
