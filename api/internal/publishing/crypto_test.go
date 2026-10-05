package publishing

import (
	"bytes"
	"testing"
)

func TestCredentialEncryptionUsesAES256GCMAndDoesNotPersistPlaintext(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32)
	plaintext := []byte("platform-cookie-secret")

	keyID, nonce, ciphertext, err := EncryptCredential(key, plaintext)
	if err != nil { t.Fatalf("encrypt: %v", err) }
	if keyID == "" || len(nonce) == 0 || len(ciphertext) == 0 { t.Fatalf("missing encrypted metadata: key=%q nonce=%d ciphertext=%d", keyID, len(nonce), len(ciphertext)) }
	if bytes.Contains(ciphertext, plaintext) || bytes.Equal(ciphertext, plaintext) { t.Fatal("ciphertext must not contain plaintext secret") }

	decrypted, err := DecryptCredential(key, nonce, ciphertext)
	if err != nil { t.Fatalf("decrypt: %v", err) }
	if !bytes.Equal(decrypted, plaintext) { t.Fatalf("decrypted=%q want %q", decrypted, plaintext) }
}

func TestCredentialEncryptionRejectsNon256BitKey(t *testing.T) {
	if _, _, _, err := EncryptCredential([]byte("short"), []byte("secret")); err == nil {
		t.Fatal("expected invalid encryption key to fail")
	}
}
