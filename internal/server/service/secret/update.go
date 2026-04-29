package secret

import (
	"context"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

// Update обновляет зашифрованное содержимое секрета.
// expectedVersion — оптимистичная блокировка: возвращает ошибку при конфликте версий.
func (s *Service) Update(ctx context.Context, userID, id uuid.UUID, payload []byte, metadata string, expectedVersion int64) (*domain.Secret, error) {
	result, err := s.secrets.Update(ctx, userID, id, payload, metadata, expectedVersion)
	if err != nil {
		return nil, fmt.Errorf("update secret: %w", err)
	}
	return result, nil
}
