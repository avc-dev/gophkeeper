package auth

import (
	"context"
	"errors"

	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	token, salt, err := h.svc.Register(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrEmailTaken) {
			return nil, status.Error(codes.AlreadyExists, "email already taken")
		}
		return nil, status.Error(codes.Internal, "registration failed")
	}

	return &pb.RegisterResponse{Token: token, KdfSalt: salt}, nil
}
