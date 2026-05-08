package secret

import (
	"context"
	"fmt"
	"time"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *storage) List(ctx context.Context, userID uuid.UUID, since *time.Time) ([]*domain.Secret, error) {
	var (
		rows pgx.Rows
		err  error
	)

	// since = nil означает "вернуть все записи пользователя"
	if since == nil {
		rows, err = s.db.Query(ctx,
			`SELECT id, user_id, type, name, payload, metadata, version, created_at, updated_at
			 FROM secrets WHERE user_id = $1 ORDER BY updated_at ASC`,
			userID,
		)
	} else {
		rows, err = s.db.Query(ctx,
			`SELECT id, user_id, type, name, payload, metadata, version, created_at, updated_at
			 FROM secrets WHERE user_id = $1 AND updated_at > $2 ORDER BY updated_at ASC`,
			userID, *since,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	defer rows.Close()

	return collectSecrets(rows)
}

func collectSecrets(rows pgx.Rows) ([]*domain.Secret, error) {
	var result []*domain.Secret
	for rows.Next() {
		s := &domain.Secret{}
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.Type, &s.Name,
			&s.Payload, &s.Metadata, &s.Version,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan secret row: %w", err)
		}
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list secrets rows: %w", err)
	}
	return result, nil
}
