package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockService — заглушка сервиса для изолированного тестирования хендлера.
type mockService struct {
	token   string
	kdfSalt []byte
	err     error
}

func (m *mockService) Register(_ context.Context, _, _ string) (string, []byte, error) {
	return m.token, m.kdfSalt, m.err
}

func (m *mockService) Login(_ context.Context, _, _ string) (string, []byte, error) {
	return m.token, m.kdfSalt, m.err
}

func TestHandlerRegister(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.RegisterRequest
		mockErr   error
		wantCode  codes.Code
		wantToken string
	}{
		{
			name:      "успешная регистрация",
			req:       &pb.RegisterRequest{Email: "user@example.com", Password: "password"},
			wantCode:  codes.OK,
			wantToken: "token",
		},
		{
			name:     "пустой email",
			req:      &pb.RegisterRequest{Email: "", Password: "password"},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "пустой пароль",
			req:      &pb.RegisterRequest{Email: "user@example.com", Password: ""},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "email уже занят",
			req:      &pb.RegisterRequest{Email: "taken@example.com", Password: "password"},
			mockErr:  domain.ErrEmailTaken,
			wantCode: codes.AlreadyExists,
		},
		{
			name:     "внутренняя ошибка",
			req:      &pb.RegisterRequest{Email: "user@example.com", Password: "password"},
			mockErr:  errors.New("unexpected db error"),
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(&mockService{token: "token", kdfSalt: []byte("salt"), err: tt.mockErr})

			resp, err := h.Register(context.Background(), tt.req)

			if tt.wantCode == codes.OK {
				require.NoError(t, err)
				assert.Equal(t, tt.wantToken, resp.Token)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.wantCode, status.Code(err))
			}
		})
	}
}

func TestHandlerLogin(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.LoginRequest
		mockErr   error
		wantCode  codes.Code
		wantToken string
	}{
		{
			name:      "успешный вход",
			req:       &pb.LoginRequest{Email: "user@example.com", Password: "password"},
			wantCode:  codes.OK,
			wantToken: "token",
		},
		{
			name:     "пустой email",
			req:      &pb.LoginRequest{Email: "", Password: "password"},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "пустой пароль",
			req:      &pb.LoginRequest{Email: "user@example.com", Password: ""},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "неверные учётные данные",
			req:      &pb.LoginRequest{Email: "user@example.com", Password: "wrong"},
			mockErr:  errors.New("invalid email or password"),
			wantCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(&mockService{token: "token", kdfSalt: []byte("salt"), err: tt.mockErr})

			resp, err := h.Login(context.Background(), tt.req)

			if tt.wantCode == codes.OK {
				require.NoError(t, err)
				assert.Equal(t, tt.wantToken, resp.Token)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.wantCode, status.Code(err))
			}
		})
	}
}
