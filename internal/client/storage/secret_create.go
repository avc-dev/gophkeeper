package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Create сохраняет новый секрет локально со статусом pending.
// ID генерируется на клиенте; серверный ID пока отсутствует.
func (s *SecretStorage) Create(ctx context.Context, sec *LocalSecret) (*LocalSecret, error) {
	if sec.ID == uuid.Nil {
		sec.ID = uuid.New()
	}
	if sec.UpdatedAt.IsZero() {
		sec.UpdatedAt = time.Now().UTC()
	}
	sec.SyncStatus = SyncStatusPending
	sec.LocalVersion = 1

	const q = `
INSERT INTO secrets(id, server_id, type, name, payload, metadata,
                   local_version, server_version, updated_at, server_updated_at,
                   sync_status, deleted)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var serverID *string
	if sec.ServerID != nil {
		s := sec.ServerID.String()
		serverID = &s
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
		nil,
		string(sec.SyncStatus),
		boolToInt(sec.Deleted),
	)
	if err != nil {
		return nil, fmt.Errorf("secret create: %w", err)
	}
	return sec, nil
}

const timeLayout = time.RFC3339Nano

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
