package secret

import (
	"context"
	"fmt"
	"time"

	"github.com/avc-dev/gophkeeper/internal/client/storage"
	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/avc-dev/gophkeeper/internal/protoconv"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) Sync(ctx context.Context, masterKey []byte, since *time.Time) error {
	var sinceProto *timestamppb.Timestamp
	if since != nil {
		sinceProto = timestamppb.New(*since)
	}

	resp, err := s.client.ListSecrets(ctx, &pb.ListSecretsRequest{Since: sinceProto})
	if err != nil {
		return fmt.Errorf("list from server: %w", err)
	}

	for _, pbSec := range resp.Secrets {
		localID, err := uuid.Parse(pbSec.Id)
		if err != nil {
			continue
		}
		serverID := localID
		updatedAt := pbSec.UpdatedAt.AsTime()

		sec := &storage.LocalSecret{
			Secret: domain.Secret{
				ID:        localID,
				Type:      protoconv.TypeToDomain(pbSec.Type),
				Name:      pbSec.Name,
				Payload:   pbSec.EncryptedPayload,
				Metadata:  pbSec.Metadata,
				Version:   pbSec.Version,
				CreatedAt: pbSec.CreatedAt.AsTime(),
				UpdatedAt: updatedAt,
			},
			ServerID:        &serverID,
			LocalVersion:    pbSec.Version,
			ServerVersion:   pbSec.Version,
			ServerUpdatedAt: &updatedAt,
			SyncStatus:      storage.SyncStatusSynced,
		}
		_ = s.secretStore.Upsert(ctx, sec)
	}
	return nil
}
