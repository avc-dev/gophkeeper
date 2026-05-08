package secret

import (
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

const otpPeriod = 30 // секунд — стандартный период TOTP (RFC 6238)

// GenerateOTP генерирует текущий TOTP-код из base32-семени.
// Возвращает шестизначный код и количество секунд до его истечения.
func GenerateOTP(seed string) (code string, secondsLeft int, err error) {
	if err = ValidateOTPSeed(seed); err != nil {
		return "", 0, err
	}
	clean := strings.ToUpper(strings.ReplaceAll(seed, " ", ""))
	code, err = totp.GenerateCode(clean, time.Now())
	if err != nil {
		return "", 0, fmt.Errorf("generate otp code: %w", err)
	}
	secondsLeft = otpPeriod - int(time.Now().Unix()%otpPeriod)
	return code, secondsLeft, nil
}
