package secret

import (
	"encoding/json"
	"fmt"
)

// CredentialPayload — расшифрованные данные логина/пароля.
type CredentialPayload struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	URL      string `json:"url,omitempty"`
	Note     string `json:"note,omitempty"`
}

// CardPayload — расшифрованные данные банковской карты.
type CardPayload struct {
	Number string `json:"number"`
	Holder string `json:"holder"`
	Expiry string `json:"expiry"` // MM/YY
	CVV    string `json:"cvv"`
	Bank   string `json:"bank,omitempty"`
	Note   string `json:"note,omitempty"`
}

// TextPayload — расшифрованный произвольный текст.
type TextPayload struct {
	Content string `json:"content"`
	Note    string `json:"note,omitempty"`
}

// BinaryPayload хранит файл в явном base64: Data — строка, не []byte,
// чтобы кодирование было явным (в AddBinary), а не скрытым в json.Marshal.
type BinaryPayload struct {
	Filename string `json:"filename"`
	Data     string `json:"data"`
	Note     string `json:"note,omitempty"`
}

func marshalPayload(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return data, nil
}

func unmarshalCredential(data []byte) (*CredentialPayload, error) {
	var p CredentialPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("unmarshal credential: %w", err)
	}
	return &p, nil
}

func unmarshalCard(data []byte) (*CardPayload, error) {
	var p CardPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("unmarshal card: %w", err)
	}
	return &p, nil
}

func unmarshalText(data []byte) (*TextPayload, error) {
	var p TextPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("unmarshal text: %w", err)
	}
	return &p, nil
}

func unmarshalBinary(data []byte) (*BinaryPayload, error) {
	var p BinaryPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("unmarshal binary: %w", err)
	}
	return &p, nil
}
