package secret

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStorage — заглушка хранилища для изолированного тестирования сервиса.
type mockStorage struct {
	secret  *domain.Secret
	secrets []*domain.Secret
	err     error
}

func (m *mockStorage) Create(_ context.Context, sec *domain.Secret) (*domain.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.secret, nil
}

func (m *mockStorage) Get(_ context.Context, _, _ uuid.UUID) (*domain.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.secret, nil
}

func (m *mockStorage) List(_ context.Context, _ uuid.UUID, _ *time.Time) ([]*domain.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.secrets, nil
}

func (m *mockStorage) Update(_ context.Context, _, _ uuid.UUID, _ []byte, _ string, _ int64) (*domain.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.secret, nil
}

func (m *mockStorage) Delete(_ context.Context, _, _ uuid.UUID) error {
	return m.err
}

// фиксированные UUID для детерминированных тестов.
var (
	testUserID   = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	testSecretID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

func testSecret() *domain.Secret {
	return &domain.Secret{
		ID:        testSecretID,
		UserID:    testUserID,
		Type:      domain.SecretTypeCredential,
		Name:      "github",
		Payload:   []byte("encrypted"),
		Metadata:  `{"url":"https://github.com"}`,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		sec     *domain.Secret
		mockErr error
		wantErr bool
	}{
		{
			name: "успешное создание",
			sec:  &domain.Secret{Type: domain.SecretTypeCredential, Name: "github", Payload: []byte("enc")},
		},
		{
			name:    "ошибка хранилища",
			sec:     &domain.Secret{Type: domain.SecretTypeCredential, Name: "github", Payload: []byte("enc")},
			mockErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(&mockStorage{secret: testSecret(), err: tt.mockErr})
			got, err := svc.Create(context.Background(), testUserID, tt.sec)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			// Create должен проставить userID в переданный объект.
			assert.Equal(t, testUserID, tt.sec.UserID)
			assert.Equal(t, testSecretID, got.ID)
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{
			name: "успешное получение",
		},
		{
			name:    "секрет не найден",
			mockErr: domain.ErrSecretNotFound,
			wantErr: true,
		},
		{
			name:    "ошибка хранилища",
			mockErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(&mockStorage{secret: testSecret(), err: tt.mockErr})
			got, err := svc.Get(context.Background(), testUserID, testSecretID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testSecretID, got.ID)
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name    string
		since   *time.Time
		mockErr error
		wantLen int
		wantErr bool
	}{
		{
			name:    "все секреты",
			since:   nil,
			wantLen: 2,
		},
		{
			name:    "с фильтром по времени",
			since:   func() *time.Time { t := time.Now().Add(-time.Hour); return &t }(),
			wantLen: 2,
		},
		{
			name:    "ошибка хранилища",
			mockErr: errors.New("db error"),
			wantErr: true,
		},
	}

	twoSecrets := []*domain.Secret{testSecret(), testSecret()}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(&mockStorage{secrets: twoSecrets, err: tt.mockErr})
			got, err := svc.List(context.Background(), testUserID, tt.since)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name            string
		payload         []byte
		metadata        string
		expectedVersion int64
		mockErr         error
		wantErr         bool
	}{
		{
			name:            "успешное обновление",
			payload:         []byte("new-enc"),
			metadata:        `{"note":"updated"}`,
			expectedVersion: 1,
		},
		{
			name:            "конфликт версий",
			payload:         []byte("new-enc"),
			expectedVersion: 99,
			mockErr:         domain.ErrVersionConflict,
			wantErr:         true,
		},
		{
			name:    "секрет не найден",
			payload: []byte("new-enc"),
			mockErr: domain.ErrSecretNotFound,
			wantErr: true,
		},
		{
			name:    "ошибка хранилища",
			payload: []byte("new-enc"),
			mockErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(&mockStorage{secret: testSecret(), err: tt.mockErr})
			got, err := svc.Update(context.Background(), testUserID, testSecretID, tt.payload, tt.metadata, tt.expectedVersion)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				// sentinel-ошибки должны пробрасываться через обёртку.
				if tt.mockErr == domain.ErrVersionConflict {
					assert.True(t, errors.Is(err, domain.ErrVersionConflict))
				}
				if tt.mockErr == domain.ErrSecretNotFound {
					assert.True(t, errors.Is(err, domain.ErrSecretNotFound))
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testSecretID, got.ID)
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{
			name: "успешное удаление",
		},
		{
			name:    "секрет не найден",
			mockErr: domain.ErrSecretNotFound,
			wantErr: true,
		},
		{
			name:    "ошибка хранилища",
			mockErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(&mockStorage{err: tt.mockErr})
			err := svc.Delete(context.Background(), testUserID, testSecretID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.mockErr == domain.ErrSecretNotFound {
					assert.True(t, errors.Is(err, domain.ErrSecretNotFound))
				}
				return
			}
			require.NoError(t, err)
		})
	}
}
