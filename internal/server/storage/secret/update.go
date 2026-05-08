package secret

import (
	"context"
	"errors"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *storage) Update(ctx context.Context, userID, id uuid.UUID, payload []byte, metadata string, expectedVersion int64) (*domain.Secret, error) {
	row := s.db.QueryRow(ctx,
		`UPDATE secrets
		 SET payload = $3, metadata = $4, version = version + 1, updated_at = now()
		 WHERE id = $1 AND user_id = $2 AND version = $5
		 RETURNING id, user_id, type, name, payload, metadata, version, created_at, updated_at`,
		id, userID, payload, metadata, expectedVersion,
	)
	sec, err := scanSecret(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, s.resolveUpdateConflict(ctx, userID, id)
		}
		return nil, fmt.Errorf("update secret: %w", err)
	}
	return sec, nil
}

// resolveUpdateConflict определяет причину провала UPDATE:
// запись не существует или версия не совпала.
func (s *storage) resolveUpdateConflict(ctx context.Context, userID, id uuid.UUID) error {
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM secrets WHERE id = $1 AND user_id = $2)`,
		id, userID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check secret existence: %w", err)
	}
	if !exists {
		return domain.ErrSecretNotFound
	}
	return domain.ErrVersionConflict
}
