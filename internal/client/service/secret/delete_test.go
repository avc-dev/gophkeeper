package secret

import (
	"context"
	"testing"

	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSecretServiceDelete(t *testing.T) {
	syncedServerID := uuid.New().String()

	tests := []struct {
		name    string
		mock    *mockSecretsGRPC
		setup   func(t *testing.T, svc *Service)
		delete  func(t *testing.T, svc *Service) error
		wantErr bool
	}{
		{
			name: "never synced — local soft delete, GetByName returns not found",
			mock: offlineMock(),
			setup: func(t *testing.T, svc *Service) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
			},
			delete: func(t *testing.T, svc *Service) error {
				if err := svc.Delete(context.Background(), "gh", domain.SecretTypeCredential); err != nil {
					return err
				}
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
			setup: func(t *testing.T, svc *Service) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
			},
			delete: func(t *testing.T, svc *Service) error {
				return svc.Delete(context.Background(), "gh", domain.SecretTypeCredential)
			},
		},
		{
			name:  "not found — returns error",
			mock:  offlineMock(),
			setup: func(t *testing.T, svc *Service) {},
			delete: func(t *testing.T, svc *Service) error {
				return svc.Delete(context.Background(), "nonexistent", domain.SecretTypeCredential)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, tt.mock)
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
