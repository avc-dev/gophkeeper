//go:build !windows

package storage

import (
	"fmt"
	"os"
)

// restrictPermissions устанавливает права 0600 на файл БД (только владелец).
// На Windows — no-op (используется ACL, см. permissions_windows.go).
func restrictPermissions(path string) error {
	// создаём файл заранее чтобы сразу выставить права; sqlite откроет его сам.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create db file: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("closing file: %w", err)
	}

	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("chmod db file: %w", err)
	}
	return nil
}
