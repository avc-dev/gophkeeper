package auth

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Login(ctx context.Context, email, password string) (token string, kdfSalt []byte, err error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		// bcrypt-вызов нормализует время ответа, чтобы атакующий не мог по задержке
		// определить, зарегистрирован ли email (защита от user enumeration via timing).
		_ = bcrypt.CompareHashAndPassword(getDummyHash(), []byte(password))
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
