package domain

import (
	"time"

	"github.com/google/uuid"
)

type SecretType string

const (
	SecretTypeCredential SecretType = "credential"
	SecretTypeCard       SecretType = "card"
	SecretTypeText       SecretType = "text"
	SecretTypeBinary     SecretType = "binary"
)

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
