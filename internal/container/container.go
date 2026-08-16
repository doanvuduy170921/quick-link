package container

import (
	"context"
	"errors"
	"fmt"
	"github.com/doanvuduy170921/quick-link/configs"
	"github.com/doanvuduy170921/quick-link/internal/infrastructure"
	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
	"github.com/jackc/pgx/v5"
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
	return errs // Return nil if all error is nil or return muti-err if have error
}
