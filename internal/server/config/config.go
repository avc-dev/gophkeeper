package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DSN               string `env:"DATABASE_URL"        envDefault:"postgres://gophkeeper:gophkeeper@localhost:5432/gophkeeper?sslmode=disable"`
	JWTPrivateKeyFile string `env:"JWT_PRIVATE_KEY_FILE"`
	JWTPublicKeyFile  string `env:"JWT_PUBLIC_KEY_FILE"`
	Addr              string `env:"SERVER_ADDR"         envDefault:":8080"`
	TLSCertFile       string `env:"TLS_CERT_FILE"`
	TLSKeyFile        string `env:"TLS_KEY_FILE"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	if cfg.JWTPrivateKeyFile == "" {
		return Config{}, fmt.Errorf("JWT_PRIVATE_KEY_FILE environment variable is required")
	}
	if cfg.JWTPublicKeyFile == "" {
		return Config{}, fmt.Errorf("JWT_PUBLIC_KEY_FILE environment variable is required")
	}
	if (cfg.TLSCertFile == "") != (cfg.TLSKeyFile == "") {
		return Config{}, fmt.Errorf("TLS_CERT_FILE and TLS_KEY_FILE must be set together")
	}
	return cfg, nil
}
