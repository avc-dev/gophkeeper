package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avc-dev/gophkeeper/internal/crypto"
	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

// Register создаёт нового пользователя с bcrypt-хешем пароля и Argon2id KDF-солью.
// Возвращает JWT и KDF-соль для немедленной деривации мастер-ключа.
func (s *Service) Register(ctx context.Context, email, password string) (token string, kdfSalt []byte, err error) {
	hash, err := hashPassword(password)
	if err != nil {
		return "", nil, fmt.Errorf("register: %w", err)
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		return "", nil, fmt.Errorf("register: generate salt: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		KDFSalt:      salt,
		CreatedAt:    time.Now(),
	}

	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrEmailTaken) {
			// sentinel пробрасывается как есть — handler проверяет через errors.Is
			return "", nil, err
		}
		return "", nil, fmt.Errorf("register: create user: %w", err)
	}

	token, err = s.issueToken(user.ID)
	if err != nil {
		return "", nil, fmt.Errorf("register: %w", err)
	}

	return token, salt, nil
}
