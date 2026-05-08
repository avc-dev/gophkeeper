package secret

import (
	"encoding/base32"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	// ErrOTPSeedInvalid возвращается, если OTP-семя не является валидным base32-ключом.
	ErrOTPSeedInvalid = errors.New("OTP seed is not valid base32")

	// ErrLuhnInvalid возвращается, если номер карты не проходит проверку по алгоритму Луна.
	ErrLuhnInvalid = errors.New("card number failed Luhn check")
	// ErrExpiryPast возвращается, если срок действия карты истёк.
	ErrExpiryPast = errors.New("card expiry date is in the past")
	// ErrExpiryFormat возвращается, если формат срока действия не соответствует MM/YY.
	ErrExpiryFormat = errors.New("card expiry must be MM/YY")
	// ErrCVVInvalid возвращается, если CVV-код не является трёх- или четырёхзначным числом.
	ErrCVVInvalid = errors.New("CVV must be 3 or 4 digits")
)

var reExpiry = regexp.MustCompile(`^(0[1-9]|1[0-2])/(\d{2})$`)

// ValidateOTPSeed проверяет, что seed является валидным base32-ключом TOTP.
// Пробелы и символы нижнего регистра допустимы и нормализуются автоматически.
func ValidateOTPSeed(seed string) error {
	clean := strings.ToUpper(strings.ReplaceAll(seed, " ", ""))
	if len(clean) == 0 {
		return fmt.Errorf("%w: seed is empty", ErrOTPSeedInvalid)
	}
	// base32 требует длину кратную 8; дополняем знаками "=" до нужного размера
	if rem := len(clean) % 8; rem != 0 {
		clean += strings.Repeat("=", 8-rem)
	}
	if _, err := base32.StdEncoding.DecodeString(clean); err != nil {
		return fmt.Errorf("%w: %s", ErrOTPSeedInvalid, err)
	}
	return nil
}

// ValidateCard проверяет реквизиты банковской карты: номер (алгоритм Луна), срок действия (MM/YY) и CVV.
func ValidateCard(number, expiry, cvv string) error {
	if err := validateLuhn(number); err != nil {
		return err
	}
	if err := validateExpiry(expiry); err != nil {
		return err
	}
	return validateCVV(cvv)
}

func validateLuhn(number string) error {
	clean := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, number)

	if len(clean) < 13 || len(clean) > 19 {
		return fmt.Errorf("%w: got %d digits", ErrLuhnInvalid, len(clean))
	}

	sum := 0
	nDigits := len(clean)
	parity := nDigits % 2

	for i := 0; i < nDigits; i++ {
		digit, _ := strconv.Atoi(string(clean[i]))
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	if sum%10 != 0 {
		return ErrLuhnInvalid
	}
	return nil
}

func validateExpiry(expiry string) error {
	m := reExpiry.FindStringSubmatch(expiry)
	if m == nil {
		return ErrExpiryFormat
	}
	month, _ := strconv.Atoi(m[1])
	year, _ := strconv.Atoi("20" + m[2])
	expTime := time.Date(year, time.Month(month)+1, 0, 23, 59, 59, 0, time.UTC)
	if expTime.Before(time.Now()) {
		return ErrExpiryPast
	}
	return nil
}

func validateCVV(cvv string) error {
	if len(cvv) != 3 && len(cvv) != 4 {
		return ErrCVVInvalid
	}
	for _, r := range cvv {
		if !unicode.IsDigit(r) {
			return ErrCVVInvalid
		}
	}
	return nil
}
