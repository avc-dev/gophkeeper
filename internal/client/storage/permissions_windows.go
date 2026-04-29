//go:build windows

package storage

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// restrictPermissions устанавливает ACL на файл БД: доступ только текущему пользователю.
func restrictPermissions(path string) error {
	// создаём файл заранее.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create db file: %w", err)
	}
	f.Close()

	// получаем SID текущего пользователя.
	token, err := windows.OpenCurrentProcessToken()
	if err != nil {
		return fmt.Errorf("open process token: %w", err)
	}
	defer token.Close()

	user, err := token.GetTokenUser()
	if err != nil {
		return fmt.Errorf("get token user: %w", err)
	}

	// строим DACL: полный доступ только для текущего SID.
	sd, err := windows.SecurityDescriptorFromString(
		fmt.Sprintf("D:(A;;FA;;;%s)", user.User.Sid.String()),
	)
	if err != nil {
		return fmt.Errorf("build security descriptor: %w", err)
	}

	pathUTF16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("utf16 path: %w", err)
	}

	dacl, _, err := sd.DACL()
	if err != nil {
		return fmt.Errorf("extract dacl: %w", err)
	}

	if err := windows.SetNamedSecurityInfo(
		pathUTF16,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil, nil, dacl, nil,
	); err != nil {
		return fmt.Errorf("set file acl: %w", err)
	}

	return nil
}
