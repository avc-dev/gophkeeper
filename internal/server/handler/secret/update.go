package secret

import (
	"context"
	"errors"

	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *Handler) UpdateSecret(ctx context.Context, req *pb.UpdateSecretRequest) (*pb.UpdateSecretResponse, error) {
	if req.EncryptedPayload == nil {
		return nil, status.Error(codes.InvalidArgument, "payload is required")
	}

	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := parseSecretID(req.Id)
	if err != nil {
		return nil, err
	}

	sec, err := h.svc.Update(ctx, userID, id, req.EncryptedPayload, req.Metadata, req.ExpectedVersion)
	if err != nil {
		if errors.Is(err, domain.ErrSecretNotFound) {
			return nil, status.Error(codes.NotFound, "secret not found")
		}
		if errors.Is(err, domain.ErrVersionConflict) {
			return nil, status.Error(codes.Aborted, "version conflict")
		}
		return nil, status.Error(codes.Internal, "failed to update secret")
	}

	return &pb.UpdateSecretResponse{
		Version:   sec.Version,
		UpdatedAt: timestamppb.New(sec.UpdatedAt),
	}, nil
}
