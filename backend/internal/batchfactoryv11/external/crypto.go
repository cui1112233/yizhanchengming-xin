package external

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// EncryptCredential encrypts a provider secret with AES-256-GCM. The nonce is
// random per value and the key identifier lets operators rotate keys without
// ever returning the plaintext credential to callers.
func EncryptCredential(key, plaintext []byte) (keyID string, nonce, ciphertext []byte, err error) {
	if len(key) != 32 {
		return "", nil, nil, fmt.Errorf("credential encryption key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", nil, nil, err
	}
	var gcm cipher.AEAD
	gcm, err = cipher.NewGCM(block)
	if err != nil {
		return "", nil, nil, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err := readRandom(nonce); err != nil {
		return "", nil, nil, err
	}
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	hash := sha256.Sum256(key)
	return hex.EncodeToString(hash[:8]), nonce, ciphertext, nil
}

func DecryptCredential(key, nonce, ciphertext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("credential encryption key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("invalid credential nonce")
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}
