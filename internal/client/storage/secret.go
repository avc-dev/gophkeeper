package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

// SyncStatus — статус синхронизации локальной записи с сервером.
type SyncStatus string

const (
	SyncStatusPending  SyncStatus = "pending"  // не отправлено на сервер
	SyncStatusSynced   SyncStatus = "synced"   // совпадает с сервером
	SyncStatusConflict SyncStatus = "conflict" // конфликт, требует разрешения
)

// LocalSecret — доменный объект клиентского хранилища.
// Расширяет domain.Secret полями для offline-работы и синхронизации.
type LocalSecret struct {
	domain.Secret
	ServerID        *uuid.UUID // nil пока не отправлен на сервер
	LocalVersion    int64
	ServerVersion   int64
	ServerUpdatedAt *time.Time
	SyncStatus      SyncStatus
	Deleted         bool
}

// SecretStorage хранит локальные секреты в SQLite.
type SecretStorage struct {
	db *sql.DB
}

// NewSecretStorage создаёт SecretStorage поверх уже открытой БД.
func NewSecretStorage(db *sql.DB) *SecretStorage {
	return &SecretStorage{db: db}
}

// scanSecret читает строку из sql.Rows в LocalSecret.
func scanSecret(rows *sql.Rows) (*LocalSecret, error) {
	var s LocalSecret
	var serverID sql.NullString
	var serverUpdatedAt sql.NullString
	var updatedAt string

	err := rows.Scan(
		&s.ID,
		&serverID,
		&s.Type,
		&s.Name,
		&s.Payload,
		&s.Metadata,
		&s.LocalVersion,
		&s.ServerVersion,
		&updatedAt,
		&serverUpdatedAt,
		&s.SyncStatus,
		&s.Deleted,
	)
	if err != nil {
		return nil, fmt.Errorf("scan secret row: %w", err)
	}

	t, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}
	s.UpdatedAt = t

	if serverID.Valid {
		id, err := uuid.Parse(serverID.String)
		if err != nil {
			return nil, fmt.Errorf("parse server_id: %w", err)
		}
		s.ServerID = &id
	}

	if serverUpdatedAt.Valid {
		t, err := time.Parse(time.RFC3339Nano, serverUpdatedAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse server_updated_at: %w", err)
		}
		s.ServerUpdatedAt = &t
	}

	return &s, nil
}

// checkAffected возвращает ErrSecretNotFound если ни одна строка не затронута операцией.
func checkAffected(res sql.Result, op string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s rows affected: %w", op, err)
	}
	if n == 0 {
		return domain.ErrSecretNotFound
	}
	return nil
}
