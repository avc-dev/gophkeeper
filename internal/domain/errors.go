package domain

import "errors"

var (
	ErrEmailTaken      = errors.New("email already taken")
	ErrSecretNotFound  = errors.New("secret not found")
	ErrVersionConflict = errors.New("version conflict")
)
