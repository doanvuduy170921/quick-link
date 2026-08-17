// configs/config.go
package configs

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	Server   ServerConfig
	Redis    RedisConfig
	Postgres PostgresConfig
	Logger   LoggerConfig
}

type ServerConfig struct {
	Port         string        `envconfig:"SERVER_PORT" default:"8080"`
	ReadTimeout  time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"10s"`
	WriteTimeout time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout  time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"30s"`
}

type RedisConfig struct {
	Addr        string `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	Password    string `envconfig:"REDIS_PASSWORD" default:""`
	DB          int    `envconfig:"REDIS_DB" default:"0"`
	MaxRetries  int    `envconfig:"REDIS_MAX_RETRIES" default:"3"`
	PoolSize    int    `envconfig:"REDIS_POOL_SIZE" default:"10"`
	ConnTimeout int    `envconfig:"REDIS_CONN_TIMEOUT" default:"5"`
}

type PostgresConfig struct {
	DBUser      string `envconfig:"POSTGRES_USER" default:"postgres"`
	DBName      string `envconfig:"POSTGRES_NAME" default:"postgres"`
	DBHost      string `envconfig:"POSTGRES_HOST" default:"localhost"`
	DBPort      string `envconfig:"POSTGRES_PORT" default:"5432"`
	DBSSLMode   string `envconfig:"POSTGRES_SSL_MODE" default:"disable"`
	DBPassword  string `envconfig:"POSTGRES_PASSWORD" default:""`
	MaxOpenConn int    `envconfig:"POSTGRES_MAX_OPEN_CONN" default:"25"`
	MaxIdleConn int    `envconfig:"POSTGRES_MAX_IDLE_CONN" default:"5"`
	ConnTimeout int    `envconfig:"POSTGRES_CONN_TIMEOUT" default:"10"`
}

type LoggerConfig struct {
	Level string `envconfig:"LOG_LEVEL" default:"info"`
}

func (p *PostgresConfig) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.DBUser, p.DBPassword, p.DBHost, p.DBPort, p.DBName, p.DBSSLMode,
	)
}

func LoadConfig() (*Config, error) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return nil, err
	}
	return &config, nil
}
