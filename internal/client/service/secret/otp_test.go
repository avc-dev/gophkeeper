package secret

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validSeed — стандартный TOTP-тестовый ключ из RFC 6238.
const validSeed = "JBSWY3DPEHPK3PXP"

func TestGenerateOTP(t *testing.T) {
	reDigits := regexp.MustCompile(`^\d{6}$`)

	tests := []struct {
		name    string
		seed    string
		wantErr bool
	}{
		{
			name: "валидный seed — код 6 цифр",
			seed: validSeed,
		},
		{
			name: "seed с пробелами — нормализуется",
			seed: "JBSWY3DP EHPK3PXP",
		},
		{
			name: "seed в нижнем регистре — нормализуется",
			seed: "jbswy3dpehpk3pxp",
		},
		{
			name:    "пустой seed — ошибка",
			seed:    "",
			wantErr: true,
		},
		{
			name:    "невалидный seed — ошибка",
			seed:    "!!!invalid!!!",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, secondsLeft, err := GenerateOTP(tt.seed)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Regexp(t, reDigits, code, "код должен быть 6 цифр")
			assert.GreaterOrEqual(t, secondsLeft, 1, "secondsLeft >= 1")
			assert.LessOrEqual(t, secondsLeft, 30, "secondsLeft <= 30")
		})
	}
}

func TestValidateOTPSeed(t *testing.T) {
	tests := []struct {
		name    string
		seed    string
		wantErr error
	}{
		{
			name: "валидный seed",
			seed: validSeed,
		},
		{
			name: "seed без выравнивания",
			seed: "JBSWY3DPEHPK3PX", // длина не кратна 8 — дополняется автоматически
		},
		{
			name: "seed с пробелами",
			seed: "JBSWY3DP EHPK3PXP",
		},
		{
			name: "seed в нижнем регистре",
			seed: "jbswy3dpehpk3pxp",
		},
		{
			name:    "недопустимые символы",
			seed:    "INVALID-SEED-!!!",
			wantErr: ErrOTPSeedInvalid,
		},
		{
			name:    "пустой seed",
			seed:    "",
			wantErr: ErrOTPSeedInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOTPSeed(tt.seed)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
