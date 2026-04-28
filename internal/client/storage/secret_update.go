package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

// Update обновляет payload, metadata и bumps local_version; статус → pending.
func (s *SecretStorage) Update(ctx context.Context, id uuid.UUID, payload []byte, metadata string) (*LocalSecret, error) {
	now := time.Now().UTC()
	const q = `
UPDATE secrets
SET    payload       = ?,
       metadata      = ?,
       local_version = local_version + 1,
       updated_at    = ?,
       sync_status   = ?
WHERE  id = ? AND deleted = 0`

	res, err := s.db.ExecContext(ctx, q,
		payload, metadata, now.Format(timeLayout), string(SyncStatusPending), id.String())
	if err != nil {
		return nil, fmt.Errorf("secret update: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("secret update rows affected: %w", err)
	}
	if affected == 0 {
		return nil, domain.ErrSecretNotFound
	}
	return s.Get(ctx, id)
}

// MarkSynced обновляет sync_status и server_version после успешной синхронизации.
func (s *SecretStorage) MarkSynced(ctx context.Context, id uuid.UUID, serverID uuid.UUID, serverVersion int64, serverUpdatedAt time.Time) error {
	const q = `
UPDATE secrets
SET    server_id         = ?,
       server_version    = ?,
       server_updated_at = ?,
       sync_status       = ?
WHERE  id = ?`

	_, err := s.db.ExecContext(ctx, q,
		serverID.String(),
		serverVersion,
		serverUpdatedAt.Format(timeLayout),
		string(SyncStatusSynced),
		id.String(),
	)
	if err != nil {
		return fmt.Errorf("secret mark synced: %w", err)
	}
	return nil
}

// Upsert создаёт или обновляет запись, пришедшую с сервера (PULL фаза).
// Не трогает записи с sync_status = 'pending' — они имеют приоритет.
func (s *SecretStorage) Upsert(ctx context.Context, sec *LocalSecret) error {
	const q = `
INSERT INTO secrets(id, server_id, type, name, payload, metadata,
                   local_version, server_version, updated_at, server_updated_at,
                   sync_status, deleted)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    server_id         = excluded.server_id,
    type              = excluded.type,
    name              = excluded.name,
    payload           = excluded.payload,
    metadata          = excluded.metadata,
    server_version    = excluded.server_version,
    server_updated_at = excluded.server_updated_at,
    updated_at        = excluded.updated_at,
    local_version     = excluded.local_version,
    sync_status       = excluded.sync_status,
    deleted           = excluded.deleted
WHERE secrets.sync_status != 'pending'`

	var serverID *string
	if sec.ServerID != nil {
		s := sec.ServerID.String()
		serverID = &s
	}
	var serverUpdatedAt *string
	if sec.ServerUpdatedAt != nil {
		s := sec.ServerUpdatedAt.Format(timeLayout)
		serverUpdatedAt = &s
	}

	_, err := s.db.ExecContext(ctx, q,
		sec.ID.String(),
		serverID,
		string(sec.Type),
		sec.Name,
		sec.Payload,
		sec.Metadata,
		sec.LocalVersion,
		sec.ServerVersion,
		sec.UpdatedAt.Format(timeLayout),
		serverUpdatedAt,
		string(sec.SyncStatus),
		boolToInt(sec.Deleted),
	)
	if err != nil {
		return fmt.Errorf("secret upsert: %w", err)
	}
	return nil
}
