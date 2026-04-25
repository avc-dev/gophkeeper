package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStorage — in-memory реализация userStorage для тестов.
type mockStorage struct {
	users map[string]*domain.User
	err   error // если задан — все вызовы возвращают эту ошибку
}

func newMockStorage() *mockStorage {
	return &mockStorage{users: make(map[string]*domain.User)}
}

func (m *mockStorage) Create(_ context.Context, u *domain.User) error {
	if m.err != nil {
		return m.err
	}
	if _, exists := m.users[u.Email]; exists {
		return domain.ErrEmailTaken
	}
	m.users[u.Email] = u
	return nil
}

func (m *mockStorage) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	u, ok := m.users[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func newTestService(store *mockStorage) *Service {
	return New(store, "test-secret-key")
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		setupFn   func(*mockStorage)
		wantErr   bool
		wantErrIs error
	}{
		{
			name:     "успешная регистрация",
			email:    "user@example.com",
			password: "password123",
		},
		{
			name:     "email уже занят",
			email:    "taken@example.com",
			password: "password123",
			setupFn: func(s *mockStorage) {
				// предварительно занимаем email
				_ = s.Create(context.Background(), &domain.User{
					ID:    uuid.New(),
					Email: "taken@example.com",
				})
			},
			wantErr:   true,
			wantErrIs: domain.ErrEmailTaken,
		},
		{
			name:     "ошибка хранилища",
			email:    "user@example.com",
			password: "password123",
			setupFn: func(s *mockStorage) {
				s.err = errors.New("db is down")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockStorage()
			if tt.setupFn != nil {
				tt.setupFn(store)
			}
			svc := newTestService(store)

			token, salt, err := svc.Register(context.Background(), tt.email, tt.password)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					assert.ErrorIs(t, err, tt.wantErrIs)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, token)
			assert.NotEmpty(t, salt)
		})
	}
}

func TestLogin(t *testing.T) {
	const (
		existingEmail    = "user@example.com"
		existingPassword = "password123"
	)

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{
			name:     "успешный вход",
			email:    existingEmail,
			password: existingPassword,
		},
		{
			name:     "неверный пароль",
			email:    existingEmail,
			password: "wrongpassword",
			wantErr:  true,
		},
		{
			name:     "пользователь не найден",
			email:    "nobody@example.com",
			password: existingPassword,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockStorage()
			svc := newTestService(store)

			// предварительно регистрируем пользователя
			_, _, err := svc.Register(context.Background(), existingEmail, existingPassword)
			require.NoError(t, err)

			token, salt, err := svc.Login(context.Background(), tt.email, tt.password)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, token)
			assert.NotEmpty(t, salt)
		})
	}
}

func TestValidateToken(t *testing.T) {
	svc := newTestService(newMockStorage())
	otherSvc := New(newMockStorage(), "different-secret")

	// получаем валидный токен через Register
	validToken, _, err := svc.Register(context.Background(), "user@example.com", "password")
	require.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		svc     *Service
		wantErr bool
	}{
		{
			name:  "валидный токен",
			token: validToken,
			svc:   svc,
		},
		{
			name:    "невалидная строка",
			token:   "not.a.token",
			svc:     svc,
			wantErr: true,
		},
		{
			name:    "токен подписан другим секретом",
			token:   validToken,
			svc:     otherSvc,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := tt.svc.ValidateToken(tt.token)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, uuid.Nil, id)
				return
			}

			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, id)
		})
	}
}
