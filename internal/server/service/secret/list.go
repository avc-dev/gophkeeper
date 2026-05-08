package secret

import (
	"context"
	"fmt"
	"time"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

// List возвращает секреты пользователя, опционально фильтруя по времени обновления (since).
func (s *Service) List(ctx context.Context, userID uuid.UUID, since *time.Time) ([]*domain.Secret, error) {
	result, err := s.secrets.List(ctx, userID, since)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	return result, nil
}
