package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

// Delete помечает секрет как удалённый (soft delete) и выставляет pending.
// Физически строка остаётся в БД до подтверждения синхронизации с сервером.
func (s *SecretStorage) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	const q = `
UPDATE secrets
SET    deleted      = 1,
       updated_at   = ?,
       sync_status  = ?
WHERE  id = ? AND deleted = 0`

	res, err := s.db.ExecContext(ctx, q,
		now.Format(timeLayout), string(SyncStatusPending), id.String())
	if err != nil {
		return fmt.Errorf("secret delete: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("secret delete rows affected: %w", err)
	}
	if affected == 0 {
		return domain.ErrSecretNotFound
	}
	return nil
}

// Purge физически удаляет синхронизированные удалённые записи.
// Вызывается после успешного подтверждения удаления на сервере.
func (s *SecretStorage) Purge(ctx context.Context, id uuid.UUID) error {
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM secrets WHERE id = ? AND deleted = 1 AND sync_status = ?`,
		id.String(), string(SyncStatusSynced),
	); err != nil {
		return fmt.Errorf("secret purge: %w", err)
	}
	return nil
}
