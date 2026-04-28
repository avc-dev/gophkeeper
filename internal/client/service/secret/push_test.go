package secret

import (
	"context"
	"testing"

	"github.com/avc-dev/gophkeeper/internal/client/storage"
	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretServicePushPending(t *testing.T) {
	serverUUID := uuid.New().String()

	tests := []struct {
		name   string
		setup  func(t *testing.T, mock *mockSecretsGRPC, svc *Service)
		verify func(t *testing.T, svc *Service)
	}{
		{
			name:  "no pending — no-op",
			setup: func(t *testing.T, mock *mockSecretsGRPC, svc *Service) {},
			verify: func(t *testing.T, svc *Service) {
				t.Helper()
				has, err := svc.HasPending(context.Background())
				require.NoError(t, err)
				assert.False(t, has)
			},
		},
		{
			name: "pushes pending create to server → synced",
			setup: func(t *testing.T, mock *mockSecretsGRPC, svc *Service) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
				mock.createErr = nil
				mock.createResp = &pb.CreateSecretResponse{Id: serverUUID, Version: 1, CreatedAt: nowProto()}
			},
			verify: func(t *testing.T, svc *Service) {
				t.Helper()
				sec, err := svc.secretStore.GetByName(context.Background(), "gh", domain.SecretTypeCredential)
				require.NoError(t, err)
				assert.Equal(t, storage.SyncStatusSynced, sec.SyncStatus)
			},
		},
		{
			name: "pushes pending delete (synced) to server → purged",
			setup: func(t *testing.T, mock *mockSecretsGRPC, svc *Service) {
				t.Helper()
				// add + sync so it has a ServerID
				mock.createResp = &pb.CreateSecretResponse{Id: serverUUID, Version: 1, CreatedAt: nowProto()}
				mock.createErr = nil
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh2", "alice", "pass", "", ""))
				// now soft-delete it (server delete will also succeed)
				mock.deleteResp = &pb.DeleteSecretResponse{}
				require.NoError(t, svc.Delete(context.Background(), "gh2", domain.SecretTypeCredential))
				// reset for push
				mock.createErr = nil
			},
			verify: func(t *testing.T, svc *Service) {
				t.Helper()
				list, err := svc.List(context.Background(), domain.SecretTypeCredential)
				require.NoError(t, err)
				assert.Empty(t, list)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := offlineMock()
			svc := newTestService(t, mock)
			tt.setup(t, mock, svc)

			require.NoError(t, svc.PushPending(context.Background()))
			tt.verify(t, svc)
		})
	}
}
