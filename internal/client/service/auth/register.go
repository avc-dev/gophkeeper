package auth

import (
	"context"
	"fmt"

	pb "github.com/avc-dev/gophkeeper/proto"
)

func (s *Service) Register(ctx context.Context, email, password string) error {
	_, err := s.client.Register(ctx, &pb.RegisterRequest{Email: email, Password: password})
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}
	return nil
}
