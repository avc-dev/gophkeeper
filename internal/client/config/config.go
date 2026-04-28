// Package config загружает настройки клиента из переменных окружения.
package config

import (
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
)

// Config — конфигурация CLI-клиента.
type Config struct {
	ServerAddr    string `env:"GOPHKEEPER_SERVER"          envDefault:"localhost:8080"`
	DBPath        string `env:"GOPHKEEPER_DB"`
	TLSEnabled    bool   `env:"GOPHKEEPER_TLS"             envDefault:"false"`
	TLSCACert     string `env:"GOPHKEEPER_TLS_CA"`
	TLSSkipVerify bool   `env:"GOPHKEEPER_TLS_SKIP_VERIFY" envDefault:"false"`
}

// Load читает конфигурацию из окружения.
// DBPath по умолчанию: os.UserConfigDir()/gophkeeper/local.db
func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	if cfg.DBPath == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			dir = "."
		}
		cfg.DBPath = filepath.Join(dir, "gophkeeper", "local.db")
	}
	return cfg, nil
}
