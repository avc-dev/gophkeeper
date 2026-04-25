package auth

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Login(ctx context.Context, email, password string) (token string, kdfSalt []byte, err error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		// намеренно не различаем "не найден" и "неверный пароль" — защита от user enumeration
		return "", nil, fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid email or password")
	}

	token, err = s.issueToken(user.ID)
	if err != nil {
		return "", nil, fmt.Errorf("login: %w", err)
	}

	return token, user.KDFSalt, nil
}
