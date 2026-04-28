package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ValidateToken проверяет JWT (EdDSA/Ed25519) и возвращает user_id.
// Проверка выполняется только публичным ключом — приватный ключ не нужен и не используется.
func (s *Service) ValidateToken(tokenStr string) (uuid.UUID, error) {
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil || !t.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid claims")
	}

	idStr, ok := claims[claimUserID].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("uid missing from token")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uid in token")
	}

	return id, nil
}
