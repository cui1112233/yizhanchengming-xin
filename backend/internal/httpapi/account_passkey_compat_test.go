package httpapi

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFinishPasskeyAuthenticationVerifiesAssertionAndConsumesChallenge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "passkeys.json")
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	publicDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	challenge := "test-challenge"
	clientData := []byte(`{"type":"webauthn.get","challenge":"test-challenge","origin":"http://localhost:4001"}`)
	rpHash := sha256.Sum256([]byte("localhost"))
	authenticatorData := make([]byte, 37)
	copy(authenticatorData[:32], rpHash[:])
	authenticatorData[32] = 0x05
	binary.BigEndian.PutUint32(authenticatorData[33:37], 1)
	clientHash := sha256.Sum256(clientData)
	signed := append(append([]byte{}, authenticatorData...), clientHash[:]...)
	digest := sha256.Sum256(signed)
	signature, err := ecdsa.SignASN1(rand.Reader, key, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	state := map[string]any{
		"credentials": []any{map[string]any{
			"id": "cred-1", "username": "alice", "credentialId": "credential-1",
			"publicKey": base64.RawURLEncoding.EncodeToString(publicDER), "algorithm": -7, "signCount": 0,
		}},
		"challenges": []any{map[string]any{
			"username": "alice", "kind": "authenticate", "challenge": challenge,
			"rpId": "localhost", "origin": "http://localhost:4001",
			"expiresAt": now.Add(time.Minute).Format(time.RFC3339), "usedAt": nil,
		}},
	}
	raw, _ := json.Marshal(state)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "http://localhost:4001/api/login/passkey/verify", nil)
	input := map[string]any{
		"challenge": challenge,
		"credential": map[string]any{
			"id": "credential-1", "clientDataJSON": base64.RawURLEncoding.EncodeToString(clientData),
			"authenticatorData": base64.RawURLEncoding.EncodeToString(authenticatorData), "signature": base64.RawURLEncoding.EncodeToString(signature),
		},
	}
	if err := finishPasskeyAuthentication(path, "alice", input, req); err != nil {
		t.Fatalf("finishPasskeyAuthentication() error = %v", err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if json.Unmarshal(updated, &decoded) != nil {
		t.Fatal("updated passkey state is invalid JSON")
	}
	credentials, _ := decoded["credentials"].([]any)
	if len(credentials) != 1 || credentials[0].(map[string]any)["signCount"] != float64(1) {
		t.Fatalf("sign count was not updated: %s", updated)
	}
	challenges, _ := decoded["challenges"].([]any)
	if len(challenges) != 1 || challenges[0].(map[string]any)["usedAt"] == nil {
		t.Fatalf("challenge was not consumed: %s", updated)
	}
}
