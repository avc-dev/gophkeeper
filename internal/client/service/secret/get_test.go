package secret

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretServiceGet(t *testing.T) {
	wrongKey := bytes.Repeat([]byte{0xDE}, 32)

	tests := []struct {
		name    string
		setup   func(t *testing.T, svc *Service)
		get     func(svc *Service) error
		wantErr bool
	}{
		{
			name: "credential — decrypts correctly",
			setup: func(t *testing.T, svc *Service) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "s3cr3t", "https://example.com", ""))
			},
			get: func(svc *Service) error {
				got, err := svc.GetCredential(context.Background(), testKey, "gh")
				if err != nil {
					return err
				}
				assert.Equal(t, "alice", got.Login)
				assert.Equal(t, "s3cr3t", got.Password)
				return nil
			},
		},
		{
			name: "credential — wrong master key",
			setup: func(t *testing.T, svc *Service) {
				t.Helper()
				require.NoError(t, svc.AddCredential(context.Background(), testKey, "gh", "alice", "pass", "", ""))
			},
			get: func(svc *Service) error {
				_, err := svc.GetCredential(context.Background(), wrongKey, "gh")
				return err
			},
			wantErr: true,
		},
		{
			name:  "credential — not found",
			setup: func(t *testing.T, svc *Service) {},
			get: func(svc *Service) error {
				_, err := svc.GetCredential(context.Background(), testKey, "nonexistent")
				return err
			},
			wantErr: true,
		},
		{
			name: "card — decrypts correctly",
			setup: func(t *testing.T, svc *Service) {
				t.Helper()
				require.NoError(t, svc.AddCard(context.Background(), testKey, "visa", "4532015112830366", "JOHN DOE", "12/26", "123", "Tinkoff", ""))
			},
			get: func(svc *Service) error {
				got, err := svc.GetCard(context.Background(), testKey, "visa")
				if err != nil {
					return err
				}
				assert.Equal(t, "4532015112830366", got.Number)
				assert.Equal(t, "123", got.CVV)
				return nil
			},
		},
		{
			name: "text — decrypts correctly",
			setup: func(t *testing.T, svc *Service) {
				t.Helper()
				require.NoError(t, svc.AddText(context.Background(), testKey, "note", "top secret text", ""))
			},
			get: func(svc *Service) error {
				got, err := svc.GetText(context.Background(), testKey, "note")
				if err != nil {
					return err
				}
				assert.Equal(t, "top secret text", got.Content)
				return nil
			},
		},
		{
			name: "binary — decrypts correctly",
			setup: func(t *testing.T, svc *Service) {
				t.Helper()
				require.NoError(t, svc.AddBinary(context.Background(), testKey, "file", "data.bin", []byte{0xDE, 0xAD, 0xBE, 0xEF}, ""))
			},
			get: func(svc *Service) error {
				got, err := svc.GetBinary(context.Background(), testKey, "file")
				if err != nil {
					return err
				}
				assert.Equal(t, "data.bin", got.Filename)
				assert.Equal(t, base64.StdEncoding.EncodeToString([]byte{0xDE, 0xAD, 0xBE, 0xEF}), got.Data)
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, offlineMock())
			tt.setup(t, svc)

			err := tt.get(svc)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
