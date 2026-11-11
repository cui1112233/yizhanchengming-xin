package external

import (
	"bytes"
	"testing"
)

func TestCredentialEncryptionRoundTripAndRandomNonce(t *testing.T) {
	key := bytes.Repeat([]byte{7}, 32)
	keyID1, nonce1, ciphertext1, err := EncryptCredential(key, []byte("secret-value"))
	if err != nil {
		t.Fatal(err)
	}
	keyID2, nonce2, ciphertext2, err := EncryptCredential(key, []byte("secret-value"))
	if err != nil {
		t.Fatal(err)
	}
	if keyID1 != keyID2 || bytes.Equal(nonce1, nonce2) || bytes.Equal(ciphertext1, ciphertext2) {
		t.Fatal("encryption must use a stable key id and unique ciphertext")
	}
	plain, err := DecryptCredential(key, nonce1, ciphertext1)
	if err != nil || string(plain) != "secret-value" {
		t.Fatalf("round trip failed: %q %v", plain, err)
	}
	if bytes.Contains(ciphertext1, []byte("secret-value")) {
		t.Fatal("ciphertext contains plaintext")
	}
}

func TestCredentialEncryptionRejectsWrongKey(t *testing.T) {
	_, nonce, ciphertext, err := EncryptCredential(bytes.Repeat([]byte{1}, 32), []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecryptCredential(bytes.Repeat([]byte{2}, 32), nonce, ciphertext); err == nil {
		t.Fatal("wrong key must fail")
	}
}
