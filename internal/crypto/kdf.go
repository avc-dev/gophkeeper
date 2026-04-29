package crypto

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	SaltSize = 16
	keySize  = 32

	// Параметры Argon2id по рекомендации OWASP (минимальные для интерактивного входа).
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
)

// DeriveKey выводит 32-байтовый ключ из пароля и соли с помощью Argon2id.
// Детерминирован: одинаковые password + salt всегда дают одинаковый ключ.
func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, keySize)
}

// GenerateSalt генерирует случайную 16-байтовую соль для KDF.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("генерация соли: %w", err)
	}
	return salt, nil
}
