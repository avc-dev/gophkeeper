package auth

import (
	"errors"

	"github.com/avc-dev/gophkeeper/internal/client/storage"
	pb "github.com/avc-dev/gophkeeper/proto"
)

const (
	keyJWT        = "jwt_token"
	keyKDFSalt    = "kdf_salt"
	keyLastSyncAt = "last_sync_at"
)

var ErrNotLoggedIn = errors.New("not logged in: run 'gophkeeper login' first")

type Service struct {
	client    pb.AuthServiceClient
	authStore *storage.AuthStorage
}

func New(client pb.AuthServiceClient, authStore *storage.AuthStorage) *Service {
	return &Service{client: client, authStore: authStore}
}
