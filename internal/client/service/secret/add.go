package secret

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/avc-dev/gophkeeper/internal/client/storage"
	"github.com/avc-dev/gophkeeper/internal/crypto"
	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/avc-dev/gophkeeper/internal/protoconv"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/google/uuid"
)

// AddCredential шифрует и сохраняет логин/пароль локально и, при наличии соединения, на сервере.
func (s *Service) AddCredential(ctx context.Context, masterKey []byte, name, login, password, url, note string) error {
	return s.add(ctx, masterKey, domain.SecretTypeCredential, name,
		CredentialPayload{Login: login, Password: password, URL: url, Note: note}, "")
}

// AddCard шифрует и сохраняет данные банковской карты.
func (s *Service) AddCard(ctx context.Context, masterKey []byte, name, number, holder, expiry, cvv, bank, note string) error {
	return s.add(ctx, masterKey, domain.SecretTypeCard, name,
		CardPayload{Number: number, Holder: holder, Expiry: expiry, CVV: cvv, Bank: bank, Note: note}, "")
}

// AddText шифрует и сохраняет произвольный текст.
func (s *Service) AddText(ctx context.Context, masterKey []byte, name, content, note string) error {
	return s.add(ctx, masterKey, domain.SecretTypeText, name,
		TextPayload{Content: content, Note: note}, "")
}

// AddBinary шифрует и сохраняет бинарный файл (данные кодируются в base64 перед шифрованием).
func (s *Service) AddBinary(ctx context.Context, masterKey []byte, name, filename string, data []byte, note string) error {
	return s.add(ctx, masterKey, domain.SecretTypeBinary, name,
		BinaryPayload{Filename: filename, Data: base64.StdEncoding.EncodeToString(data), Note: note}, "")
}

// AddOTP шифрует и сохраняет TOTP-семя. Seed должен быть валидным base32-ключом.
func (s *Service) AddOTP(ctx context.Context, masterKey []byte, name, seed, issuer, account, note string) error {
	return s.add(ctx, masterKey, domain.SecretTypeOTP, name,
		OTPPayload{Seed: seed, Issuer: issuer, Account: account, Note: note}, "")
}

func (s *Service) add(ctx context.Context, masterKey []byte, typ domain.SecretType, name string, payload any, metadata string) error {
	raw, err := marshalPayload(payload)
	if err != nil {
		return fmt.Errorf("add secret: %w", err)
	}
	// name передаётся как AAD: при расшифровке с другим именем GCM-тег не пройдёт проверку.
	encrypted, err := crypto.Encrypt(masterKey, raw, []byte(name))
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	local := &storage.LocalSecret{
		Secret: domain.Secret{Type: typ, Name: name, Payload: encrypted},
	}
	saved, err := s.secretStore.Create(ctx, local)
	if err != nil {
		return fmt.Errorf("save locally: %w", err)
	}

	resp, err := s.client.CreateSecret(ctx, &pb.CreateSecretRequest{
		Type:             protoconv.TypeToProto(typ),
		Name:             name,
		EncryptedPayload: encrypted,
		Metadata:         metadata,
	})
	if err != nil {
		return nil // сохранено локально, отправится при следующем sync
	}

	serverID, err := uuid.Parse(resp.Id)
	if err != nil {
		return nil
	}
	return s.secretStore.MarkSynced(ctx, saved.ID, serverID, resp.Version, resp.CreatedAt.AsTime())
}
