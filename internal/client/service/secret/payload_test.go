package secret

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPayloadRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		verify  func(t *testing.T, data []byte)
	}{
		{
			name:    "credential — full fields",
			payload: CredentialPayload{Login: "alice", Password: "s3cr3t", URL: "https://example.com", Note: "work"},
			verify: func(t *testing.T, data []byte) {
				t.Helper()
				got, err := unmarshalCredential(data)
				require.NoError(t, err)
				assert.Equal(t, CredentialPayload{Login: "alice", Password: "s3cr3t", URL: "https://example.com", Note: "work"}, *got)
			},
		},
		{
			name:    "credential — optional fields omitted",
			payload: CredentialPayload{Login: "bob", Password: "pass"},
			verify: func(t *testing.T, data []byte) {
				t.Helper()
				got, err := unmarshalCredential(data)
				require.NoError(t, err)
				assert.Equal(t, "bob", got.Login)
				assert.Empty(t, got.URL)
				assert.Empty(t, got.Note)
			},
		},
		{
			name:    "card",
			payload: CardPayload{Number: "4532015112830366", Holder: "JOHN DOE", Expiry: "12/26", CVV: "123", Bank: "Tinkoff", Note: "main"},
			verify: func(t *testing.T, data []byte) {
				t.Helper()
				got, err := unmarshalCard(data)
				require.NoError(t, err)
				assert.Equal(t, CardPayload{Number: "4532015112830366", Holder: "JOHN DOE", Expiry: "12/26", CVV: "123", Bank: "Tinkoff", Note: "main"}, *got)
			},
		},
		{
			name:    "text — with newlines",
			payload: TextPayload{Content: "line1\nline2", Note: "important"},
			verify: func(t *testing.T, data []byte) {
				t.Helper()
				got, err := unmarshalText(data)
				require.NoError(t, err)
				assert.Equal(t, "line1\nline2", got.Content)
			},
		},
		{
			name:    "binary",
			payload: BinaryPayload{Filename: "key.pem", Data: "AAH//g==", Note: "ssh key"},
			verify: func(t *testing.T, data []byte) {
				t.Helper()
				got, err := unmarshalBinary(data)
				require.NoError(t, err)
				assert.Equal(t, "key.pem", got.Filename)
				assert.Equal(t, "AAH//g==", got.Data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := marshalPayload(tt.payload)
			require.NoError(t, err)
			tt.verify(t, data)
		})
	}
}

func TestPayloadUnmarshalInvalidJSON(t *testing.T) {
	bad := []byte("not valid json {{{")

	tests := []struct {
		name      string
		unmarshal func([]byte) error
	}{
		{"credential", func(d []byte) error { _, err := unmarshalCredential(d); return err }},
		{"card", func(d []byte) error { _, err := unmarshalCard(d); return err }},
		{"text", func(d []byte) error { _, err := unmarshalText(d); return err }},
		{"binary", func(d []byte) error { _, err := unmarshalBinary(d); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.unmarshal(bad)
			require.Error(t, err)
		})
	}
}
