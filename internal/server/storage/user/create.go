package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *storage) Create(ctx context.Context, u *domain.User) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, kdf_salt, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		u.ID, u.Email, u.PasswordHash, u.KDFSalt, u.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		// код 23505 — нарушение unique constraint
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailTaken
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}
