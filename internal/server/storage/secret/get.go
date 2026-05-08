package secret

import (
	"context"
	"errors"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *storage) Get(ctx context.Context, userID, id uuid.UUID) (*domain.Secret, error) {
	row := s.db.QueryRow(ctx,
		`SELECT id, user_id, type, name, payload, metadata, version, created_at, updated_at
		 FROM secrets WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	sec, err := scanSecret(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSecretNotFound
		}
		return nil, fmt.Errorf("get secret: %w", err)
	}
	return sec, nil
}
