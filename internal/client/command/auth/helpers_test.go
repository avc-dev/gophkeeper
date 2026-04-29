package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	authsvc "github.com/avc-dev/gophkeeper/internal/client/service/auth"
	secretsvc "github.com/avc-dev/gophkeeper/internal/client/service/secret"
	"github.com/avc-dev/gophkeeper/internal/client/storage"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// ─── mock gRPC clients ────────────────────────────────────────────────────────

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

type mockSecretsGRPC struct{ listErr error }

func (m *mockSecretsGRPC) Ping(_ context.Context, _ *pb.PingRequest, _ ...grpc.CallOption) (*pb.PingResponse, error) {
	return &pb.PingResponse{}, nil
}
func (m *mockSecretsGRPC) CreateSecret(_ context.Context, _ *pb.CreateSecretRequest, _ ...grpc.CallOption) (*pb.CreateSecretResponse, error) {
	return nil, errors.New("offline")
}
func (m *mockSecretsGRPC) GetSecret(_ context.Context, _ *pb.GetSecretRequest, _ ...grpc.CallOption) (*pb.GetSecretResponse, error) {
	return nil, errors.New("offline")
}
func (m *mockSecretsGRPC) ListSecrets(_ context.Context, _ *pb.ListSecretsRequest, _ ...grpc.CallOption) (*pb.ListSecretsResponse, error) {
	return nil, m.listErr
}
func (m *mockSecretsGRPC) UpdateSecret(_ context.Context, _ *pb.UpdateSecretRequest, _ ...grpc.CallOption) (*pb.UpdateSecretResponse, error) {
	return nil, errors.New("offline")
}
func (m *mockSecretsGRPC) DeleteSecret(_ context.Context, _ *pb.DeleteSecretRequest, _ ...grpc.CallOption) (*pb.DeleteSecretResponse, error) {
	return nil, errors.New("offline")
}

// ─── helper constructors ──────────────────────────────────────────────────────

func newTestApp(t *testing.T, authClient pb.AuthServiceClient, secretsClient pb.SecretsServiceClient) *cmdutil.App {
	t.Helper()
	db, err := storage.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return &cmdutil.App{
		AuthSvc:   authsvc.New(authClient, storage.NewAuthStorage(db)),
		SecretSvc: secretsvc.New(secretsClient, storage.NewSecretStorage(db)),
	}
}

// loginApp pre-populates token and kdf_salt in DB so AuthedContext / ResolveMasterKey work.
func loginApp(t *testing.T, app *cmdutil.App, authMock *mockAuthGRPC) {
	t.Helper()
	authMock.loginResp = &pb.LoginResponse{Token: "test-jwt", KdfSalt: make([]byte, 32)}
	_, err := app.AuthSvc.Login(context.Background(), "user@example.com", "password")
	require.NoError(t, err)
}
