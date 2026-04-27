package storage

import (
	"context"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

const selectAll = `
SELECT id, server_id, type, name, payload, metadata,
       local_version, server_version, updated_at, server_updated_at,
       sync_status, deleted
FROM secrets`

// Get возвращает секрет по локальному ID. Удалённые записи не возвращаются.
func (s *SecretStorage) Get(ctx context.Context, id uuid.UUID) (*LocalSecret, error) {
	rows, err := s.db.QueryContext(ctx,
		selectAll+` WHERE id = ? AND deleted = 0`, id.String())
	if err != nil {
		return nil, fmt.Errorf("secret get: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("secret get rows: %w", err)
		}
		return nil, domain.ErrSecretNotFound
	}

	sec, err := scanSecret(rows)
	if err != nil {
		return nil, fmt.Errorf("secret get scan: %w", err)
	}
	return sec, nil
}

// GetByName возвращает секрет по имени и типу. Удалённые не возвращаются.
func (s *SecretStorage) GetByName(ctx context.Context, name string, typ domain.SecretType) (*LocalSecret, error) {
	rows, err := s.db.QueryContext(ctx,
		selectAll+` WHERE name = ? AND type = ? AND deleted = 0`, name, string(typ))
	if err != nil {
		return nil, fmt.Errorf("secret get by name: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("secret get by name rows: %w", err)
		}
		return nil, domain.ErrSecretNotFound
	}

	sec, err := scanSecret(rows)
	if err != nil {
		return nil, fmt.Errorf("secret get by name scan: %w", err)
	}
	return sec, nil
}

// GetByServerID возвращает секрет по серверному UUID (используется при синхронизации).
func (s *SecretStorage) GetByServerID(ctx context.Context, serverID uuid.UUID) (*LocalSecret, error) {
	rows, err := s.db.QueryContext(ctx,
		selectAll+` WHERE server_id = ?`, serverID.String())
	if err != nil {
		return nil, fmt.Errorf("secret get by server id: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("secret get by server id rows: %w", err)
		}
		return nil, domain.ErrSecretNotFound
	}

	sec, err := scanSecret(rows)
	if err != nil {
		return nil, fmt.Errorf("secret get by server id scan: %w", err)
	}
	return sec, nil
}
