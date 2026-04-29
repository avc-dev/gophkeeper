package domain

import (
	"time"

	"github.com/google/uuid"
)

// User — учётная запись пользователя на сервере.
// PasswordHash хранит bcrypt-хеш пароля (cost=12).
// KDFSalt генерируется сервером при регистрации и передаётся клиенту при Login
// для деривации мастер-ключа; сервер его не использует.
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	KDFSalt      []byte
	CreatedAt    time.Time
}
