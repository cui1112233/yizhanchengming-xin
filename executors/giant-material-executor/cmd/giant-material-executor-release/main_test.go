package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	"qiantie/giant-material-executor/internal/update"
)

func TestSignedManifestVerifiesForItsPlatform(t *testing.T) {
	privateKey := ed25519.NewKeyFromSeed([]byte("01234567890123456789012345678901"))
	manifest, err := signedManifest(releaseInput{
		Version: "0.6.0", Platform: "windows", Architecture: "amd64",
		URL:                    "https://updates.example/releases/windows/GiantMaterialExecutor-0.6.0.zip",
		SHA256:                 "45a40dcea5c1a33d9bde6098917f9a452f215301c00e149cc3c8d2932640ab78",
		MinimumExecutorVersion: "0.5.0", PublishedAt: "2026-10-01T00:00:00Z",
	}, privateKey)
	if err != nil {
		t.Fatalf("signedManifest() error = %v", err)
	}
	if manifest.Signature == "" {
		t.Fatal("signed manifest is missing signature")
	}
	if _, err := update.VerifyReleaseManifest(manifest, privateKey.Public().(ed25519.PublicKey), update.ReleaseTarget{Platform: "windows", Architecture: "amd64"}, "0.5.0"); err != nil {
		t.Fatalf("VerifyReleaseManifest() error = %v", err)
	}
	if _, err := base64.StdEncoding.DecodeString(manifest.Signature); err != nil {
		t.Fatalf("signature is not base64: %v", err)
	}
}
