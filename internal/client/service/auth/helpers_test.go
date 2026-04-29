package auth

import (
	"context"
	"testing"

	"github.com/avc-dev/gophkeeper/internal/client/storage"
	pb "github.com/avc-dev/gophkeeper/proto"
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
