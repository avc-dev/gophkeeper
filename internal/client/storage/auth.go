package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// AuthStorage хранит токен, kdf_salt и метаданные синхронизации в таблице auth.
type AuthStorage struct {
	db *sql.DB
}

// NewAuthStorage создаёт AuthStorage поверх уже открытой БД.
func NewAuthStorage(db *sql.DB) *AuthStorage {
	return &AuthStorage{db: db}
}

// Set сохраняет или перезаписывает значение по ключу.
func (s *AuthStorage) Set(ctx context.Context, key, value string) error {
	const q = `INSERT INTO auth(key, value) VALUES(?, ?)
	           ON CONFLICT(key) DO UPDATE SET value = excluded.value`
	if _, err := s.db.ExecContext(ctx, q, key, value); err != nil {
		return fmt.Errorf("auth set %q: %w", key, err)
	}
	return nil
}

// Get возвращает значение по ключу. Если ключа нет — возвращает ("", nil).
func (s *AuthStorage) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM auth WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("auth get %q: %w", key, err)
	}
	return value, nil
}

// Delete удаляет ключ. Если ключа не было — не возвращает ошибку.
func (s *AuthStorage) Delete(ctx context.Context, key string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM auth WHERE key = ?`, key); err != nil {
		return fmt.Errorf("auth delete %q: %w", key, err)
	}
	return nil
}
