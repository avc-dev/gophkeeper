package secret

import (
	"context"
	"errors"

	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) GetSecret(ctx context.Context, req *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := parseSecretID(req.Id)
	if err != nil {
		return nil, err
	}

	sec, err := h.svc.Get(ctx, userID, id)
	if err != nil {
		if errors.Is(err, domain.ErrSecretNotFound) {
			return nil, status.Error(codes.NotFound, "secret not found")
		}
		return nil, status.Error(codes.Internal, "failed to get secret")
	}

	return &pb.GetSecretResponse{Secret: secretToProto(sec)}, nil
}
