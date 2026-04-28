package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/avc-dev/gophkeeper/internal/client/storage"
	"github.com/avc-dev/gophkeeper/internal/crypto"
	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SecretService выполняет CRUD операции над секретами:
// шифрует данные на клиенте, кеширует в SQLite, синхронизирует с сервером.
type SecretService struct {
	client      pb.SecretsServiceClient
	secretStore *storage.SecretStorage
}

// NewSecretService создаёт SecretService.
func NewSecretService(client pb.SecretsServiceClient, secretStore *storage.SecretStorage) *SecretService {
	return &SecretService{client: client, secretStore: secretStore}
}

// AddCredential шифрует и сохраняет логин/пароль.
func (s *SecretService) AddCredential(ctx context.Context, masterKey []byte, name, login, password, url, note string) error {
	p := CredentialPayload{Login: login, Password: password, URL: url, Note: note}
	return s.add(ctx, masterKey, domain.SecretTypeCredential, name, p, "")
}

// AddCard шифрует и сохраняет данные банковской карты.
func (s *SecretService) AddCard(ctx context.Context, masterKey []byte, name, number, holder, expiry, cvv, bank, note string) error {
	p := CardPayload{Number: number, Holder: holder, Expiry: expiry, CVV: cvv, Bank: bank, Note: note}
	return s.add(ctx, masterKey, domain.SecretTypeCard, name, p, "")
}

// AddText шифрует и сохраняет произвольный текст.
func (s *SecretService) AddText(ctx context.Context, masterKey []byte, name, content, note string) error {
	p := TextPayload{Content: content, Note: note}
	return s.add(ctx, masterKey, domain.SecretTypeText, name, p, "")
}

// AddBinary шифрует и сохраняет бинарный файл.
func (s *SecretService) AddBinary(ctx context.Context, masterKey []byte, name, filename string, data []byte, note string) error {
	p := BinaryPayload{
		Filename: filename,
		Data:     base64.StdEncoding.EncodeToString(data),
		Note:     note,
	}
	return s.add(ctx, masterKey, domain.SecretTypeBinary, name, p, "")
}

// add — общая реализация сохранения секрета: шифрует, кеширует, отправляет на сервер.
func (s *SecretService) add(ctx context.Context, masterKey []byte, typ domain.SecretType, name string, payload any, metadata string) error {
	raw, err := marshalPayload(payload)
	if err != nil {
		return fmt.Errorf("add secret: %w", err)
	}
	// name передаётся как AAD: сервер не может подменить payload одного секрета другим —
	// при расшифровке с другим именем тег аутентификации GCM не пройдёт проверку.
	encrypted, err := crypto.Encrypt(masterKey, raw, []byte(name))
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	local := &storage.LocalSecret{
		Secret: domain.Secret{
			Type:    typ,
			Name:    name,
			Payload: encrypted,
		},
	}
	saved, err := s.secretStore.Create(ctx, local)
	if err != nil {
		return fmt.Errorf("save locally: %w", err)
	}

	// пытаемся отправить на сервер; при ошибке остаётся pending для следующей синхронизации.
	resp, err := s.client.CreateSecret(ctx, &pb.CreateSecretRequest{
		Type:             typeToProto(typ),
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

// GetCredential расшифровывает и возвращает данные логин/пароля.
func (s *SecretService) GetCredential(ctx context.Context, masterKey []byte, name string) (*CredentialPayload, error) {
	raw, err := s.decrypt(ctx, masterKey, name, domain.SecretTypeCredential)
	if err != nil {
		return nil, fmt.Errorf("get credential %q: %w", name, err)
	}
	return unmarshalCredential(raw)
}

// GetCard расшифровывает и возвращает данные карты.
func (s *SecretService) GetCard(ctx context.Context, masterKey []byte, name string) (*CardPayload, error) {
	raw, err := s.decrypt(ctx, masterKey, name, domain.SecretTypeCard)
	if err != nil {
		return nil, fmt.Errorf("get card %q: %w", name, err)
	}
	return unmarshalCard(raw)
}

// GetText расшифровывает и возвращает текст.
func (s *SecretService) GetText(ctx context.Context, masterKey []byte, name string) (*TextPayload, error) {
	raw, err := s.decrypt(ctx, masterKey, name, domain.SecretTypeText)
	if err != nil {
		return nil, fmt.Errorf("get text %q: %w", name, err)
	}
	return unmarshalText(raw)
}

// GetBinary расшифровывает и возвращает бинарные данные.
func (s *SecretService) GetBinary(ctx context.Context, masterKey []byte, name string) (*BinaryPayload, error) {
	raw, err := s.decrypt(ctx, masterKey, name, domain.SecretTypeBinary)
	if err != nil {
		return nil, fmt.Errorf("get binary %q: %w", name, err)
	}
	return unmarshalBinary(raw)
}

// decrypt читает из локального кеша и расшифровывает payload.
func (s *SecretService) decrypt(ctx context.Context, masterKey []byte, name string, typ domain.SecretType) ([]byte, error) {
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

// List возвращает все не удалённые секреты из локального кеша.
func (s *SecretService) List(ctx context.Context, typ domain.SecretType) ([]*storage.LocalSecret, error) {
	return s.secretStore.List(ctx, typ)
}

// HasPending сообщает, есть ли записи ожидающие отправки на сервер.
func (s *SecretService) HasPending(ctx context.Context) (bool, error) {
	pending, err := s.secretStore.ListPending(ctx)
	if err != nil {
		return false, fmt.Errorf("has pending: %w", err)
	}
	return len(pending) > 0, nil
}

// Delete помечает секрет удалённым локально и отправляет удаление на сервер.
func (s *SecretService) Delete(ctx context.Context, name string, typ domain.SecretType) error {
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

// Sync загружает все секреты с сервера и обновляет локальный кеш.
// Pending записи не перезаписываются (ON CONFLICT WHERE sync_status != 'pending').
func (s *SecretService) Sync(ctx context.Context, masterKey []byte, since *time.Time) error {
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
		serverID := localID // для pull: используем server id как local id (единый namespace)
		updatedAt := pbSec.UpdatedAt.AsTime()

		sec := &storage.LocalSecret{
			Secret: domain.Secret{
				ID:        localID,
				Type:      typeToDomain(pbSec.Type),
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

// PushPending отправляет на сервер все pending-записи.
func (s *SecretService) PushPending(ctx context.Context) error {
	pending, err := s.secretStore.ListPending(ctx)
	if err != nil {
		return fmt.Errorf("list pending: %w", err)
	}

	for _, sec := range pending {
		if sec.Deleted {
			if sec.ServerID == nil {
				// никогда не было на сервере — просто чистим.
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
			// новая запись — создаём на сервере.
			resp, err := s.client.CreateSecret(ctx, &pb.CreateSecretRequest{
				Type:             typeToProto(sec.Type),
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
			// обновление существующей.
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

func typeToProto(t domain.SecretType) pb.SecretType {
	switch t {
	case domain.SecretTypeCredential:
		return pb.SecretType_SECRET_TYPE_CREDENTIAL
	case domain.SecretTypeCard:
		return pb.SecretType_SECRET_TYPE_CARD
	case domain.SecretTypeText:
		return pb.SecretType_SECRET_TYPE_TEXT
	case domain.SecretTypeBinary:
		return pb.SecretType_SECRET_TYPE_BINARY
	default:
		return pb.SecretType_SECRET_TYPE_UNSPECIFIED
	}
}

func typeToDomain(t pb.SecretType) domain.SecretType {
	switch t {
	case pb.SecretType_SECRET_TYPE_CREDENTIAL:
		return domain.SecretTypeCredential
	case pb.SecretType_SECRET_TYPE_CARD:
		return domain.SecretTypeCard
	case pb.SecretType_SECRET_TYPE_TEXT:
		return domain.SecretTypeText
	case pb.SecretType_SECRET_TYPE_BINARY:
		return domain.SecretTypeBinary
	default:
		return ""
	}
}
