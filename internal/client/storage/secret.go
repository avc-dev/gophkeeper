package storage

import (
	"context"
	"database/sql"
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
		return nil, err
	}

	t, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return nil, err
	}
	s.UpdatedAt = t

	if serverID.Valid {
		id, err := uuid.Parse(serverID.String)
		if err != nil {
			return nil, err
		}
		s.ServerID = &id
	}

	if serverUpdatedAt.Valid {
		t, err := time.Parse(time.RFC3339Nano, serverUpdatedAt.String)
		if err != nil {
			return nil, err
		}
		s.ServerUpdatedAt = &t
	}

	return &s, nil
}

// defaultCtx используется только в helper-функциях без контекста — в продакшн коде всегда передаём ctx явно.
func withCtx(ctx context.Context) context.Context { return ctx }
