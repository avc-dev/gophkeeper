package service

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	ErrLuhnInvalid  = errors.New("card number failed Luhn check")
	ErrExpiryPast   = errors.New("card expiry date is in the past")
	ErrExpiryFormat = errors.New("card expiry must be MM/YY")
	ErrCVVInvalid   = errors.New("CVV must be 3 or 4 digits")
)

var reExpiry = regexp.MustCompile(`^(0[1-9]|1[0-2])/(\d{2})$`)

// ValidateCard проверяет номер карты (Luhn), срок действия и CVV.
func ValidateCard(number, expiry, cvv string) error {
	if err := validateLuhn(number); err != nil {
		return err
	}
	if err := validateExpiry(expiry); err != nil {
		return err
	}
	if err := validateCVV(cvv); err != nil {
		return err
	}
	return nil
}

// validateLuhn проверяет контрольную сумму номера карты по алгоритму Луна.
func validateLuhn(number string) error {
	// убираем пробелы и дефисы для удобства ввода.
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

// validateExpiry проверяет формат MM/YY и что дата не в прошлом.
func validateExpiry(expiry string) error {
	m := reExpiry.FindStringSubmatch(expiry)
	if m == nil {
		return ErrExpiryFormat
	}
	month, _ := strconv.Atoi(m[1])
	year, _ := strconv.Atoi("20" + m[2])

	now := time.Now()
	// карта действительна до последнего дня месяца истечения.
	expTime := time.Date(year, time.Month(month)+1, 0, 23, 59, 59, 0, time.UTC)
	if expTime.Before(now) {
		return ErrExpiryPast
	}
	return nil
}

// validateCVV проверяет что CVV состоит из 3 или 4 цифр.
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
