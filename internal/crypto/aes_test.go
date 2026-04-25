package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeKey() []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return key
}

func TestEncryptDecrypt(t *testing.T) {
	key := makeKey()

	tests := []struct {
		name      string
		plaintext []byte
		aad       []byte
	}{
		{
			name:      "обычные данные",
			plaintext: []byte("секретные данные"),
			aad:       []byte("record-id-123"),
		},
		{
			name:      "пустой plaintext",
			plaintext: []byte{},
			aad:       []byte("id"),
		},
		{
			name:      "nil aad",
			plaintext: []byte("data"),
			aad:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, err := Encrypt(key, tt.plaintext, tt.aad)
			require.NoError(t, err)
			assert.Greater(t, len(ct), nonceSize)

			got, err := Decrypt(key, ct, tt.aad)
			require.NoError(t, err)
			// bytes.Equal корректно обрабатывает nil == []byte{}
			assert.True(t, bytes.Equal(tt.plaintext, got))
		})
	}
}

func TestEncrypt_UniqueNonce(t *testing.T) {
	key := makeKey()
	plaintext := []byte("одинаковый текст")

	ct1, err := Encrypt(key, plaintext, nil)
	require.NoError(t, err)

	ct2, err := Encrypt(key, plaintext, nil)
	require.NoError(t, err)

	assert.NotEqual(t, ct1, ct2, "два шифрования не должны давать одинаковый результат")
}

func TestDecrypt_Errors(t *testing.T) {
	key := makeKey()
	ct, err := Encrypt(key, []byte("данные"), []byte("id"))
	require.NoError(t, err)

	tests := []struct {
		name string
		key  []byte
		data []byte
		aad  []byte
	}{
		{
			name: "неверный размер ключа",
			key:  []byte("short"),
			data: ct,
			aad:  []byte("id"),
		},
		{
			name: "слишком короткие данные",
			key:  key,
			data: []byte("short"),
			aad:  []byte("id"),
		},
		{
			name: "повреждённый ciphertext",
			key:  key,
			data: func() []byte { c := append([]byte{}, ct...); c[len(c)-1] ^= 0xFF; return c }(),
			aad:  []byte("id"),
		},
		{
			name: "неверный aad",
			key:  key,
			data: ct,
			aad:  []byte("wrong-id"),
		},
		{
			name: "неверный ключ",
			key:  make([]byte, 32),
			data: ct,
			aad:  []byte("id"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.key, tt.data, tt.aad)
			require.Error(t, err)
		})
	}
}

func TestEncrypt_BadKeySize(t *testing.T) {
	_, err := Encrypt([]byte("short"), []byte("data"), nil)
	require.Error(t, err)
}
