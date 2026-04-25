package user

import (
	"context"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

type storage struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) Storage {
	return &storage{db: db}
}
