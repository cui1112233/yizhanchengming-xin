package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/scrypt"
	"strings"
)

const PasswordAlgorithmScryptV1 = "scrypt-v1"

func HashPassword(password string) (string, string, error) {
	if len(password) < 8 || len(password) > 128 {
		return "", "", ErrInvalid
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", "", err
	}
	sum, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("scrypt$%s$%s", hex.EncodeToString(salt), hex.EncodeToString(sum)), PasswordAlgorithmScryptV1, nil
}

func VerifyPassword(password, encoded, algorithm string) bool {
	saltHex, hashHex, ok := strings.Cut(strings.TrimPrefix(encoded, "scrypt$"), "$")
	if !ok || (algorithm != "scrypt" && algorithm != PasswordAlgorithmScryptV1) {
		return false
	}
	salt, err1 := hex.DecodeString(saltHex)
	expected, err2 := hex.DecodeString(hashHex)
	if err1 != nil || err2 != nil || len(salt) == 0 || len(expected) == 0 {
		return false
	}
	N, r, p := 16384, 8, 1
	if algorithm == PasswordAlgorithmScryptV1 {
		N = 32768
	}
	actual, err := scrypt.Key([]byte(password), salt, N, r, p, len(expected))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func sessionTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
