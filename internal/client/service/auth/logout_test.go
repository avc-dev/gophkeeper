package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthServiceLogout(t *testing.T) {
	// одна сессия: сначала логин, потом последовательные проверки.
	svc := newTestService(t, &mockAuthGRPC{loginResp: validLoginResp("tok")})
	_, err := svc.Login(context.Background(), "user@example.com", "pass")
	require.NoError(t, err)

	t.Run("token available before logout", func(t *testing.T) {
		_, err := svc.Token(context.Background())
		require.NoError(t, err)
	})

	t.Run("token not available after logout", func(t *testing.T) {
		require.NoError(t, svc.Logout(context.Background()))
		_, err := svc.Token(context.Background())
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNotLoggedIn)
	})

	t.Run("kdf_salt not available after logout", func(t *testing.T) {
		_, err := svc.DeriveMasterKey(context.Background(), "pass")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNotLoggedIn)
	})
}
