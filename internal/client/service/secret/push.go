package secret

import (
	"context"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/protoconv"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/google/uuid"
)

func (s *Service) PushPending(ctx context.Context) error {
	pending, err := s.secretStore.ListPending(ctx)
	if err != nil {
		return fmt.Errorf("list pending: %w", err)
	}

	for _, sec := range pending {
		if sec.Deleted {
			if sec.ServerID == nil {
				_ = s.secretStore.Purge(ctx, sec.ID)
				continue
			}
			if _, err := s.client.DeleteSecret(ctx, &pb.DeleteSecretRequest{
				Id: sec.ServerID.String(),
			}); err == nil {
				_ = s.secretStore.Purge(ctx, sec.ID)
			}
			continue
		}

		if sec.ServerID == nil {
			resp, err := s.client.CreateSecret(ctx, &pb.CreateSecretRequest{
				Type:             protoconv.TypeToProto(sec.Type),
				Name:             sec.Name,
				EncryptedPayload: sec.Payload,
				Metadata:         sec.Metadata,
			})
			if err != nil {
				continue
			}
			serverID, _ := uuid.Parse(resp.Id)
			_ = s.secretStore.MarkSynced(ctx, sec.ID, serverID, resp.Version, resp.CreatedAt.AsTime())
		} else {
			resp, err := s.client.UpdateSecret(ctx, &pb.UpdateSecretRequest{
				Id:               sec.ServerID.String(),
				EncryptedPayload: sec.Payload,
				Metadata:         sec.Metadata,
				ExpectedVersion:  sec.ServerVersion,
			})
			if err != nil {
				continue
			}
			_ = s.secretStore.MarkSynced(ctx, sec.ID, *sec.ServerID, resp.Version, resp.UpdatedAt.AsTime())
		}
	}
	return nil
}
