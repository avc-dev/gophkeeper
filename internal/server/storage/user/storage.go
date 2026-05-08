package user

import (
	"context"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Storage описывает операции хранилища пользователей на сервере (PostgreSQL).
type Storage interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

type storage struct {
	db *pgxpool.Pool
}

// New создаёт Storage поверх пула соединений PostgreSQL.
func New(db *pgxpool.Pool) Storage {
	return &storage{db: db}
}
