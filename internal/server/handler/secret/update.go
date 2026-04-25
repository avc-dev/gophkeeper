package secret

import (
	"context"
	"errors"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/avc-dev/gophkeeper/internal/server/handler"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *Handler) UpdateSecret(ctx context.Context, req *pb.UpdateSecretRequest) (*pb.UpdateSecretResponse, error) {
	if req.EncryptedPayload == nil {
		return nil, status.Error(codes.InvalidArgument, "payload is required")
	}

	userID, ok := handler.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid secret id")
	}

	sec, err := h.svc.Update(ctx, userID, id, req.EncryptedPayload, req.Metadata, req.ExpectedVersion)
	if err != nil {
		if errors.Is(err, domain.ErrSecretNotFound) {
			return nil, status.Error(codes.NotFound, "secret not found")
		}
		if errors.Is(err, domain.ErrVersionConflict) {
			// Aborted — стандартный gRPC код для конфликта оптимистичной блокировки
			return nil, status.Error(codes.Aborted, "version conflict")
		}
		return nil, status.Error(codes.Internal, "failed to update secret")
	}

	return &pb.UpdateSecretResponse{
		Version:   sec.Version,
		UpdatedAt: timestamppb.New(sec.UpdatedAt),
	}, nil
}
