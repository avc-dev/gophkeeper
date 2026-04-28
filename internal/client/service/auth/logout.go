package auth

import (
	"context"
	"fmt"
)

func (s *Service) Logout(ctx context.Context) error {
	if err := s.authStore.Delete(ctx, keyJWT); err != nil {
		return fmt.Errorf("delete jwt: %w", err)
	}
	if err := s.authStore.Delete(ctx, keyKDFSalt); err != nil {
		return fmt.Errorf("delete kdf salt: %w", err)
	}
	return nil
}
