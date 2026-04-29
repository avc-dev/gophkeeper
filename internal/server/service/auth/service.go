package auth

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"sync"
	"time"

	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost  = 12
	tokenTTL    = 72 * time.Hour
	claimUserID = "uid"
)

// userStorage — локальный интерфейс; реализуется storage/user.Storage.
// Сервис не импортирует storage напрямую — зависимость инвертирована.
type userStorage interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

// dummyHash используется для нормализации времени ответа при несуществующем пользователе,
// чтобы предотвратить timing-атаку на перебор email-адресов (user enumeration).
var (
	dummyHashOnce sync.Once
	dummyHash     []byte
)

func getDummyHash() []byte {
	dummyHashOnce.Do(func() {
		h, _ := bcrypt.GenerateFromPassword([]byte("dummy-normalize-timing"), bcryptCost)
		dummyHash = h
	})
	return dummyHash
}

type Service struct {
	users      userStorage
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// New создаёт Service с EdDSA (Ed25519) JWT-подписью.
// privateKey используется только для подписи токенов (issueToken).
// publicKey используется только для проверки токенов (ValidateToken).
func New(users userStorage, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) *Service {
	return &Service{
		users:      users,
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (s *Service) issueToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		claimUserID: userID.String(),
		"iat":       now.Unix(),
		"exp":       now.Add(tokenTTL).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	signed, err := t.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(b), nil
}
