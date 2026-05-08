package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

const (
	// интервалы адаптивного polling'а.
	intervalPending = 10 * time.Second // есть несинхронизированные записи
	intervalIdle    = 60 * time.Second // всё синхронизировано
	intervalOffline = 5 * time.Minute  // сервер недоступен — редкие попытки
)

// syncer — интерфейсы зависимостей SyncService (для тестирования).
type syncer interface {
	PushPending(ctx context.Context) error
	Sync(ctx context.Context, masterKey []byte, since *time.Time) error
	HasPending(ctx context.Context) (bool, error)
}

type authReader interface {
	Token(ctx context.Context) (string, error)
	GetLastSyncAt(ctx context.Context) (*time.Time, error)
	SetLastSyncAt(ctx context.Context, t time.Time) error
}

// SyncService запускает фоновый цикл синхронизации с адаптивным интервалом.
// masterKey хранится в памяти только на время работы горутины и обнуляется при Stop().
type SyncService struct {
	secrets   syncer
	auth      authReader
	masterKey []byte
	log       *slog.Logger
	stop      chan struct{}
	done      chan struct{}
}

// NewSyncService создаёт SyncService.
// masterKey копируется внутрь; оригинальный буфер можно обнулить после вызова.
func NewSyncService(secrets syncer, auth authReader, masterKey []byte, log *slog.Logger) *SyncService {
	key := make([]byte, len(masterKey))
	copy(key, masterKey)
	return &SyncService{
		secrets:   secrets,
		auth:      auth,
		masterKey: key,
		log:       log,
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
	}
}

// Start запускает фоновую горутину. Возвращает немедленно.
func (s *SyncService) Start(ctx context.Context) {
	go s.loop(ctx)
}

// Stop сигнализирует горутине остановиться и ждёт её завершения.
// После возврата masterKey обнулён.
func (s *SyncService) Stop() {
	close(s.stop)
	<-s.done
}

// loop — основной цикл синхронизации.
func (s *SyncService) loop(ctx context.Context) {
	defer func() {
		ZeroKey(s.masterKey)
		close(s.done)
	}()

	// первая синхронизация сразу при запуске.
	interval := s.runCycle(ctx)

	for {
		select {
		case <-s.stop:
			return
		case <-ctx.Done():
			return
		case <-time.After(interval):
			interval = s.runCycle(ctx)
		}
	}
}

// runCycle выполняет один цикл: push pending → pull с сервера.
// Возвращает следующий интервал ожидания.
func (s *SyncService) runCycle(ctx context.Context) time.Duration {
	authedCtx, err := s.authedCtx(ctx)
	if err != nil {
		// не залогинены или сервер недоступен.
		return intervalOffline
	}

	// PUSH: отправляем локальные изменения.
	if err := s.secrets.PushPending(authedCtx); err != nil {
		s.log.Warn("sync push failed", "err", err)
		return intervalOffline
	}

	// PULL: получаем изменения с сервера.
	since, _ := s.auth.GetLastSyncAt(ctx)
	if err := s.secrets.Sync(authedCtx, s.masterKey, since); err != nil {
		s.log.Warn("sync pull failed", "err", err)
		return intervalOffline
	}

	// обновляем время последней синхронизации.
	now := time.Now().UTC()
	if err := s.auth.SetLastSyncAt(ctx, now); err != nil {
		s.log.Warn("failed to update last_sync_at", "err", err)
	}

	// выбираем следующий интервал в зависимости от наличия pending-записей.
	hasPending, _ := s.secrets.HasPending(ctx)
	if hasPending {
		return intervalPending
	}
	return intervalIdle
}

// authedCtx строит контекст с JWT токеном для gRPC вызовов.
// Дублирует логику из command/root.go, чтобы сервис не зависел от пакета command.
func (s *SyncService) authedCtx(ctx context.Context) (context.Context, error) {
	token, err := s.auth.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("get auth token: %w", err)
	}
	return ContextWithBearerToken(ctx, token), nil
}
