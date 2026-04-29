package secret

import (
	"context"
	"time"

	pb "github.com/avc-dev/gophkeeper/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) ListSecrets(ctx context.Context, req *pb.ListSecretsRequest) (*pb.ListSecretsResponse, error) {
	userID, err := extractUserID(ctx)
	if err != nil {
		return nil, err
	}

	var since *time.Time
	if req.Since != nil && !req.Since.AsTime().IsZero() {
		t := req.Since.AsTime()
		since = &t
	}

	secrets, err := h.svc.List(ctx, userID, since)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list secrets")
	}

	out := make([]*pb.Secret, len(secrets))
	for i, s := range secrets {
		out[i] = secretToProto(s)
	}

	return &pb.ListSecretsResponse{Secrets: out}, nil
}
