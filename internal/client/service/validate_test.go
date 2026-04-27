package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		wantErr error
	}{
		{
			name:   "валидная карта Visa",
			number: "4532015112830366",
		},
		{
			name:   "валидная карта Mastercard",
			number: "5425233430109903",
		},
		{
			name:   "номер с пробелами",
			number: "4532 0151 1283 0366",
		},
		{
			name:   "номер с дефисами",
			number: "4532-0151-1283-0366",
		},
		{
			name:    "неверная контрольная сумма",
			number:  "4532015112830367",
			wantErr: ErrLuhnInvalid,
		},
		{
			name:    "слишком короткий",
			number:  "123456",
			wantErr: ErrLuhnInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLuhn(tt.number)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateExpiry(t *testing.T) {
	// генерируем валидный срок: следующий год.
	futureYear := time.Now().AddDate(1, 0, 0).Format("01/06")

	tests := []struct {
		name    string
		expiry  string
		wantErr error
	}{
		{
			name:   "валидный срок",
			expiry: futureYear,
		},
		{
			name:    "истёкший срок",
			expiry:  "01/20",
			wantErr: ErrExpiryPast,
		},
		{
			name:    "неверный формат — год четыре цифры",
			expiry:  "01/2025",
			wantErr: ErrExpiryFormat,
		},
		{
			name:    "неверный формат — без слэша",
			expiry:  "0125",
			wantErr: ErrExpiryFormat,
		},
		{
			name:    "неверный месяц",
			expiry:  "13/25",
			wantErr: ErrExpiryFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExpiry(tt.expiry)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateCVV(t *testing.T) {
	tests := []struct {
		name    string
		cvv     string
		wantErr error
	}{
		{name: "3 цифры", cvv: "123"},
		{name: "4 цифры (Amex)", cvv: "1234"},
		{name: "2 цифры", cvv: "12", wantErr: ErrCVVInvalid},
		{name: "5 цифр", cvv: "12345", wantErr: ErrCVVInvalid},
		{name: "буквы", cvv: "abc", wantErr: ErrCVVInvalid},
		{name: "пустой", cvv: "", wantErr: ErrCVVInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCVV(tt.cvv)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateCard(t *testing.T) {
	futureYear := time.Now().AddDate(1, 0, 0).Format("01/06")

	tests := []struct {
		name    string
		number  string
		expiry  string
		cvv     string
		wantErr bool
	}{
		{
			name:    "валидная карта",
			number:  "4532015112830366",
			expiry:  futureYear,
			cvv:     "123",
			wantErr: false,
		},
		{
			name:    "неверный номер",
			number:  "1234567890123456",
			expiry:  futureYear,
			cvv:     "123",
			wantErr: true,
		},
		{
			name:    "истёкшая карта",
			number:  "4532015112830366",
			expiry:  "01/20",
			cvv:     "123",
			wantErr: true,
		},
		{
			name:    "неверный CVV",
			number:  "4532015112830366",
			expiry:  futureYear,
			cvv:     "12",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCard(tt.number, tt.expiry, tt.cvv)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
