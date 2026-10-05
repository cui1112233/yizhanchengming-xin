package video

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func newGCM(masterKey []byte) (cipher.AEAD, error) {
	if len(masterKey) != 32 {
		return nil, providerError(ErrorProviderConfigDecryptFailed, "provider master key must be 32 bytes", nil)
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, providerError(ErrorProviderConfigDecryptFailed, "create provider secret cipher", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, providerError(ErrorProviderConfigDecryptFailed, "create provider secret gcm", err)
	}
	return gcm, nil
}

func EncryptSecret(masterKey []byte, plaintext string) ([]byte, []byte, error) {
	if plaintext == "" {
		return nil, nil, providerError(ErrorProviderUnconfigured, "provider secret is required", nil)
	}
	gcm, err := newGCM(masterKey)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("video: generate provider secret nonce: %w", err)
	}
	return gcm.Seal(nil, nonce, []byte(plaintext), nil), nonce, nil
}

func DecryptSecret(masterKey, ciphertext, nonce []byte) (string, error) {
	if len(ciphertext) == 0 || len(nonce) == 0 {
		return "", providerError(ErrorProviderUnconfigured, "provider secret is not configured", nil)
	}
	gcm, err := newGCM(masterKey)
	if err != nil {
		return "", err
	}
	if len(nonce) != gcm.NonceSize() {
		return "", providerError(ErrorProviderConfigDecryptFailed, "provider secret nonce is invalid", nil)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", providerError(ErrorProviderConfigDecryptFailed, "provider secret decryption failed", err)
	}
	return string(plaintext), nil
}
