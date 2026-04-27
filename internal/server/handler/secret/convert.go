package secret

import (
	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// secretToProto конвертирует доменный Secret в protobuf-сообщение.
func secretToProto(s *domain.Secret) *pb.Secret {
	return &pb.Secret{
		Id:               s.ID.String(),
		Name:             s.Name,
		Type:             typeToProto(s.Type),
		EncryptedPayload: s.Payload,
		Metadata:         s.Metadata,
		Version:          s.Version,
		CreatedAt:        timestamppb.New(s.CreatedAt),
		UpdatedAt:        timestamppb.New(s.UpdatedAt),
	}
}

// typeToProto конвертирует тип секрета из domain в proto enum.
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

// typeToDomain конвертирует тип секрета из proto enum в domain.
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
