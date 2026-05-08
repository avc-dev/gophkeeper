package secret

import (
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
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
		verify  func(t *testing.T, svc *Service)
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
			verify: func(t *testing.T, svc *Service) {
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
			svc := newTestService(t, tt.mock)

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
