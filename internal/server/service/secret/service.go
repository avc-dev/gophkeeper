package secret

import (
	"context"
	"time"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

// storage — локальный интерфейс; реализуется storage/secret.Storage.
type storage interface {
	Create(ctx context.Context, secret *domain.Secret) (*domain.Secret, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*domain.Secret, error)
	List(ctx context.Context, userID uuid.UUID, since *time.Time) ([]*domain.Secret, error)
	Update(ctx context.Context, userID, id uuid.UUID, payload []byte, metadata string, expectedVersion int64) (*domain.Secret, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

// Service реализует бизнес-логику работы с секретами на сервере.
type Service struct {
	secrets storage
}

// New создаёт новый Service с переданным хранилищем.
func New(secrets storage) *Service {
	return &Service{secrets: secrets}
}
