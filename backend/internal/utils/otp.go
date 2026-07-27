package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateNumericOTP returns an n-digit numeric code (zero-padded).
func GenerateNumericOTP(digits int) (string, error) {
	if digits < 4 || digits > 8 {
		return "", fmt.Errorf("digits must be between 4 and 8")
	}

	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, n.Int64()), nil
}
