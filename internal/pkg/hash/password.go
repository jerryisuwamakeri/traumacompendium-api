// Package hash provides password and token hashing helpers backed by
// bcrypt (passwords, where cost matters) and SHA-256 (opaque tokens like
// OTPs and refresh tokens, where we only need a fast, irreversible digest
// to store instead of the raw secret).
package hash

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// defaultCost mirrors bcrypt's recommended default; raising it trades
// login latency for brute-force resistance.
const defaultCost = bcrypt.DefaultCost

// Password hashes a plaintext password for storage.
func Password(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), defaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// PasswordMatches reports whether plain matches the given bcrypt hash.
func PasswordMatches(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

// Token returns the hex-encoded SHA-256 digest of an opaque secret (OTP
// code, refresh token). We store this digest instead of the raw value so a
// database leak doesn't hand out usable credentials directly.
func Token(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// TokenMatches performs a constant-time comparison between a stored digest
// and the digest of a candidate raw token.
func TokenMatches(storedDigest, candidateRaw string) bool {
	candidateDigest := Token(candidateRaw)
	return subtle.ConstantTimeCompare([]byte(storedDigest), []byte(candidateDigest)) == 1
}
