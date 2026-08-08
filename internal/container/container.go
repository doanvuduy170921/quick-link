package container

import (
	"context"
	"errors"
	"fmt"
	"github.com/doanvuduy170921/quick-link/configs"
	"github.com/doanvuduy170921/quick-link/internal/infrastructure"
	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"time"
)

type Container struct {
	Config  *configs.Config
	Redis   infrastructure.RedisClient
	PgxConn *pgx.Conn
	Query   *db.Queries
}

func New(cfg *configs.Config) (*Container, error) {
	c := &Container{Config: cfg}

	// Redis
	if err := c.initRedis(); err != nil {
		return nil, err
	}

	// Postgres
	if err := c.initPostgres(); err != nil {
		c.Redis.Close()
		return nil, err
	}

	return c, nil
}

func (c *Container) initRedis() error {
	redisClient := redis.NewClient(&redis.Options{
		Addr:       c.Config.Redis.Addr,
		Password:   c.Config.Redis.Password,
		DB:         c.Config.Redis.DB,
		MaxRetries: c.Config.Redis.MaxRetries,
		PoolSize:   c.Config.Redis.PoolSize})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	// ← Wrap *redis.Client thành interface
	c.Redis = infrastructure.NewRedisClient(
		configs.RedisConfig{
			Addr:     c.Config.Redis.Addr,
			Password: c.Config.Redis.Password,
			DB:       c.Config.Redis.DB,
		},
	)

	return nil
}

func (c *Container) initPostgres() error {
	config, err := pgx.ParseConfig(c.Config.Postgres.GetDSN())
	if err != nil {
		return fmt.Errorf("parse postgres config failed: %w", err)
	}

	conn, err := pgx.ConnectConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("postgres connect failed: %w", err)
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("postgres ping failed: %w", err)
	}

	c.PgxConn = conn
	c.Query = db.New(conn)
	return nil
}

func (c *Container) Close() error {
	var errs error
	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if c.PgxConn != nil {
		if err := c.PgxConn.Close(context.Background()); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return nil // Return nil if all error is nil or return muti-err if have error
}
