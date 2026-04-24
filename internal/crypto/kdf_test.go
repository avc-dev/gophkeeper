package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKey_Deterministic(t *testing.T) {
	salt := make([]byte, SaltSize)
	password := "мой-мастер-пароль"

	k1 := DeriveKey(password, salt)
	k2 := DeriveKey(password, salt)

	if !bytes.Equal(k1, k2) {
		t.Fatal("DeriveKey не детерминирован: одинаковые входы дали разные ключи")
	}
}

func TestDeriveKey_Length(t *testing.T) {
	key := DeriveKey("password", make([]byte, SaltSize))
	if len(key) != keySize {
		t.Fatalf("длина ключа %d, ожидалось %d", len(key), keySize)
	}
}

func TestDeriveKey_DifferentSalt(t *testing.T) {
	salt1 := make([]byte, SaltSize)
	salt2 := make([]byte, SaltSize)
	salt2[0] = 1

	k1 := DeriveKey("password", salt1)
	k2 := DeriveKey("password", salt2)

	if bytes.Equal(k1, k2) {
		t.Fatal("разные соли дали одинаковый ключ")
	}
}

func TestDeriveKey_DifferentPassword(t *testing.T) {
	salt := make([]byte, SaltSize)

	k1 := DeriveKey("password1", salt)
	k2 := DeriveKey("password2", salt)

	if bytes.Equal(k1, k2) {
		t.Fatal("разные пароли дали одинаковый ключ")
	}
}

func TestGenerateSalt_Length(t *testing.T) {
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt: %v", err)
	}

	if len(salt) != SaltSize {
		t.Fatalf("длина соли %d, ожидалось %d", len(salt), SaltSize)
	}
}

func TestGenerateSalt_Uniqueness(t *testing.T) {
	s1, _ := GenerateSalt()
	s2, _ := GenerateSalt()

	if bytes.Equal(s1, s2) {
		t.Fatal("две соли совпадают — генератор не случаен")
	}
}
