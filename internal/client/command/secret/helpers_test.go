package secret

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avc-dev/gophkeeper/internal/client/command/cmdutil"
	authsvc "github.com/avc-dev/gophkeeper/internal/client/service/auth"
	secretsvc "github.com/avc-dev/gophkeeper/internal/client/service/secret"
	"github.com/avc-dev/gophkeeper/internal/client/storage"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ─── mock gRPC clients ────────────────────────────────────────────────────────

type mockAuthGRPC struct {
	loginResp *pb.LoginResponse
	loginErr  error
}

func (m *mockAuthGRPC) Register(_ context.Context, _ *pb.RegisterRequest, _ ...grpc.CallOption) (*pb.RegisterResponse, error) {
	return &pb.RegisterResponse{}, nil
}
func (m *mockAuthGRPC) Login(_ context.Context, _ *pb.LoginRequest, _ ...grpc.CallOption) (*pb.LoginResponse, error) {
	return m.loginResp, m.loginErr
}

type mockSecretsGRPC struct {
	createResp *pb.CreateSecretResponse
	createErr  error
	listResp   *pb.ListSecretsResponse
	listErr    error
	deleteResp *pb.DeleteSecretResponse
	deleteErr  error
}

func (m *mockSecretsGRPC) Ping(_ context.Context, _ *pb.PingRequest, _ ...grpc.CallOption) (*pb.PingResponse, error) {
	return &pb.PingResponse{}, nil
}
func (m *mockSecretsGRPC) CreateSecret(_ context.Context, _ *pb.CreateSecretRequest, _ ...grpc.CallOption) (*pb.CreateSecretResponse, error) {
	return m.createResp, m.createErr
}
func (m *mockSecretsGRPC) GetSecret(_ context.Context, _ *pb.GetSecretRequest, _ ...grpc.CallOption) (*pb.GetSecretResponse, error) {
	return nil, errors.New("not implemented")
}
func (m *mockSecretsGRPC) ListSecrets(_ context.Context, _ *pb.ListSecretsRequest, _ ...grpc.CallOption) (*pb.ListSecretsResponse, error) {
	return m.listResp, m.listErr
}
func (m *mockSecretsGRPC) UpdateSecret(_ context.Context, _ *pb.UpdateSecretRequest, _ ...grpc.CallOption) (*pb.UpdateSecretResponse, error) {
	return nil, errors.New("not implemented")
}
func (m *mockSecretsGRPC) DeleteSecret(_ context.Context, _ *pb.DeleteSecretRequest, _ ...grpc.CallOption) (*pb.DeleteSecretResponse, error) {
	return m.deleteResp, m.deleteErr
}

// ─── helper constructors ──────────────────────────────────────────────────────

var testMasterPwd = "test-master-password"

// testKey — 32-byte master key derived from testMasterPwd with zero kdf_salt.
// Actual value is determined by crypto.DeriveKey; we just need consistency within tests.

func newTestApp(t *testing.T) (*cmdutil.App, *mockAuthGRPC, *mockSecretsGRPC) {
	t.Helper()
	db, err := storage.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	authMock := &mockAuthGRPC{
		loginResp: &pb.LoginResponse{Token: "test-jwt", KdfSalt: make([]byte, 32)},
	}
	secretsMock := &mockSecretsGRPC{
		createErr: errors.New("offline"),
		listErr:   errors.New("offline"),
	}

	app := &cmdutil.App{
		AuthSvc:   authsvc.New(authMock, storage.NewAuthStorage(db)),
		SecretSvc: secretsvc.New(secretsMock, storage.NewSecretStorage(db)),
	}
	return app, authMock, secretsMock
}

// loginApp populates JWT + kdf_salt in SQLite so commands needing auth work.
func loginApp(t *testing.T, app *cmdutil.App) {
	t.Helper()
	_, err := app.AuthSvc.Login(context.Background(), "user@example.com", testMasterPwd)
	require.NoError(t, err)
}

// addTestCredential adds a credential to local storage for Get/Copy/Delete tests.
func addTestCredential(t *testing.T, app *cmdutil.App, name string) {
	t.Helper()
	loginApp(t, app)
	masterKey, err := app.AuthSvc.DeriveMasterKey(context.Background(), testMasterPwd)
	require.NoError(t, err)
	require.NoError(t, app.SecretSvc.AddCredential(context.Background(), masterKey, name, "alice", "s3cr3t", "https://example.com", ""))
}

func nowProto() *timestamppb.Timestamp { return timestamppb.New(time.Now()) }

func newServerUUID() string { return uuid.New().String() }

// silentCmd sets up stdout/stderr buffers and returns the stdout buffer.
func silentCmd(buf *bytes.Buffer) *bytes.Buffer {
	if buf == nil {
		buf = &bytes.Buffer{}
	}
	return buf
}
