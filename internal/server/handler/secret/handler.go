package secret

import (
	"context"
	"time"

	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/google/uuid"
)

// service — локальный интерфейс; реализуется service/secret.Service.
type service interface {
	Create(ctx context.Context, userID uuid.UUID, sec *domain.Secret) (*domain.Secret, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*domain.Secret, error)
	List(ctx context.Context, userID uuid.UUID, since *time.Time) ([]*domain.Secret, error)
	Update(ctx context.Context, userID, id uuid.UUID, payload []byte, metadata string, expectedVersion int64) (*domain.Secret, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

type Handler struct {
	pb.UnimplementedSecretsServiceServer
	svc service
}

func New(svc service) *Handler {
	return &Handler{svc: svc}
}
