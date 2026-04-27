package secret

import (
	"context"
	"time"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	Create(ctx context.Context, secret *domain.Secret) (*domain.Secret, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*domain.Secret, error)
	List(ctx context.Context, userID uuid.UUID, since *time.Time) ([]*domain.Secret, error)
	Update(ctx context.Context, userID, id uuid.UUID, payload []byte, metadata string, expectedVersion int64) (*domain.Secret, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

type storage struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) Storage {
	return &storage{db: db}
}
