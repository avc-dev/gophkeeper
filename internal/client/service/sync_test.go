package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSyncer реализует интерфейс syncer для тестов.
type mockSyncer struct {
	pushErr    error
	syncErr    error
	hasPending bool
	pendingErr error
	pushCalls  int
	syncCalls  int
}

func (m *mockSyncer) PushPending(_ context.Context) error {
	m.pushCalls++
	return m.pushErr
}

func (m *mockSyncer) Sync(_ context.Context, _ []byte, _ *time.Time) error {
	m.syncCalls++
	return m.syncErr
}

func (m *mockSyncer) HasPending(_ context.Context) (bool, error) {
	return m.hasPending, m.pendingErr
}

// mockAuthReader реализует интерфейс authReader для тестов.
type mockAuthReader struct {
	token      string
	tokenErr   error
	lastSyncAt *time.Time
	setErr     error
	setCalled  bool
}

func (m *mockAuthReader) Token(_ context.Context) (string, error) {
	return m.token, m.tokenErr
}

func (m *mockAuthReader) GetLastSyncAt(_ context.Context) (*time.Time, error) {
	return m.lastSyncAt, nil
}

func (m *mockAuthReader) SetLastSyncAt(_ context.Context, t time.Time) error {
	m.setCalled = true
	m.lastSyncAt = &t
	return m.setErr
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestSyncServiceRunCycle(t *testing.T) {
	tests := []struct {
		name            string
		tokenErr        error
		pushErr         error
		syncErr         error
		hasPending      bool
		wantInterval    time.Duration
		wantPushCalls   int
		wantSyncCalls   int
		wantLastSyncSet bool
	}{
		{
			name:            "no pending — idle interval",
			hasPending:      false,
			wantInterval:    intervalIdle,
			wantPushCalls:   1,
			wantSyncCalls:   1,
			wantLastSyncSet: true,
		},
		{
			name:            "has pending — short interval",
			hasPending:      true,
			wantInterval:    intervalPending,
			wantPushCalls:   1,
			wantSyncCalls:   1,
			wantLastSyncSet: true,
		},
		{
			name:         "no token — offline interval",
			tokenErr:     errors.New("not logged in"),
			wantInterval: intervalOffline,
		},
		{
			name:          "push fails — offline interval",
			pushErr:       errors.New("connection refused"),
			wantInterval:  intervalOffline,
			wantPushCalls: 1,
		},
		{
			name:          "sync fails — offline interval",
			syncErr:       errors.New("connection refused"),
			wantInterval:  intervalOffline,
			wantPushCalls: 1,
			wantSyncCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secrets := &mockSyncer{
				pushErr:    tt.pushErr,
				syncErr:    tt.syncErr,
				hasPending: tt.hasPending,
			}
			auth := &mockAuthReader{
				token:    "valid-token",
				tokenErr: tt.tokenErr,
			}
			svc := NewSyncService(secrets, auth, []byte("master-key"), testLogger())

			got := svc.runCycle(context.Background())

			assert.Equal(t, tt.wantInterval, got)
			assert.Equal(t, tt.wantPushCalls, secrets.pushCalls)
			assert.Equal(t, tt.wantSyncCalls, secrets.syncCalls)
			assert.Equal(t, tt.wantLastSyncSet, auth.setCalled)
		})
	}
}

func TestSyncServiceStartStop(t *testing.T) {
	secrets := &mockSyncer{}
	auth := &mockAuthReader{token: "valid-token"}

	masterKey := []byte{1, 2, 3, 4}
	svc := NewSyncService(secrets, auth, masterKey, testLogger())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc.Start(ctx)

	// даём горутине отработать первый цикл.
	time.Sleep(50 * time.Millisecond)

	svc.Stop()

	// после Stop master key должен быть обнулён.
	assert.Equal(t, make([]byte, 4), svc.masterKey)
	// хотя бы один цикл успел выполниться.
	assert.GreaterOrEqual(t, secrets.pushCalls, 1)
}

func TestSyncServiceStopOnContextCancel(t *testing.T) {
	secrets := &mockSyncer{}
	auth := &mockAuthReader{token: "valid-token"}

	svc := NewSyncService(secrets, auth, []byte("key"), testLogger())

	ctx, cancel := context.WithCancel(context.Background())
	svc.Start(ctx)

	// отменяем контекст — горутина должна завершиться.
	cancel()

	// ждём завершения через done-канал с таймаутом.
	select {
	case <-svc.done:
		// ожидаемо
	case <-time.After(2 * time.Second):
		t.Fatal("sync goroutine did not stop after context cancel")
	}
}

func TestSyncServiceMasterKeyCopied(t *testing.T) {
	// оригинальный ключ не должен быть связан с внутренним буфером SyncService.
	original := []byte{0xAA, 0xBB, 0xCC}
	auth := &mockAuthReader{token: "tok"}
	svc := NewSyncService(&mockSyncer{}, auth, original, testLogger())

	// изменяем оригинал.
	original[0] = 0x00

	// внутренний ключ не должен измениться.
	assert.Equal(t, byte(0xAA), svc.masterKey[0])
}

func TestZeroKey(t *testing.T) {
	key := []byte{1, 2, 3, 4, 5}
	ZeroKey(key)
	assert.Equal(t, make([]byte, 5), key)
}

func TestLastSyncAtTracking(t *testing.T) {
	secrets := &mockSyncer{}
	auth := &mockAuthReader{token: "valid-token"}

	svc := NewSyncService(secrets, auth, []byte("key"), testLogger())
	svc.runCycle(context.Background())

	require.True(t, auth.setCalled)
	require.NotNil(t, auth.lastSyncAt)

	// второй цикл должен передавать since = time предыдущей синхронизации.
	secrets2 := &mockSyncer{}
	auth.setCalled = false
	svc2 := NewSyncService(secrets2, auth, []byte("key"), testLogger())
	svc2.runCycle(context.Background())

	assert.True(t, auth.setCalled)
	assert.Equal(t, 1, secrets2.syncCalls)
}
