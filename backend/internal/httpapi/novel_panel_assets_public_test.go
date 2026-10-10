package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
)

func TestReferenceAssetSignatureRejectsTamperingAndAcceptsValidPayload(t *testing.T) {
	secret := "asset-secret"
	username, assetType, assetID, variant, expiresAt := "alice", "character", "hero_1", "main", int64(1893456000)
	payload := strings.Join([]string{username, assetType, assetID, variant, strconv.FormatInt(expiresAt, 10)}, "\n")
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))
	if !validReferenceAssetSignature(secret, username, assetType, assetID, variant, expiresAt, signature) {
		t.Fatal("expected valid reference asset signature")
	}
	if validReferenceAssetSignature(secret, username, assetType, assetID, "thumb", expiresAt, signature) {
		t.Fatal("tampered variant must be rejected")
	}
}
