package auth

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/crypto"
	pb "github.com/avc-dev/gophkeeper/proto"
)

// Login выполняет вход через gRPC, сохраняет JWT и kdf_salt локально,
// возвращает master key в памяти (никогда не сохраняется на диск).
func (s *Service) Login(ctx context.Context, email, password string) (masterKey []byte, err error) {
	resp, err := s.client.Login(ctx, &pb.LoginRequest{Email: email, Password: password})
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	if err := s.authStore.Set(ctx, keyJWT, resp.Token); err != nil {
		return nil, fmt.Errorf("store jwt: %w", err)
	}
	salt64 := base64.StdEncoding.EncodeToString(resp.KdfSalt)
	if err := s.authStore.Set(ctx, keyKDFSalt, salt64); err != nil {
		return nil, fmt.Errorf("store kdf salt: %w", err)
	}

	return crypto.DeriveKey(password, resp.KdfSalt), nil
}
