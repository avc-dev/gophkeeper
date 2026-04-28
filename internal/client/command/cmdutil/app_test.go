package cmdutil

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	authsvc "github.com/avc-dev/gophkeeper/internal/client/service/auth"
	secretsvc "github.com/avc-dev/gophkeeper/internal/client/service/secret"
	"github.com/avc-dev/gophkeeper/internal/client/storage"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// ─── minimal mock gRPC clients ────────────────────────────────────────────────

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

type mockSecretsGRPC struct{}

func (m *mockSecretsGRPC) Ping(_ context.Context, _ *pb.PingRequest, _ ...grpc.CallOption) (*pb.PingResponse, error) {
	return nil, nil
}
func (m *mockSecretsGRPC) CreateSecret(_ context.Context, _ *pb.CreateSecretRequest, _ ...grpc.CallOption) (*pb.CreateSecretResponse, error) {
	return nil, errors.New("offline")
}
func (m *mockSecretsGRPC) GetSecret(_ context.Context, _ *pb.GetSecretRequest, _ ...grpc.CallOption) (*pb.GetSecretResponse, error) {
	return nil, errors.New("offline")
}
func (m *mockSecretsGRPC) ListSecrets(_ context.Context, _ *pb.ListSecretsRequest, _ ...grpc.CallOption) (*pb.ListSecretsResponse, error) {
	return nil, errors.New("offline")
}
func (m *mockSecretsGRPC) UpdateSecret(_ context.Context, _ *pb.UpdateSecretRequest, _ ...grpc.CallOption) (*pb.UpdateSecretResponse, error) {
	return nil, errors.New("offline")
}
func (m *mockSecretsGRPC) DeleteSecret(_ context.Context, _ *pb.DeleteSecretRequest, _ ...grpc.CallOption) (*pb.DeleteSecretResponse, error) {
	return nil, errors.New("offline")
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func newTestApp(t *testing.T, authMock *mockAuthGRPC) *App {
	t.Helper()
	db, err := storage.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return &App{
		AuthSvc:   authsvc.New(authMock, storage.NewAuthStorage(db)),
		SecretSvc: secretsvc.New(&mockSecretsGRPC{}, storage.NewSecretStorage(db)),
	}
}

// ─── ZeroKey ──────────────────────────────────────────────────────────────────

func TestZeroKey(t *testing.T) {
	key := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	ZeroKey(key)
	assert.Equal(t, make([]byte, 4), key)
}

// ─── NowUTC ───────────────────────────────────────────────────────────────────

func TestNowUTC_IsUTC(t *testing.T) {
	now := NowUTC()
	assert.Equal(t, time.UTC, now.Location())
	assert.WithinDuration(t, time.Now(), now, time.Second)
}

// ─── AddMasterPasswordFlag ────────────────────────────────────────────────────

func TestAddMasterPasswordFlag(t *testing.T) {
	var target string
	cmd := &cobra.Command{Use: "test", RunE: func(cmd *cobra.Command, args []string) error { return nil }}
	AddMasterPasswordFlag(cmd, &target)

	cmd.SetArgs([]string{"--master-password", "mypass"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "mypass", target)
}

// ─── AuthedContext ────────────────────────────────────────────────────────────

func TestAuthedContext_NotLoggedIn(t *testing.T) {
	app := newTestApp(t, &mockAuthGRPC{loginErr: errors.New("not called")})
	_, err := app.AuthedContext(context.Background())
	require.Error(t, err)
}

func TestAuthedContext_LoggedIn(t *testing.T) {
	authMock := &mockAuthGRPC{
		loginResp: &pb.LoginResponse{Token: "test-jwt", KdfSalt: make([]byte, 32)},
	}
	app := newTestApp(t, authMock)
	_, err := app.AuthSvc.Login(context.Background(), "u@u.com", "pass")
	require.NoError(t, err)

	ctx, err := app.AuthedContext(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, ctx)
}

// ─── ResolveMasterKey ─────────────────────────────────────────────────────────

func TestResolveMasterKey_NotLoggedIn(t *testing.T) {
	app := newTestApp(t, &mockAuthGRPC{})
	_, err := app.ResolveMasterKey(context.Background(), "some-password")
	require.Error(t, err)
}

func TestResolveMasterKey_LoggedIn(t *testing.T) {
	authMock := &mockAuthGRPC{
		loginResp: &pb.LoginResponse{Token: "tok", KdfSalt: make([]byte, 32)},
	}
	app := newTestApp(t, authMock)
	_, err := app.AuthSvc.Login(context.Background(), "u@u.com", "pass")
	require.NoError(t, err)

	key, err := app.ResolveMasterKey(context.Background(), "pass")
	require.NoError(t, err)
	assert.Len(t, key, 32)
}

// ─── ReadPassword (non-terminal path via stdin pipe) ─────────────────────────

func TestReadPassword_FromPipe(t *testing.T) {
	// In CI / test environments stdin is not a terminal, so ReadPassword falls
	// through to the fmt.Fscan branch. Redirect os.Stdin to a pipe to cover it.
	r, w, err := os.Pipe()
	require.NoError(t, err)

	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()

	go func() {
		_, _ = w.WriteString("pipepassword\n")
		w.Close()
	}()

	pwd, err := ReadPassword("Password: ")
	require.NoError(t, err)
	assert.Equal(t, "pipepassword", pwd)
}
