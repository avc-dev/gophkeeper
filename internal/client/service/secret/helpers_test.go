package secret

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/avc-dev/gophkeeper/internal/client/storage"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ─── mock SecretsServiceClient ───────────────────────────────────────────────

type mockSecretsGRPC struct {
	createResp *pb.CreateSecretResponse
	createErr  error
	listResp   *pb.ListSecretsResponse
	listErr    error
	updateResp *pb.UpdateSecretResponse
	updateErr  error
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
	return m.updateResp, m.updateErr
}
func (m *mockSecretsGRPC) DeleteSecret(_ context.Context, _ *pb.DeleteSecretRequest, _ ...grpc.CallOption) (*pb.DeleteSecretResponse, error) {
	return m.deleteResp, m.deleteErr
}

// ─── helpers ──────────────────────────────────────────────────────────────────

var testKey = bytes.Repeat([]byte{0x42}, 32)

func newTestService(t *testing.T, client pb.SecretsServiceClient) *Service {
	t.Helper()
	svc, _ := newTestServiceWithDB(t, client)
	return svc
}

func newTestServiceWithDB(t *testing.T, client pb.SecretsServiceClient) (*Service, *sql.DB) {
	t.Helper()
	db, err := storage.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return New(client, storage.NewSecretStorage(db)), db
}

func offlineMock() *mockSecretsGRPC {
	return &mockSecretsGRPC{createErr: errors.New("offline")}
}

func nowProto() *timestamppb.Timestamp { return timestamppb.New(time.Now()) }
