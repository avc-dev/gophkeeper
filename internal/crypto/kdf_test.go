package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveKey(t *testing.T) {
	salt := make([]byte, SaltSize)

	tests := []struct {
		name         string
		password     string
		salt         []byte
		wantLen      int
		wantSameAs   []byte // если задан — ключ должен совпасть
		wantDiffFrom []byte // если задан — ключ должен отличаться
	}{
		{
			name:       "детерминированность",
			password:   "my-password",
			salt:       salt,
			wantLen:    keySize,
			wantSameAs: DeriveKey("my-password", salt),
		},
		{
			name:         "разные пароли дают разные ключи",
			password:     "password-2",
			salt:         salt,
			wantLen:      keySize,
			wantDiffFrom: DeriveKey("password-1", salt),
		},
		{
			name:         "разные соли дают разные ключи",
			password:     "password",
			salt:         func() []byte { s := make([]byte, SaltSize); s[0] = 1; return s }(),
			wantLen:      keySize,
			wantDiffFrom: DeriveKey("password", salt),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := DeriveKey(tt.password, tt.salt)

			assert.Len(t, key, tt.wantLen)

			if tt.wantSameAs != nil {
				assert.Equal(t, tt.wantSameAs, key)
			}
			if tt.wantDiffFrom != nil {
				assert.NotEqual(t, tt.wantDiffFrom, key)
			}
		})
	}
}

func TestGenerateSalt(t *testing.T) {
	t.Run("корректная длина", func(t *testing.T) {
		salt, err := GenerateSalt()
		require.NoError(t, err)
		assert.Len(t, salt, SaltSize)
	})

	t.Run("уникальность", func(t *testing.T) {
		s1, err := GenerateSalt()
		require.NoError(t, err)

		s2, err := GenerateSalt()
		require.NoError(t, err)

		assert.NotEqual(t, s1, s2)
	})
}
