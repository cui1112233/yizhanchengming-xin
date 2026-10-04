package providerconfig

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"strings"
)

type Cipher struct { aead cipher.AEAD }

func NewCipher(key []byte) (*Cipher, error) {
	if len(key) != 32 { return nil, errors.New("provider credential key must be exactly 32 bytes") }
	block, err := aes.NewCipher(key)
	if err != nil { return nil, err }
	aead, err := cipher.NewGCM(block)
	if err != nil { return nil, err }
	return &Cipher{aead: aead}, nil
}

func (c *Cipher) Encrypt(owner, kind, provider string, plaintext []byte) ([]byte, []byte, error) {
	if c == nil || c.aead == nil { return nil, nil, errors.New("credential cipher unavailable") }
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil { return nil, nil, err }
	ciphertext := c.aead.Seal(nil, nonce, plaintext, credentialAAD(owner, kind, provider))
	return nonce, ciphertext, nil
}

func (c *Cipher) Decrypt(owner, kind, provider string, nonce, ciphertext []byte) ([]byte, error) {
	if c == nil || c.aead == nil { return nil, errors.New("credential cipher unavailable") }
	if len(nonce) != c.aead.NonceSize() { return nil, errors.New("invalid credential nonce") }
	plaintext, err := c.aead.Open(nil, nonce, ciphertext, credentialAAD(owner, kind, provider))
	if err != nil { return nil, errors.New("decrypt provider credential") }
	return plaintext, nil
}

func credentialAAD(owner, kind, provider string) []byte {
	return []byte(strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(kind) + "\x00" + strings.TrimSpace(provider))
}
