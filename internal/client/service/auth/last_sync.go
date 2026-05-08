package auth

import (
	"context"
	"fmt"
	"time"
)

// GetLastSyncAt возвращает время последней успешной синхронизации или nil.
func (s *Service) GetLastSyncAt(ctx context.Context) (*time.Time, error) {
	v, err := s.authStore.Get(ctx, keyLastSyncAt)
	if err != nil {
		return nil, fmt.Errorf("read last_sync_at: %w", err)
	}
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return nil, fmt.Errorf("parse last_sync_at: %w", err)
	}
	return &t, nil
}

// SetLastSyncAt сохраняет время последней успешной синхронизации.
func (s *Service) SetLastSyncAt(ctx context.Context, t time.Time) error {
	if err := s.authStore.Set(ctx, keyLastSyncAt, t.UTC().Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("set last_sync_at: %w", err)
	}
	return nil
}
