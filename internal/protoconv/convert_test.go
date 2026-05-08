package protoconv

import (
	"testing"

	"github.com/avc-dev/gophkeeper/internal/domain"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/stretchr/testify/assert"
)

func TestTypeToProto(t *testing.T) {
	tests := []struct {
		in   domain.SecretType
		want pb.SecretType
	}{
		{domain.SecretTypeCredential, pb.SecretType_SECRET_TYPE_CREDENTIAL},
		{domain.SecretTypeCard, pb.SecretType_SECRET_TYPE_CARD},
		{domain.SecretTypeText, pb.SecretType_SECRET_TYPE_TEXT},
		{domain.SecretTypeBinary, pb.SecretType_SECRET_TYPE_BINARY},
		{"unknown", pb.SecretType_SECRET_TYPE_UNSPECIFIED},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, TypeToProto(tt.in), "input: %q", tt.in)
	}
}

func TestTypeToDomain(t *testing.T) {
	tests := []struct {
		in   pb.SecretType
		want domain.SecretType
	}{
		{pb.SecretType_SECRET_TYPE_CREDENTIAL, domain.SecretTypeCredential},
		{pb.SecretType_SECRET_TYPE_CARD, domain.SecretTypeCard},
		{pb.SecretType_SECRET_TYPE_TEXT, domain.SecretTypeText},
		{pb.SecretType_SECRET_TYPE_BINARY, domain.SecretTypeBinary},
		{pb.SecretType_SECRET_TYPE_UNSPECIFIED, ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, TypeToDomain(tt.in), "input: %v", tt.in)
	}
}
