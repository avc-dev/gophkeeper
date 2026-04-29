package secret

import (
	"context"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *storage) Create(ctx context.Context, sec *domain.Secret) (*domain.Secret, error) {
	sec.ID = uuid.New()
	row := s.db.QueryRow(ctx,
		`INSERT INTO secrets (id, user_id, type, name, payload, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, user_id, type, name, payload, metadata, version, created_at, updated_at`,
		sec.ID, sec.UserID, sec.Type, sec.Name, sec.Payload, sec.Metadata,
	)
	return scanSecret(row)
}

func scanSecret(row pgx.Row) (*domain.Secret, error) {
	s := &domain.Secret{}
	err := row.Scan(
		&s.ID, &s.UserID, &s.Type, &s.Name,
		&s.Payload, &s.Metadata, &s.Version,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan secret: %w", err)
	}
	return s, nil
}
