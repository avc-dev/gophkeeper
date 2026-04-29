package domain

import "errors"

// Sentinel-ошибки доменного слоя. Используйте errors.Is для сравнения.
var (
	// ErrEmailTaken возвращается при попытке зарегистрировать уже занятый email.
	ErrEmailTaken = errors.New("email already taken")
	// ErrSecretNotFound возвращается, когда секрет не найден в хранилище.
	ErrSecretNotFound = errors.New("secret not found")
	// ErrVersionConflict возвращается при обновлении с устаревшей expected_version.
	ErrVersionConflict = errors.New("version conflict")
)
