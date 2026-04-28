package secret

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
