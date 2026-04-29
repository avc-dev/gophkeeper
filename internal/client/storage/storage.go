// Package storage реализует локальное SQLite-хранилище клиента.
// Использует modernc.org/sqlite (pure Go, без CGO).
// БД открывается в WAL-режиме; допустим только один открытый коннект.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // регистрирует драйвер "sqlite"
)

const driverName = "sqlite"

// Open открывает (или создаёт) локальную БД и применяет схему.
// path == ":memory:" допустим для тестов.
func Open(path string) (*sql.DB, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
		if err := restrictPermissions(path); err != nil {
			return nil, fmt.Errorf("set permissions: %w", err)
		}
	}

	db, err := sql.Open(driverName, path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// один коннект — обязателен при WAL для предотвращения SQLITE_BUSY.
	db.SetMaxOpenConns(1)

	if err := applyPragmas(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("init pragmas: %w", err)
	}

	if err := applySchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return db, nil
}

// applyPragmas включает WAL и внешние ключи.
func applyPragmas(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("pragma %q: %w", p, err)
		}
	}
	return nil
}

// Checkpoint сбрасывает WAL на диск. Вызывать перед закрытием процесса.
func Checkpoint(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		return fmt.Errorf("wal checkpoint: %w", err)
	}
	return nil
}
