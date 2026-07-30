package configs

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port  string `envconfig:"PORT" default:"8080"`
	Redis RedisConfig
}

type RedisConfig struct {
	Addr     string `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	Password string `envconfig:"REDIS_PASSWORD" default:""`
	DB       int    `envconfig:"REDIS_DB" default:"0"`
}

func LoadConfig() (*Config, error) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return &Config{}, err
	}
	return &config, nil
}
