package protoconv

import (
	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
)

// TypeToProto converts a domain secret type to the protobuf enum.
func TypeToProto(t domain.SecretType) pb.SecretType {
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

// TypeToDomain converts a protobuf secret type enum to the domain type.
func TypeToDomain(t pb.SecretType) domain.SecretType {
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
