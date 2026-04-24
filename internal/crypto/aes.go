package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

const nonceSize = 12

// Encrypt шифрует plaintext с помощью AES-256-GCM.
// Возвращает nonce || ciphertext. aad защищает от перестановки зашифрованных блоков.
func Encrypt(key, plaintext, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}

	// Seal добавляет тег аутентификации и возвращает nonce || ciphertext || tag.
	return gcm.Seal(nonce, nonce, plaintext, aad), nil
}

// Decrypt расшифровывает данные, зашифрованные Encrypt.
func Decrypt(key, data, aad []byte) ([]byte, error) {
	if len(data) < nonceSize {
		return nil, errors.New("данные повреждены: слишком короткие")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], aad)
	if err != nil {
		return nil, fmt.Errorf("расшифровка: %w", err)
	}

	return plaintext, nil
}
