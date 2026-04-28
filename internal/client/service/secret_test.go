package service

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avc-dev/gophkeeper/internal/client/storage"
	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

// testKey — 32-байтный мастер-ключ для AES-256-GCM в тестах.
var testKey = bytes.Repeat([]byte{0x42}, 32)

// newTestSecretService возвращает SecretService с in-memory SQLite.
func newTestSecretService(t *testing.T, client pb.SecretsServiceClient) *SecretService {
	t.Helper()
	db, err := storage.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewSecretService(client, storage.NewSecretStorage(db))
}

// offlineMock возвращает mock, у которого CreateSecret всегда возвращает ошибку (offline).
func offlineMock() *mockSecretsGRPC {
	return &mockSecretsGRPC{createErr: errors.New("offline")}
}

// nowProto — текущее время как protobuf Timestamp.
func nowProto() *timestamppb.Timestamp { return timestamppb.New(time.Now()) }

// ─── Add ──────────────────────────────────────────────────────────────────────

func TestSecretServiceAdd(t *testing.T) {
	serverUUID := uuid.New().String()

	tests := []struct {
		name           string
		mock           *mockSecretsGRPC
		add            func(svc *SecretService) error
		secretName     string
		secretType     domain.SecretType
		wantSyncStatus storage.SyncStatus
	}{
		{
			name: "credential — server online → synced",
			mock: &mockSecretsGRPC{
				createResp: &pb.CreateSecretResponse{Id: serverUUID, Version: 1, CreatedAt: nowProto()},
			},
			add: func(svc *SecretService) error {
				return svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "https://github.com", "")
			},
			secretName:     "gh",
			secretType:     domain.SecretTypeCredential,
			wantSyncStatus: storage.SyncStatusSynced,
		},
		{
			name: "credential — server offline → pending",
			mock: offlineMock(),
			add: func(svc *SecretService) error {
				return svc.AddCredential(context.Background(), testKey, "gh-offline", "alice", "pass", "", "")
			},
			secretName:     "gh-offline",
			secretType:     domain.SecretTypeCredential,
			wantSyncStatus: storage.SyncStatusPending,
		},
		{
			name: "card — offline",
			mock: offlineMock(),
			add: func(svc *SecretService) error {
				return svc.AddCard(context.Background(), testKey, "visa", "4532015112830366", "JOHN DOE", "12/26", "123", "Tinkoff", "")
			},
			secretName:     "visa",
			secretType:     domain.SecretTypeCard,
			wantSyncStatus: storage.SyncStatusPending,
		},
		{
			name: "text — offline",
			mock: offlineMock(),
			add: func(svc *SecretService) error {
				return svc.AddText(context.Background(), testKey, "note", "my secret note", "")
			},
			secretName:     "note",
			secretType:     domain.SecretTypeText,
			wantSyncStatus: storage.SyncStatusPending,
		},
		{
			name: "binary — offline",
			mock: offlineMock(),
			add: func(svc *SecretService) error {
				return svc.AddBinary(context.Background(), testKey, "key", "key.pem", []byte{0x01, 0x02}, "")
			},
			secretName:     "key",
			secretType:     domain.SecretTypeBinary,
			wantSyncStatus: storage.SyncStatusPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestSecretService(t, tt.mock)

			err := tt.add(svc)
			require.NoError(t, err)

			sec, err := svc.secretStore.GetByName(context.Background(), tt.secretName, tt.secretType)
			require.NoError(t, err)
			assert.Equal(t, tt.wantSyncStatus, sec.SyncStatus)
		})
	}
}

// ─── Get ──────────────────────────────────────────────────────────────────────

func TestSecretServiceGet(t *testing.T) {
	wrongKey := bytes.Repeat([]byte{0xDE}, 32)

	tests := []struct {
		name    string
		setup   func(t *testing.T, svc *SecretService)
		get     func(svc *SecretService) error
		wantErr bool
	}{
		{
			name: "credential — decrypts correctly",
			setup: func(t *testing.T, svc *SecretService) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "s3cr3t", "https://example.com", ""))
			},
			get: func(svc *SecretService) error {
				got, err := svc.GetCredential(context.Background(), testKey, "gh")
				if err != nil {
					return err
				}
				assert.Equal(t, "alice", got.Login)
				assert.Equal(t, "s3cr3t", got.Password)
				return nil
			},
		},
		{
			name: "credential — wrong master key",
			setup: func(t *testing.T, svc *SecretService) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
			},
			get: func(svc *SecretService) error {
				_, err := svc.GetCredential(context.Background(), wrongKey, "gh")
				return err
			},
			wantErr: true,
		},
		{
			name:  "credential — not found",
			setup: func(t *testing.T, svc *SecretService) {},
			get: func(svc *SecretService) error {
				_, err := svc.GetCredential(context.Background(), testKey, "nonexistent")
				return err
			},
			wantErr: true,
		},
		{
			name: "card — decrypts correctly",
			setup: func(t *testing.T, svc *SecretService) {
				t.Helper()
				require.NoError(t, svc.AddCard(context.Background(), testKey, "visa", "4532015112830366", "JOHN DOE", "12/26", "123", "Tinkoff", ""))
			},
			get: func(svc *SecretService) error {
				got, err := svc.GetCard(context.Background(), testKey, "visa")
				if err != nil {
					return err
				}
				assert.Equal(t, "4532015112830366", got.Number)
				assert.Equal(t, "123", got.CVV)
				return nil
			},
		},
		{
			name: "text — decrypts correctly",
			setup: func(t *testing.T, svc *SecretService) {
				t.Helper()
				require.NoError(t, svc.AddText(context.Background(), testKey, "note", "top secret text", ""))
			},
			get: func(svc *SecretService) error {
				got, err := svc.GetText(context.Background(), testKey, "note")
				if err != nil {
					return err
				}
				assert.Equal(t, "top secret text", got.Content)
				return nil
			},
		},
		{
			name: "binary — decrypts correctly",
			setup: func(t *testing.T, svc *SecretService) {
				t.Helper()
				require.NoError(t, svc.AddBinary(context.Background(), testKey, "file", "data.bin", []byte{0xDE, 0xAD, 0xBE, 0xEF}, ""))
			},
			get: func(svc *SecretService) error {
				got, err := svc.GetBinary(context.Background(), testKey, "file")
				if err != nil {
					return err
				}
				assert.Equal(t, "data.bin", got.Filename)
				assert.Equal(t, []byte{0xDE, 0xAD, 0xBE, 0xEF}, got.Data)
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestSecretService(t, offlineMock())
			tt.setup(t, svc)

			err := tt.get(svc)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ─── List ─────────────────────────────────────────────────────────────────────

func TestSecretServiceList(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, svc *SecretService)
		filter  domain.SecretType
		wantLen int
	}{
		{
			name:    "empty store",
			setup:   func(t *testing.T, svc *SecretService) {},
			wantLen: 0,
		},
		{
			name: "all types",
			setup: func(t *testing.T, svc *SecretService) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
				require.NoError(t, svc.AddText(context.Background(), testKey, "note", "text", ""))
			},
			wantLen: 2,
		},
		{
			name: "filter by credential type",
			setup: func(t *testing.T, svc *SecretService) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
				require.NoError(t, svc.AddText(context.Background(), testKey, "note", "text", ""))
			},
			filter:  domain.SecretTypeCredential,
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestSecretService(t, offlineMock())
			tt.setup(t, svc)

			list, err := svc.List(context.Background(), tt.filter)
			require.NoError(t, err)
			assert.Len(t, list, tt.wantLen)
		})
	}
}

// ─── HasPending ───────────────────────────────────────────────────────────────

func TestSecretServiceHasPending(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, svc *SecretService)
		wantPending bool
	}{
		{
			name:        "empty store",
			setup:       func(t *testing.T, svc *SecretService) {},
			wantPending: false,
		},
		{
			name: "after offline add",
			setup: func(t *testing.T, svc *SecretService) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
			},
			wantPending: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestSecretService(t, offlineMock())
			tt.setup(t, svc)

			has, err := svc.HasPending(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.wantPending, has)
		})
	}
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func TestSecretServiceDelete(t *testing.T) {
	syncedServerID := uuid.New().String()

	tests := []struct {
		name    string
		mock    *mockSecretsGRPC
		setup   func(t *testing.T, svc *SecretService)
		delete  func(t *testing.T, svc *SecretService) error
		wantErr bool
	}{
		{
			name: "never synced — local soft delete, GetByName returns not found",
			mock: offlineMock(),
			setup: func(t *testing.T, svc *SecretService) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
			},
			delete: func(t *testing.T, svc *SecretService) error {
				err := svc.Delete(context.Background(), "gh", domain.SecretTypeCredential)
				if err != nil {
					return err
				}
				// GetByName фильтрует deleted=0, поэтому удалённая запись недоступна.
				_, getErr := svc.secretStore.GetByName(context.Background(), "gh", domain.SecretTypeCredential)
				require.ErrorIs(t, getErr, domain.ErrSecretNotFound)
				return nil
			},
		},
		{
			name: "synced — sends delete to server",
			mock: &mockSecretsGRPC{
				createResp: &pb.CreateSecretResponse{Id: syncedServerID, Version: 1, CreatedAt: nowProto()},
				deleteResp: &pb.DeleteSecretResponse{},
			},
			setup: func(t *testing.T, svc *SecretService) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
			},
			delete: func(t *testing.T, svc *SecretService) error {
				return svc.Delete(context.Background(), "gh", domain.SecretTypeCredential)
			},
		},
		{
			name:  "not found — returns error",
			mock:  offlineMock(),
			setup: func(t *testing.T, svc *SecretService) {},
			delete: func(t *testing.T, svc *SecretService) error {
				return svc.Delete(context.Background(), "nonexistent", domain.SecretTypeCredential)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestSecretService(t, tt.mock)
			tt.setup(t, svc)

			err := tt.delete(t, svc)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ─── Sync ─────────────────────────────────────────────────────────────────────

func TestSecretServiceSync(t *testing.T) {
	serverID := uuid.New()
	now := time.Now().UTC()

	rawCred, err := marshalPayload(CredentialPayload{Login: "bob", Password: "secret"})
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name    string
		mock    *mockSecretsGRPC
		wantErr bool
		verify  func(t *testing.T, svc *SecretService)
	}{
		{
			name: "downloads and caches secrets from server",
			mock: &mockSecretsGRPC{
				listResp: &pb.ListSecretsResponse{
					Secrets: []*pb.Secret{
						{
							Id:               serverID.String(),
							Name:             "server-cred",
							Type:             pb.SecretType_SECRET_TYPE_CREDENTIAL,
							EncryptedPayload: rawCred,
							Version:          1,
							UpdatedAt:        timestamppb.New(now),
						},
					},
				},
			},
			verify: func(t *testing.T, svc *SecretService) {
				t.Helper()
				sec, err := svc.secretStore.GetByName(context.Background(), "server-cred", domain.SecretTypeCredential)
				require.NoError(t, err)
				assert.Equal(t, storage.SyncStatusSynced, sec.SyncStatus)
			},
		},
		{
			name:    "server error — returns error",
			mock:    &mockSecretsGRPC{listErr: errors.New("unavailable")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestSecretService(t, tt.mock)

			err := svc.Sync(context.Background(), testKey, nil)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, svc)
			}
		})
	}
}

// ─── PushPending ──────────────────────────────────────────────────────────────

func TestSecretServicePushPending(t *testing.T) {
	serverUUID := uuid.New().String()

	tests := []struct {
		name   string
		setup  func(t *testing.T, mock *mockSecretsGRPC, svc *SecretService)
		verify func(t *testing.T, svc *SecretService)
	}{
		{
			name:  "no pending — no-op",
			setup: func(t *testing.T, mock *mockSecretsGRPC, svc *SecretService) {},
			verify: func(t *testing.T, svc *SecretService) {
				t.Helper()
				has, err := svc.HasPending(context.Background())
				require.NoError(t, err)
				assert.False(t, has)
			},
		},
		{
			name: "pushes pending create to server → synced",
			setup: func(t *testing.T, mock *mockSecretsGRPC, svc *SecretService) {
				t.Helper()
				// добавляем в offline — записывается как pending.
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
				// "восстанавливаем" сеть.
				mock.createErr = nil
				mock.createResp = &pb.CreateSecretResponse{Id: serverUUID, Version: 1, CreatedAt: nowProto()}
			},
			verify: func(t *testing.T, svc *SecretService) {
				t.Helper()
				sec, err := svc.secretStore.GetByName(context.Background(), "gh", domain.SecretTypeCredential)
				require.NoError(t, err)
				assert.Equal(t, storage.SyncStatusSynced, sec.SyncStatus)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := offlineMock()
			svc := newTestSecretService(t, mock)
			tt.setup(t, mock, svc)

			err := svc.PushPending(context.Background())
			require.NoError(t, err)
			tt.verify(t, svc)
		})
	}
}

// ─── typeToProto / typeToDomain ───────────────────────────────────────────────

func TestTypeToProto(t *testing.T) {
	tests := []struct {
		name      string
		input     domain.SecretType
		wantProto pb.SecretType
	}{
		{"credential", domain.SecretTypeCredential, pb.SecretType_SECRET_TYPE_CREDENTIAL},
		{"card", domain.SecretTypeCard, pb.SecretType_SECRET_TYPE_CARD},
		{"text", domain.SecretTypeText, pb.SecretType_SECRET_TYPE_TEXT},
		{"binary", domain.SecretTypeBinary, pb.SecretType_SECRET_TYPE_BINARY},
		{"unknown → unspecified", "unknown", pb.SecretType_SECRET_TYPE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantProto, typeToProto(tt.input))
		})
	}
}

func TestTypeToDomain(t *testing.T) {
	tests := []struct {
		name       string
		input      pb.SecretType
		wantDomain domain.SecretType
	}{
		{"credential", pb.SecretType_SECRET_TYPE_CREDENTIAL, domain.SecretTypeCredential},
		{"card", pb.SecretType_SECRET_TYPE_CARD, domain.SecretTypeCard},
		{"text", pb.SecretType_SECRET_TYPE_TEXT, domain.SecretTypeText},
		{"binary", pb.SecretType_SECRET_TYPE_BINARY, domain.SecretTypeBinary},
		{"unspecified → empty string", pb.SecretType_SECRET_TYPE_UNSPECIFIED, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantDomain, typeToDomain(tt.input))
		})
	}
}
