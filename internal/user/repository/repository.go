package repository

import (
	"context"

	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
)

// UserRepository định nghĩa contract cho tầng Persistence
type UserRepository interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByID(ctx context.Context, id int64) (db.User, error)
}

type userRepository struct {
	store db.Querier
}

func NewUserRepository(store db.Querier) UserRepository {
	return &userRepository{store: store}
}

func (r *userRepository) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return r.store.CreateUser(ctx, arg)
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return r.store.GetUserByEmail(ctx, email)
}

func (r *userRepository) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	return r.store.GetUserByID(ctx, id)
}
