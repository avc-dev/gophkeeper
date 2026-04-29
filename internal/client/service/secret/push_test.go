package secret

import (
	"context"
	"errors"
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
			name: "deleted without ServerID (never synced) → purged locally",
			setup: func(t *testing.T, mock *mockSecretsGRPC, svc *Service) {
				t.Helper()
				// secret never reached the server (offline), so has no ServerID
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "local-only", "u", "p", "", ""))
				require.NoError(t, svc.Delete(context.Background(), "local-only", domain.SecretTypeCredential))
			},
			verify: func(t *testing.T, svc *Service) {
				t.Helper()
				list, err := svc.List(context.Background(), domain.SecretTypeCredential)
				require.NoError(t, err)
				assert.Empty(t, list)
			},
		},
		{
			name: "update synced secret → MarkSynced with new version",
			setup: func(t *testing.T, mock *mockSecretsGRPC, svc *Service) {
				t.Helper()
				// create locally and mark it synced (simulates a prior successful push)
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh3", "alice", "oldpass", "", ""))
				sec, err := svc.secretStore.GetByName(context.Background(), "gh3", domain.SecretTypeCredential)
				require.NoError(t, err)
				require.NoError(t, svc.secretStore.MarkSynced(context.Background(), sec.ID, uuid.MustParse(serverUUID), 1, nowProto().AsTime()))
				// local update — marks it pending again with the same ServerID
				_, err = svc.secretStore.Update(context.Background(), sec.ID, []byte("new-payload"), "")
				require.NoError(t, err)
				// server accepts the update
				mock.updateResp = &pb.UpdateSecretResponse{Version: 2, UpdatedAt: nowProto()}
				mock.updateErr = nil
			},
			verify: func(t *testing.T, svc *Service) {
				t.Helper()
				sec, err := svc.secretStore.GetByName(context.Background(), "gh3", domain.SecretTypeCredential)
				require.NoError(t, err)
				assert.Equal(t, storage.SyncStatusSynced, sec.SyncStatus)
				assert.Equal(t, int64(2), sec.ServerVersion)
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

func TestPushPending_ListPendingError(t *testing.T) {
	mock := offlineMock()
	svc, db := newTestServiceWithDB(t, mock)
	db.Close() // force all storage calls to fail

	err := svc.PushPending(context.Background())
	require.Error(t, err)
}

func TestPushPending_UpdateError(t *testing.T) {
	serverUUID := uuid.New().String()
	mock := offlineMock()
	svc := newTestService(t, mock)

	// create + sync so secret has a ServerID
	mock.createErr = nil
	mock.createResp = &pb.CreateSecretResponse{Id: serverUUID, Version: 1, CreatedAt: nowProto()}
	require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
	require.NoError(t, svc.PushPending(context.Background()))

	// local update → marks pending again
	sec, err := svc.secretStore.GetByName(context.Background(), "gh", domain.SecretTypeCredential)
	require.NoError(t, err)
	_, err = svc.secretStore.Update(context.Background(), sec.ID, []byte("updated"), "")
	require.NoError(t, err)

	// server rejects update — PushPending skips it silently
	mock.updateErr = errors.New("conflict")
	require.NoError(t, svc.PushPending(context.Background()))

	// secret remains pending
	sec, err = svc.secretStore.GetByName(context.Background(), "gh", domain.SecretTypeCredential)
	require.NoError(t, err)
	assert.Equal(t, storage.SyncStatusPending, sec.SyncStatus)
}
