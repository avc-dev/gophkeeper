package crypto

import (
	"bytes"
	"testing"
)

func makeKey() []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return key
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := makeKey()
	plaintext := []byte("секретные данные")
	aad := []byte("record-id-123")

	ciphertext, err := Encrypt(key, plaintext, aad)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	if len(ciphertext) <= nonceSize {
		t.Fatal("зашифрованные данные слишком короткие")
	}

	got, err := Decrypt(key, ciphertext, aad)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(got, plaintext) {
		t.Fatalf("got %q, want %q", got, plaintext)
	}
}

func TestEncrypt_DifferentNonceEachCall(t *testing.T) {
	key := makeKey()
	plaintext := []byte("одинаковый текст")
	aad := []byte("id")

	ct1, _ := Encrypt(key, plaintext, aad)
	ct2, _ := Encrypt(key, plaintext, aad)

	if bytes.Equal(ct1, ct2) {
		t.Fatal("два шифрования дали одинаковый результат — nonce не случаен")
	}
}

func TestEncrypt_BadKeySize(t *testing.T) {
	_, err := Encrypt([]byte("short"), []byte("data"), nil)
	if err == nil {
		t.Fatal("ожидалась ошибка при неверном размере ключа")
	}
}

func TestDecrypt_TooShortData(t *testing.T) {
	_, err := Decrypt(makeKey(), []byte("short"), nil)
	if err == nil {
		t.Fatal("ожидалась ошибка на слишком коротких данных")
	}
}

func TestDecrypt_BadKeySize(t *testing.T) {
	data := make([]byte, nonceSize+16)
	_, err := Decrypt([]byte("short"), data, nil)
	if err == nil {
		t.Fatal("ожидалась ошибка при неверном размере ключа")
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	key := makeKey()
	ct, _ := Encrypt(key, []byte("данные"), []byte("id"))

	ct[len(ct)-1] ^= 0xFF // портим последний байт тега

	_, err := Decrypt(key, ct, []byte("id"))
	if err == nil {
		t.Fatal("ожидалась ошибка при повреждённом ciphertext")
	}
}

func TestDecrypt_WrongAAD(t *testing.T) {
	key := makeKey()
	ct, _ := Encrypt(key, []byte("данные"), []byte("correct-id"))

	_, err := Decrypt(key, ct, []byte("wrong-id"))
	if err == nil {
		t.Fatal("ожидалась ошибка при неверном aad")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key := makeKey()
	ct, _ := Encrypt(key, []byte("данные"), []byte("id"))

	wrongKey := make([]byte, 32)
	_, err := Decrypt(wrongKey, ct, []byte("id"))
	if err == nil {
		t.Fatal("ожидалась ошибка при неверном ключе")
	}
}

func TestEncryptDecrypt_EmptyPlaintext(t *testing.T) {
	key := makeKey()
	ct, err := Encrypt(key, []byte{}, []byte("id"))
	if err != nil {
		t.Fatalf("Encrypt пустого plaintext: %v", err)
	}

	got, err := Decrypt(key, ct, []byte("id"))
	if err != nil {
		t.Fatalf("Decrypt пустого plaintext: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("ожидался пустой plaintext, got %q", got)
	}
}
