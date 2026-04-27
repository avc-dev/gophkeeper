package storage

import (
	"context"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/domain"
)

// List возвращает все не удалённые секреты опционально отфильтрованные по типу.
// typ == "" означает "все типы".
func (s *SecretStorage) List(ctx context.Context, typ domain.SecretType) ([]*LocalSecret, error) {
	query := selectAll + ` WHERE deleted = 0`
	args := []any{}

	if typ != "" {
		query += ` AND type = ?`
		args = append(args, string(typ))
	}

	query += ` ORDER BY name`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("secret list: %w", err)
	}
	defer rows.Close()

	var result []*LocalSecret
	for rows.Next() {
		sec, err := scanSecret(rows)
		if err != nil {
			return nil, fmt.Errorf("secret list scan: %w", err)
		}
		result = append(result, sec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("secret list rows: %w", err)
	}
	return result, nil
}

// ListPending возвращает записи, ожидающие отправки на сервер.
func (s *SecretStorage) ListPending(ctx context.Context) ([]*LocalSecret, error) {
	rows, err := s.db.QueryContext(ctx,
		selectAll+` WHERE sync_status = ?`, string(SyncStatusPending))
	if err != nil {
		return nil, fmt.Errorf("secret list pending: %w", err)
	}
	defer rows.Close()

	var result []*LocalSecret
	for rows.Next() {
		sec, err := scanSecret(rows)
		if err != nil {
			return nil, fmt.Errorf("secret list pending scan: %w", err)
		}
		result = append(result, sec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("secret list pending rows: %w", err)
	}
	return result, nil
}
