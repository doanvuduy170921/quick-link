package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
	"github.com/doanvuduy170921/quick-link/internal/user/repository"
	pkgJwt "github.com/doanvuduy170921/quick-link/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type UserUseCase interface {
	Register(ctx context.Context, email, password string) (*db.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type userUseCase struct {
	repo       repository.UserRepository
	jwtManager *pkgJwt.JWTManager
	jwtTTL     time.Duration
}

func NewUserUseCase(
	repo repository.UserRepository,
	jwtManager *pkgJwt.JWTManager,
	jwtTTL time.Duration,
) UserUseCase {
	return &userUseCase{
		repo:       repo,
		jwtManager: jwtManager,
		jwtTTL:     jwtTTL,
	}
}

func (u *userUseCase) Register(ctx context.Context, email, password string) (*db.User, error) {
	// 2. Hash Password bằng Bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 3. Save User vào Database
	user, err := u.repo.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (u *userUseCase) Login(ctx context.Context, email, password string) (string, error) {
	// 1. Tìm User theo Email
	user, err := u.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// 2. So sánh Password Hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// 3. Sinh Access Token ký bằng RSA Private Key
	token, err := u.jwtManager.GenerateToken(user.ID, u.jwtTTL)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}
