package secret

import (
	"context"
	"errors"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *storage) Delete(ctx context.Context, userID, id uuid.UUID) error {
	row := s.db.QueryRow(ctx,
		`DELETE FROM secrets WHERE id = $1 AND user_id = $2 RETURNING id`,
		id, userID,
	)
	var deleted uuid.UUID
	if err := row.Scan(&deleted); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrSecretNotFound
		}
		return fmt.Errorf("delete secret: %w", err)
	}
	return nil
}
