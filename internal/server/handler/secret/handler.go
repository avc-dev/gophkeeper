package secret

import (
	"context"
	"time"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/avc-dev/gophkeeper/internal/server/handler"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func extractUserID(ctx context.Context) (uuid.UUID, error) {
	id, ok := handler.UserIDFromContext(ctx)
	if !ok {
		return uuid.Nil, status.Error(codes.Unauthenticated, "missing user id")
	}
	return id, nil
}

func parseSecretID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, status.Error(codes.InvalidArgument, "invalid secret id")
	}
	return id, nil
}
