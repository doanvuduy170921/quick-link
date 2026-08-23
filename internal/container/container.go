package container

import (
	"context"
	"errors"
	"fmt"
	"github.com/doanvuduy170921/quick-link/configs"
	"github.com/doanvuduy170921/quick-link/internal/infrastructure"
	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Container struct {
	Config  *configs.Config
	Redis   infrastructure.RedisClient
	PgxPool *pgxpool.Pool
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
	redisCli := infrastructure.NewRedisClient(c.Config.Redis)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisCli.Ping(ctx); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	c.Redis = redisCli
	return nil
}

func (c *Container) initPostgres() error {
	config, err := pgxpool.ParseConfig(c.Config.Postgres.GetDSN())
	if err != nil {
		return fmt.Errorf("parse postgres config failed: %w", err)
	}
	config.MaxConns = int32(c.Config.Postgres.MaxOpenConn)
	config.MinConns = int32(c.Config.Postgres.MaxIdleConn)
	config.ConnConfig.ConnectTimeout = time.Duration(c.Config.Postgres.ConnTimeout) * time.Second
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("postgres connect failed: %w", err)
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("postgres ping failed: %w", err)
	}

	c.PgxPool = pool
	c.Query = db.New(pool)
	return nil
}

func (c *Container) Close() error {
	var errs error
	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if c.PgxPool != nil {
		c.PgxPool.Close()

	}
	return errs // Return nil if all error is nil or return muti-err if you have an error
}
