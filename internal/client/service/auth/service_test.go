package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avc-dev/gophkeeper/internal/client/storage"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// ─── mock AuthServiceClient ───────────────────────────────────────────────────

type mockAuthGRPC struct {
	registerErr error
	loginResp   *pb.LoginResponse
	loginErr    error
}

func (m *mockAuthGRPC) Register(_ context.Context, _ *pb.RegisterRequest, _ ...grpc.CallOption) (*pb.RegisterResponse, error) {
	return &pb.RegisterResponse{}, m.registerErr
}

func (m *mockAuthGRPC) Login(_ context.Context, _ *pb.LoginRequest, _ ...grpc.CallOption) (*pb.LoginResponse, error) {
	return m.loginResp, m.loginErr
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func newTestService(t *testing.T, client pb.AuthServiceClient) *Service {
	t.Helper()
	db, err := storage.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return New(client, storage.NewAuthStorage(db))
}

func validLoginResp(token string) *pb.LoginResponse {
	return &pb.LoginResponse{Token: token, KdfSalt: make([]byte, 32)}
}

// ─── Register ─────────────────────────────────────────────────────────────────

func TestRegister(t *testing.T) {
	tests := []struct {
		name    string
		grpcErr error
		wantErr bool
	}{
		{name: "success"},
		{name: "server unavailable", grpcErr: errors.New("unavailable"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, &mockAuthGRPC{registerErr: tt.grpcErr})
			err := svc.Register(context.Background(), "user@example.com", "password")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ─── Login ────────────────────────────────────────────────────────────────────

func TestLogin(t *testing.T) {
	tests := []struct {
		name       string
		resp       *pb.LoginResponse
		grpcErr    error
		wantErr    bool
		wantKeyLen int
	}{
		{
			name:       "success — returns 32-byte master key",
			resp:       validLoginResp("jwt.tok.en"),
			wantKeyLen: 32,
		},
		{
			name:    "server error",
			grpcErr: errors.New("unauthorized"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, &mockAuthGRPC{loginResp: tt.resp, loginErr: tt.grpcErr})
			key, err := svc.Login(context.Background(), "user@example.com", "password")
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, key, tt.wantKeyLen)
		})
	}
}

// ─── DeriveMasterKey ──────────────────────────────────────────────────────────

func TestDeriveMasterKey(t *testing.T) {
	tests := []struct {
		name       string
		loginFirst bool
		wantErr    error
		wantKeyLen int
	}{
		{
			name:    "ErrNotLoggedIn before login",
			wantErr: ErrNotLoggedIn,
		},
		{
			name:       "returns deterministic 32-byte key after login",
			loginFirst: true,
			wantKeyLen: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, &mockAuthGRPC{loginResp: validLoginResp("tok")})
			if tt.loginFirst {
				_, err := svc.Login(context.Background(), "user@example.com", "pass")
				require.NoError(t, err)
			}

			key, err := svc.DeriveMasterKey(context.Background(), "pass")
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, key, tt.wantKeyLen)

			key2, err := svc.DeriveMasterKey(context.Background(), "pass")
			require.NoError(t, err)
			assert.Equal(t, key, key2)
		})
	}
}

// ─── Token ────────────────────────────────────────────────────────────────────

func TestToken(t *testing.T) {
	tests := []struct {
		name       string
		loginFirst bool
		loginToken string
		wantErr    error
		wantToken  string
	}{
		{
			name:    "ErrNotLoggedIn before login",
			wantErr: ErrNotLoggedIn,
		},
		{
			name:       "returns stored token after login",
			loginFirst: true,
			loginToken: "my.jwt.token",
			wantToken:  "my.jwt.token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := validLoginResp(tt.loginToken)
			svc := newTestService(t, &mockAuthGRPC{loginResp: resp})
			if tt.loginFirst {
				_, err := svc.Login(context.Background(), "user@example.com", "pass")
				require.NoError(t, err)
			}

			tok, err := svc.Token(context.Background())
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantToken, tok)
		})
	}
}

// ─── GetLastSyncAt / SetLastSyncAt ────────────────────────────────────────────

func TestLastSyncAt(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	later := now.Add(time.Hour)

	tests := []struct {
		name     string
		setTime  *time.Time
		wantNil  bool
		wantTime *time.Time
	}{
		{name: "nil before first sync", wantNil: true},
		{name: "stores and retrieves time", setTime: &now, wantTime: &now},
		{name: "overwrite updates time", setTime: &later, wantTime: &later},
	}

	svc := newTestService(t, &mockAuthGRPC{})
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setTime != nil {
				require.NoError(t, svc.SetLastSyncAt(ctx, *tt.setTime))
			}
			got, err := svc.GetLastSyncAt(ctx)
			require.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.WithinDuration(t, *tt.wantTime, *got, time.Second)
		})
	}
}

// ─── Logout ───────────────────────────────────────────────────────────────────

func TestLogout(t *testing.T) {
	tests := []struct {
		name    string
		checkFn func(t *testing.T, svc *Service)
	}{
		{
			name: "token available before logout",
			checkFn: func(t *testing.T, svc *Service) {
				t.Helper()
				_, err := svc.Token(context.Background())
				require.NoError(t, err)
			},
		},
		{
			name: "token not available after logout",
			checkFn: func(t *testing.T, svc *Service) {
				t.Helper()
				require.NoError(t, svc.Logout(context.Background()))
				_, err := svc.Token(context.Background())
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrNotLoggedIn)
			},
		},
		{
			name: "kdf_salt not available after logout",
			checkFn: func(t *testing.T, svc *Service) {
				t.Helper()
				_, err := svc.DeriveMasterKey(context.Background(), "pass")
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrNotLoggedIn)
			},
		},
	}

	svc := newTestService(t, &mockAuthGRPC{loginResp: validLoginResp("tok")})
	_, err := svc.Login(context.Background(), "user@example.com", "pass")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checkFn(t, svc)
		})
	}
}
