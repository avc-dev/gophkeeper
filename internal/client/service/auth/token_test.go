package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthServiceToken(t *testing.T) {
	tests := []struct {
		name       string
		loginFirst bool
		loginToken string
		wantErr    error
		wantToken  string
	}{
		{
			name:    "ErrNotLoggedIn before login",
			wantErr: ErrNotLoggedIn,
		},
		{
			name:       "returns stored token after login",
			loginFirst: true,
			loginToken: "my.jwt.token",
			wantToken:  "my.jwt.token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, &mockAuthGRPC{loginResp: validLoginResp(tt.loginToken)})
			if tt.loginFirst {
				_, err := svc.Login(context.Background(), "user@example.com", "pass")
				require.NoError(t, err)
			}

			tok, err := svc.Token(context.Background())
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantToken, tok)
		})
	}
}

func TestAuthServiceDeriveMasterKey(t *testing.T) {
	tests := []struct {
		name       string
		loginFirst bool
		wantErr    error
		wantKeyLen int
	}{
		{
			name:    "ErrNotLoggedIn before login",
			wantErr: ErrNotLoggedIn,
		},
		{
			name:       "returns deterministic 32-byte key after login",
			loginFirst: true,
			wantKeyLen: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, &mockAuthGRPC{loginResp: validLoginResp("tok")})
			if tt.loginFirst {
				_, err := svc.Login(context.Background(), "user@example.com", "pass")
				require.NoError(t, err)
			}

			key, err := svc.DeriveMasterKey(context.Background(), "pass")
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, key, tt.wantKeyLen)

			// деривация детерминирована
			key2, err := svc.DeriveMasterKey(context.Background(), "pass")
			require.NoError(t, err)
			assert.Equal(t, key, key2)
		})
	}
}
