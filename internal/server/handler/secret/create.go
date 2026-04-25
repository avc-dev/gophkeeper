package secret

import (
	"context"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/avc-dev/gophkeeper/internal/server/handler"
	pb "github.com/avc-dev/gophkeeper/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *Handler) CreateSecret(ctx context.Context, req *pb.CreateSecretRequest) (*pb.CreateSecretResponse, error) {
	if req.Name == "" || req.EncryptedPayload == nil {
		return nil, status.Error(codes.InvalidArgument, "name and payload are required")
	}

	userID, ok := handler.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	sec, err := h.svc.Create(ctx, userID, &domain.Secret{
		Type:     typeToDomain(req.Type),
		Name:     req.Name,
		Payload:  req.EncryptedPayload,
		Metadata: req.Metadata,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create secret")
	}

	return &pb.CreateSecretResponse{
		Id:        sec.ID.String(),
		Version:   sec.Version,
		CreatedAt: timestamppb.New(sec.CreatedAt),
	}, nil
}
