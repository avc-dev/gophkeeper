package secret

import (
	"testing"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/avc-dev/gophkeeper/internal/protoconv"
	pb "github.com/avc-dev/gophkeeper/proto"
	"github.com/stretchr/testify/assert"
)

func TestTypeToProto(t *testing.T) {
	tests := []struct {
		name      string
		input     domain.SecretType
		wantProto pb.SecretType
	}{
		{"credential", domain.SecretTypeCredential, pb.SecretType_SECRET_TYPE_CREDENTIAL},
		{"card", domain.SecretTypeCard, pb.SecretType_SECRET_TYPE_CARD},
		{"text", domain.SecretTypeText, pb.SecretType_SECRET_TYPE_TEXT},
		{"binary", domain.SecretTypeBinary, pb.SecretType_SECRET_TYPE_BINARY},
		{"otp", domain.SecretTypeOTP, pb.SecretType_SECRET_TYPE_OTP},
		{"unknown → unspecified", "unknown", pb.SecretType_SECRET_TYPE_UNSPECIFIED},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantProto, protoconv.TypeToProto(tt.input))
		})
	}
}

func TestTypeToDomain(t *testing.T) {
	tests := []struct {
		name       string
		input      pb.SecretType
		wantDomain domain.SecretType
	}{
		{"credential", pb.SecretType_SECRET_TYPE_CREDENTIAL, domain.SecretTypeCredential},
		{"card", pb.SecretType_SECRET_TYPE_CARD, domain.SecretTypeCard},
		{"text", pb.SecretType_SECRET_TYPE_TEXT, domain.SecretTypeText},
		{"binary", pb.SecretType_SECRET_TYPE_BINARY, domain.SecretTypeBinary},
		{"otp", pb.SecretType_SECRET_TYPE_OTP, domain.SecretTypeOTP},
		{"unspecified → empty", pb.SecretType_SECRET_TYPE_UNSPECIFIED, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantDomain, protoconv.TypeToDomain(tt.input))
		})
	}
}
