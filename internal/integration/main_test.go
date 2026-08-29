//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"github.com/doanvuduy170921/quick-link/configs"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/doanvuduy170921/quick-link/internal/infrastructure"
	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
)

var (
	testPgPool   *pgxpool.Pool
	testQueries  *db.Queries
	testRedisCli infrastructure.RedisClient
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// ===== 1. Dựng Postgres container =====
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("quicklink_test"),
		tcpostgres.WithUsername("test_user"),
		tcpostgres.WithPassword("test_pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}

	pgDSN, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get postgres DSN: %v", err)
	}

	if err := runMigrations(pgDSN); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	testPgPool, err = pgxpool.New(ctx, pgDSN)
	if err != nil {
		log.Fatalf("failed to create pgxpool: %v", err)
	}
	testQueries = db.New(testPgPool)

	// ===== Init Redis container =====
	redisContainer, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		log.Fatalf("failed to start redis container: %v", err)
	}

	redisConnStr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get redis connection string: %v", err)
	}

	// redisConnStr dạng: "redis://localhost:32774"
	opt, err := redis.ParseURL(redisConnStr)
	if err != nil {
		log.Fatalf("failed to parse redis URL: %v", err)
	}

	testRedisCli = infrastructure.NewRedisClient(configs.RedisConfig{
		Addr: opt.Addr,
	})

	code := m.Run()

	testPgPool.Close()
	_ = pgContainer.Terminate(ctx)
	_ = redisContainer.Terminate(ctx)

	os.Exit(code)
}

func runMigrations(dsn string) error {
	m, err := migrate.New(
		"file://../../migrations",
		dsn,
	)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate.Up: %w", err)
	}
	return nil
}
