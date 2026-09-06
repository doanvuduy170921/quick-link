package repository

import (
	"context"
	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
)

type ClickRepository interface {
	BatchInsertClick(ctx context.Context, arg []db.BatchInsertClickParams) (int64, error)
}

type clickRepository struct {
	db db.Querier
}

func NewClickRepository(db db.Querier) ClickRepository {
	return &clickRepository{
		db: db,
	}
}

func (c *clickRepository) BatchInsertClick(ctx context.Context, arg []db.BatchInsertClickParams) (int64, error) {
	return c.db.BatchInsertClick(ctx, arg)
}
