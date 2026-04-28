package auth

import (
	"context"
	"fmt"
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

type Service struct {
	users     userStorage
	jwtSecret []byte
}

func New(users userStorage, jwtSecret string) *Service {
	return &Service{
		users:     users,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *Service) issueToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		claimUserID: userID.String(),
		"exp":       time.Now().Add(tokenTTL).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(s.jwtSecret)
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
