package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (s *storage) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u := &domain.User{}
	err := s.db.QueryRow(ctx,
		`SELECT id, email, password_hash, kdf_salt, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.KDFSalt, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}
