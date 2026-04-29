package domain

import (
	"time"

	"github.com/google/uuid"
)

// SecretType — дискриминатор типа секрета. Определяет структуру зашифрованного payload-а.
type SecretType string

// Допустимые типы секретов.
const (
	SecretTypeCredential SecretType = "credential" // логин + пароль (+ URL, заметка)
	SecretTypeCard       SecretType = "card"       // данные банковской карты
	SecretTypeText       SecretType = "text"       // произвольный текст
	SecretTypeBinary     SecretType = "binary"     // бинарный файл
	SecretTypeOTP        SecretType = "otp"        // TOTP-семя для двухфакторной аутентификации
)

// Secret — зашифрованный секрет пользователя.
// Payload хранится в зашифрованном виде (AES-256-GCM); сервер не знает его содержимого.
// Version используется для оптимистичной блокировки при обновлении.
type Secret struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      SecretType
	Name      string
	Payload   []byte // зашифрован на клиенте, сервер хранит как есть
	Metadata  string // JSON, plaintext (MVP)
	Version   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
