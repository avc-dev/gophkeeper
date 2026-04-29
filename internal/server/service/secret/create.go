package secret

import (
	"context"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

// Create сохраняет новый секрет, проставляя ему userID.
func (s *Service) Create(ctx context.Context, userID uuid.UUID, sec *domain.Secret) (*domain.Secret, error) {
	sec.UserID = userID
	result, err := s.secrets.Create(ctx, sec)
	if err != nil {
		return nil, fmt.Errorf("create secret: %w", err)
	}
	return result, nil
}
