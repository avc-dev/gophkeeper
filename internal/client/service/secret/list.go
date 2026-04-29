package secret

import (
	"context"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/client/storage"
	"github.com/avc-dev/gophkeeper/internal/domain"
)

// List возвращает все не удалённые секреты заданного типа из локального кеша.
func (s *Service) List(ctx context.Context, typ domain.SecretType) ([]*storage.LocalSecret, error) {
	return s.secretStore.List(ctx, typ)
}

// HasPending сообщает, есть ли записи ожидающие отправки на сервер.
func (s *Service) HasPending(ctx context.Context) (bool, error) {
	pending, err := s.secretStore.ListPending(ctx)
	if err != nil {
		return false, fmt.Errorf("has pending: %w", err)
	}
	return len(pending) > 0, nil
}
