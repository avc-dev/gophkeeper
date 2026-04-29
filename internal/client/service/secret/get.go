package secret

import (
	"context"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/crypto"
	"github.com/avc-dev/gophkeeper/internal/domain"
)

func (s *Service) GetCredential(ctx context.Context, masterKey []byte, name string) (*CredentialPayload, error) {
	raw, err := s.decrypt(ctx, masterKey, name, domain.SecretTypeCredential)
	if err != nil {
		return nil, fmt.Errorf("get credential %q: %w", name, err)
	}
	return unmarshalCredential(raw)
}

func (s *Service) GetCard(ctx context.Context, masterKey []byte, name string) (*CardPayload, error) {
	raw, err := s.decrypt(ctx, masterKey, name, domain.SecretTypeCard)
	if err != nil {
		return nil, fmt.Errorf("get card %q: %w", name, err)
	}
	return unmarshalCard(raw)
}

func (s *Service) GetText(ctx context.Context, masterKey []byte, name string) (*TextPayload, error) {
	raw, err := s.decrypt(ctx, masterKey, name, domain.SecretTypeText)
	if err != nil {
		return nil, fmt.Errorf("get text %q: %w", name, err)
	}
	return unmarshalText(raw)
}

func (s *Service) GetBinary(ctx context.Context, masterKey []byte, name string) (*BinaryPayload, error) {
	raw, err := s.decrypt(ctx, masterKey, name, domain.SecretTypeBinary)
	if err != nil {
		return nil, fmt.Errorf("get binary %q: %w", name, err)
	}
	return unmarshalBinary(raw)
}

func (s *Service) decrypt(ctx context.Context, masterKey []byte, name string, typ domain.SecretType) ([]byte, error) {
	sec, err := s.secretStore.GetByName(ctx, name, typ)
	if err != nil {
		return nil, fmt.Errorf("lookup %s %q: %w", typ, name, err)
	}
	raw, err := crypto.Decrypt(masterKey, sec.Payload, []byte(sec.Name))
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return raw, nil
}
