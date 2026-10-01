// Package otp generates short numeric one-time codes for email and phone
// verification flows.
package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Generate returns a cryptographically random numeric code of the given
// length (e.g. length=6 -> "048213"). crypto/rand is used deliberately
// instead of math/rand — these codes gate account verification and
// password resets, so they must not be predictable.
func Generate(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("otp: length must be positive, got %d", length)
	}

	digits := make([]byte, length)
	for i := range digits {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("otp: generate digit: %w", err)
		}
		digits[i] = byte('0') + byte(n.Int64())
	}

	return string(digits), nil
}
