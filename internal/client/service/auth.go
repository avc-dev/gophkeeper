package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/avc-dev/gophkeeper/internal/client/storage"
	"github.com/avc-dev/gophkeeper/internal/crypto"
	pb "github.com/avc-dev/gophkeeper/proto"
)

const (
	keyJWT     = "jwt_token"
	keyKDFSalt = "kdf_salt"
)

// ErrNotLoggedIn — пользователь не выполнил вход.
var ErrNotLoggedIn = errors.New("not logged in: run 'gophkeeper login' first")

// AuthService выполняет регистрацию и вход через gRPC, хранит JWT + kdf_salt локально.
type AuthService struct {
	client    pb.AuthServiceClient
	authStore *storage.AuthStorage
}

// NewAuthService создаёт AuthService.
func NewAuthService(client pb.AuthServiceClient, authStore *storage.AuthStorage) *AuthService {
	return &AuthService{client: client, authStore: authStore}
}

// Register регистрирует нового пользователя на сервере.
func (s *AuthService) Register(ctx context.Context, email, password string) error {
	_, err := s.client.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}
	return nil
}

// Login выполняет вход, сохраняет JWT и kdf_salt, возвращает master key (только в памяти).
func (s *AuthService) Login(ctx context.Context, email, password string) (masterKey []byte, err error) {
	resp, err := s.client.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: password,
	})
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

	key := crypto.DeriveKey(password, resp.KdfSalt)
	return key, nil
}

// DeriveMasterKey читает kdf_salt из локального хранилища и деривирует master key.
// Используется командами, требующими шифрования, без обращения к серверу.
func (s *AuthService) DeriveMasterKey(ctx context.Context, password string) ([]byte, error) {
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

// Token возвращает сохранённый JWT или ErrNotLoggedIn.
func (s *AuthService) Token(ctx context.Context) (string, error) {
	token, err := s.authStore.Get(ctx, keyJWT)
	if err != nil {
		return "", fmt.Errorf("read token: %w", err)
	}
	if token == "" {
		return "", ErrNotLoggedIn
	}
	return token, nil
}

// GetLastSyncAt возвращает время последней успешной синхронизации или nil.
func (s *AuthService) GetLastSyncAt(ctx context.Context) (*time.Time, error) {
	v, err := s.authStore.Get(ctx, keyLastSyncAt)
	if err != nil {
		return nil, fmt.Errorf("read last_sync_at: %w", err)
	}
	if v == "" {
		// nilnil в данном случае допустим,
		// поскольку предыдущей синхронизации может не быть,
		// и это не является ошибкой
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return nil, fmt.Errorf("parse last_sync_at: %w", err)
	}
	return &t, nil
}

// SetLastSyncAt сохраняет время последней синхронизации.
func (s *AuthService) SetLastSyncAt(ctx context.Context, t time.Time) error {
	if err := s.authStore.Set(ctx, keyLastSyncAt, t.UTC().Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("set last_sync_at: %w", err)
	}
	return nil
}

// Logout удаляет локальные учётные данные.
func (s *AuthService) Logout(ctx context.Context) error {
	if err := s.authStore.Delete(ctx, keyJWT); err != nil {
		return fmt.Errorf("delete jwt: %w", err)
	}
	if err := s.authStore.Delete(ctx, keyKDFSalt); err != nil {
		return fmt.Errorf("delete kdf salt: %w", err)
	}
	return nil
}
