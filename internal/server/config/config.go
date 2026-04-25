package config

import "github.com/caarlos0/env/v11"

type Config struct {
	DSN       string `env:"DATABASE_URL" envDefault:"postgres://gophkeeper:gophkeeper@localhost:5432/gophkeeper?sslmode=disable"`
	JWTSecret string `env:"JWT_SECRET"   envDefault:"dev-secret-change-in-production"`
	Addr      string `env:"SERVER_ADDR"  envDefault:":8080"`
}

func Load() (Config, error) {
	var cfg Config
	return cfg, env.Parse(&cfg)
}
