package secret

import (
	"context"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
)

func (s *Service) Delete(ctx context.Context, name string, typ domain.SecretType) error {
	sec, err := s.secretStore.GetByName(ctx, name, typ)
	if err != nil {
		return fmt.Errorf("find %s %q: %w", typ, name, err)
	}
	if err := s.secretStore.Delete(ctx, sec.ID); err != nil {
		return fmt.Errorf("delete locally: %w", err)
	}
	if sec.ServerID == nil {
		return nil // никогда не было на сервере
	}
	if _, err := s.client.DeleteSecret(ctx, &pb.DeleteSecretRequest{
		Id: sec.ServerID.String(),
	}); err != nil {
		return nil // локально удалено, на сервере удалится при sync
	}
	return s.secretStore.Purge(ctx, sec.ID)
}
