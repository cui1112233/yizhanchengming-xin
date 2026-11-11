package apisecrets

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testKey() []byte {
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return key
}

func clearKeyEnv(t *testing.T) {
	t.Helper()
	t.Setenv(KeyEnvVar, "")
	t.Setenv(KeyFileEnvVar, "")
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	cipher, err := NewCipher(testKey(), "test")
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	encrypted, err := cipher.Encrypt("sk-live-secret")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if !IsEncrypted(encrypted) || !strings.HasPrefix(encrypted, Prefix) {
		t.Fatalf("ciphertext must carry prefix, got %q", encrypted)
	}
	if strings.Contains(encrypted, "sk-live-secret") {
		t.Fatal("ciphertext must not contain plaintext")
	}
	plain, err := cipher.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if plain != "sk-live-secret" {
		t.Fatalf("round trip mismatch: %q", plain)
	}
}

func TestDecryptKeepsLegacyPlaintext(t *testing.T) {
	cipher, err := NewCipher(testKey(), "test")
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	plain, err := cipher.Decrypt("sk-legacy-plaintext")
	if err != nil {
		t.Fatalf("plaintext must stay readable: %v", err)
	}
	if plain != "sk-legacy-plaintext" {
		t.Fatalf("legacy plaintext changed: %q", plain)
	}
}

func TestEncryptIsIdempotent(t *testing.T) {
	cipher, err := NewCipher(testKey(), "test")
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	once, err := cipher.Encrypt("sk-once")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	twice, err := cipher.Encrypt(once)
	if err != nil {
		t.Fatalf("Encrypt(ciphertext): %v", err)
	}
	if twice != once {
		t.Fatal("already encrypted value must not be encrypted twice")
	}
	if empty, err := cipher.Encrypt(""); err != nil || empty != "" {
		t.Fatalf("empty value should stay empty, got %q err=%v", empty, err)
	}
}

func TestDecryptRejectsWrongKeyAndTampering(t *testing.T) {
	cipher, err := NewCipher(testKey(), "test")
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	encrypted, err := cipher.Encrypt("sk-tamper")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	otherKey := testKey()
	otherKey[0] = 99
	other, err := NewCipher(otherKey, "other")
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	if _, err := other.Decrypt(encrypted); err == nil {
		t.Fatal("wrong key must fail decryption")
	}
	tampered := encrypted[:len(encrypted)-2] + "AA"
	if _, err := cipher.Decrypt(tampered); err == nil {
		t.Fatal("tampered ciphertext must fail decryption")
	}
	if _, err := cipher.Decrypt(Prefix + "not-base64"); err == nil {
		t.Fatal("invalid payload must fail decryption")
	}
}

func TestProviderCreatesKeyFileOnlyOnWrite(t *testing.T) {
	clearKeyEnv(t)
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "system", "config-secret.key")
	provider := &Provider{KeyFile: keyFile}

	if _, err := provider.ReadCipher(); err == nil {
		t.Fatal("read must not fabricate a key")
	}
	if _, err := os.Stat(keyFile); !os.IsNotExist(err) {
		t.Fatal("read path must not create the key file")
	}

	write, err := provider.WriteCipher()
	if err != nil {
		t.Fatalf("WriteCipher: %v", err)
	}
	info, err := os.Stat(keyFile)
	if err != nil {
		t.Fatalf("key file should exist after write: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("key file must not be empty")
	}

	read, err := provider.ReadCipher()
	if err != nil {
		t.Fatalf("ReadCipher after create: %v", err)
	}
	encrypted, err := write.Encrypt("sk-from-file")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	plain, err := read.Decrypt(encrypted)
	if err != nil || plain != "sk-from-file" {
		t.Fatalf("key file round trip failed: %q err=%v", plain, err)
	}
}

func TestProviderPrefersEnvironmentKey(t *testing.T) {
	clearKeyEnv(t)
	envKey := make([]byte, KeySize)
	for i := range envKey {
		envKey[i] = byte(255 - i)
	}
	t.Setenv(KeyEnvVar, base64.StdEncoding.EncodeToString(envKey))
	provider := &Provider{KeyFile: filepath.Join(t.TempDir(), "unused.key")}
	cipher, err := provider.WriteCipher()
	if err != nil {
		t.Fatalf("WriteCipher: %v", err)
	}
	if cipher.Source() != KeyEnvVar {
		t.Fatalf("env key should win, got source %q", cipher.Source())
	}
	if _, err := os.Stat(provider.KeyFile); !os.IsNotExist(err) {
		t.Fatal("env key must not touch the default key file")
	}
}

func TestDecodeKeyAcceptsBase64AndRaw(t *testing.T) {
	raw := strings.Repeat("k", KeySize)
	if key, err := decodeKey(raw); err != nil || string(key) != raw {
		t.Fatalf("raw key should decode: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))
	if key, err := decodeKey(encoded); err != nil || string(key) != raw {
		t.Fatalf("base64 key should decode: %v", err)
	}
	if _, err := decodeKey("too-short"); err == nil {
		t.Fatal("invalid key length must fail")
	}
}
