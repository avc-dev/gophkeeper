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

func TestSecretServiceAdd(t *testing.T) {
	serverUUID := uuid.New().String()

	tests := []struct {
		name           string
		mock           *mockSecretsGRPC
		add            func(svc *Service) error
		secretName     string
		secretType     domain.SecretType
		wantSyncStatus storage.SyncStatus
	}{
		{
			name: "credential — server online → synced",
			mock: &mockSecretsGRPC{
				createResp: &pb.CreateSecretResponse{Id: serverUUID, Version: 1, CreatedAt: nowProto()},
			},
			add: func(svc *Service) error {
				return svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "https://github.com", "")
			},
			secretName:     "gh",
			secretType:     domain.SecretTypeCredential,
			wantSyncStatus: storage.SyncStatusSynced,
		},
		{
			name: "credential — server offline → pending",
			mock: offlineMock(),
			add: func(svc *Service) error {
				return svc.AddCredential(context.Background(), testKey, "gh-offline", "alice", "pass", "", "")
			},
			secretName:     "gh-offline",
			secretType:     domain.SecretTypeCredential,
			wantSyncStatus: storage.SyncStatusPending,
		},
		{
			name: "card — offline",
			mock: offlineMock(),
			add: func(svc *Service) error {
				return svc.AddCard(context.Background(), testKey, "visa", "4532015112830366", "JOHN DOE", "12/26", "123", "Tinkoff", "")
			},
			secretName:     "visa",
			secretType:     domain.SecretTypeCard,
			wantSyncStatus: storage.SyncStatusPending,
		},
		{
			name: "text — offline",
			mock: offlineMock(),
			add: func(svc *Service) error {
				return svc.AddText(context.Background(), testKey, "note", "my secret note", "")
			},
			secretName:     "note",
			secretType:     domain.SecretTypeText,
			wantSyncStatus: storage.SyncStatusPending,
		},
		{
			name: "binary — offline",
			mock: offlineMock(),
			add: func(svc *Service) error {
				return svc.AddBinary(context.Background(), testKey, "key", "key.pem", []byte{0x01, 0x02}, "")
			},
			secretName:     "key",
			secretType:     domain.SecretTypeBinary,
			wantSyncStatus: storage.SyncStatusPending,
		},
		{
			name: "otp — offline",
			mock: offlineMock(),
			add: func(svc *Service) error {
				return svc.AddOTP(context.Background(), testKey, "github2fa", "JBSWY3DPEHPK3PXP", "GitHub", "alice@example.com", "")
			},
			secretName:     "github2fa",
			secretType:     domain.SecretTypeOTP,
			wantSyncStatus: storage.SyncStatusPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, tt.mock)
			require.NoError(t, tt.add(svc))

			sec, err := svc.secretStore.GetByName(context.Background(), tt.secretName, tt.secretType)
			require.NoError(t, err)
			assert.Equal(t, tt.wantSyncStatus, sec.SyncStatus)
		})
	}
}
