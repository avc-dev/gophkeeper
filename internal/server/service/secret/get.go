package secret

import (
	"context"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

// Get возвращает секрет по ID. Доступ ограничен userID — чужие секреты не возвращаются.
func (s *Service) Get(ctx context.Context, userID, id uuid.UUID) (*domain.Secret, error) {
	result, err := s.secrets.Get(ctx, userID, id)
	if err != nil {
		return nil, fmt.Errorf("get secret: %w", err)
	}
	return result, nil
}
