package secret

import (
	"github.com/avc-dev/gophkeeper/internal/client/storage"
	pb "github.com/avc-dev/gophkeeper/proto"
)

type Service struct {
	client      pb.SecretsServiceClient
	secretStore *storage.SecretStorage
}

func New(client pb.SecretsServiceClient, secretStore *storage.SecretStorage) *Service {
	return &Service{client: client, secretStore: secretStore}
}
