package providerconfig

import (
	"bytes"
	"testing"
)

func TestCipherRoundTripsProviderCredential(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32)
	cipher, err := NewCipher(key)
	if err != nil { t.Fatal(err) }
	nonce, encrypted, err := cipher.Encrypt("owner-1", "video", "personal_api", []byte("secret-api-key"))
	if err != nil { t.Fatal(err) }
	if bytes.Contains(encrypted, []byte("secret-api-key")) { t.Fatal("ciphertext must not contain plaintext") }
	plain, err := cipher.Decrypt("owner-1", "video", "personal_api", nonce, encrypted)
	if err != nil { t.Fatal(err) }
	if string(plain) != "secret-api-key" { t.Fatalf("plain=%q", plain) }
}

func TestCipherBindsCredentialToOwnerKindAndProvider(t *testing.T) {
	cipher, _ := NewCipher(bytes.Repeat([]byte{0x24}, 32))
	nonce, encrypted, err := cipher.Encrypt("alice", "video", "personal_api", []byte("key"))
	if err != nil { t.Fatal(err) }
	for _, tc := range [][3]string{{"bob", "video", "personal_api"}, {"alice", "image", "personal_api"}, {"alice", "video", "yfai_seedance"}} {
		if _, err := cipher.Decrypt(tc[0], tc[1], tc[2], nonce, encrypted); err == nil { t.Fatalf("expected AAD mismatch for %#v", tc) }
	}
}

func TestNewCipherRequiresAES256Key(t *testing.T) {
	for _, size := range []int{0, 16, 24, 31, 33} {
		if _, err := NewCipher(make([]byte, size)); err == nil { t.Fatalf("expected size %d to fail", size) }
	}
	if _, err := NewCipher(make([]byte, 32)); err != nil { t.Fatal(err) }
}
