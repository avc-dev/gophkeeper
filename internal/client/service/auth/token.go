package auth

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/crypto"
)

// Token возвращает сохранённый JWT или ErrNotLoggedIn.
func (s *Service) Token(ctx context.Context) (string, error) {
	token, err := s.authStore.Get(ctx, keyJWT)
	if err != nil {
		return "", fmt.Errorf("read token: %w", err)
	}
	if token == "" {
		return "", ErrNotLoggedIn
	}
	return token, nil
}

// DeriveMasterKey читает kdf_salt из локального хранилища и деривирует master key.
// Используется командами, требующими шифрования, без обращения к серверу.
func (s *Service) DeriveMasterKey(ctx context.Context, password string) ([]byte, error) {
	salt64, err := s.authStore.Get(ctx, keyKDFSalt)
	if err != nil {
		return nil, fmt.Errorf("read kdf salt: %w", err)
	}
	if salt64 == "" {
		return nil, ErrNotLoggedIn
	}
	salt, err := base64.StdEncoding.DecodeString(salt64)
	if err != nil {
		return nil, fmt.Errorf("decode kdf salt: %w", err)
	}
	return crypto.DeriveKey(password, salt), nil
}
