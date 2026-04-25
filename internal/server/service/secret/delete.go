package secret

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) Delete(ctx context.Context, userID, id uuid.UUID) error {
	if err := s.secrets.Delete(ctx, userID, id); err != nil {
		return fmt.Errorf("delete secret: %w", err)
	}
	return nil
}
