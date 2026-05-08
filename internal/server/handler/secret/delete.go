package secret

import (
	"context"
	"errors"

	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) DeleteSecret(ctx context.Context, req *pb.DeleteSecretRequest) (*pb.DeleteSecretResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := parseSecretID(req.Id)
	if err != nil {
		return nil, err
	}

	if err := h.svc.Delete(ctx, userID, id); err != nil {
		if errors.Is(err, domain.ErrSecretNotFound) {
			return nil, status.Error(codes.NotFound, "secret not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete secret")
	}

	return &pb.DeleteSecretResponse{}, nil
}
