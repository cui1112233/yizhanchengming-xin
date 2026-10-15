package update

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"
)

func TestVerifyReleaseManifestAcceptsSignedMatchingTarget(t *testing.T) {
	privateKey := ed25519.NewKeyFromSeed([]byte("01234567890123456789012345678901"))
	manifest := ReleaseManifest{
		SchemaVersion:          1,
		Version:                "0.6.0",
		Platform:               "windows",
		Architecture:           "amd64",
		URL:                    "https://updates.example/releases/windows/GiantMaterialExecutor-0.6.0.zip",
		SHA256:                 "45a40dcea5c1a33d9bde6098917f9a452f215301c00e149cc3c8d2932640ab78",
		MinimumExecutorVersion: "0.5.0",
		PublishedAt:            "2026-10-01T00:00:00Z",
	}
	signed, err := manifest.SigningBytes()
	if err != nil {
		t.Fatalf("SigningBytes() error = %v", err)
	}
	manifest.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, signed))

	verified, err := VerifyReleaseManifest(manifest, privateKey.Public().(ed25519.PublicKey), ReleaseTarget{Platform: "windows", Architecture: "amd64"}, "0.5.0")
	if err != nil {
		t.Fatalf("VerifyReleaseManifest() error = %v", err)
	}
	if verified.Version != "0.6.0" {
		t.Fatalf("verified version = %q, want 0.6.0", verified.Version)
	}
}

func TestVerifyReleaseManifestRejectsTamperedTarget(t *testing.T) {
	privateKey := ed25519.NewKeyFromSeed([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	manifest := ReleaseManifest{
		SchemaVersion:          1,
		Version:                "0.6.0",
		Platform:               "macos",
		Architecture:           "universal",
		URL:                    "https://updates.example/releases/macos/GiantMaterialExecutor-0.6.0.zip",
		SHA256:                 "45a40dcea5c1a33d9bde6098917f9a452f215301c00e149cc3c8d2932640ab78",
		MinimumExecutorVersion: "0.5.0",
		PublishedAt:            "2026-10-01T00:00:00Z",
	}
	signed, err := manifest.SigningBytes()
	if err != nil {
		t.Fatalf("SigningBytes() error = %v", err)
	}
	manifest.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, signed))
	manifest.URL = "https://attacker.example/other.zip"

	_, err = VerifyReleaseManifest(manifest, privateKey.Public().(ed25519.PublicKey), ReleaseTarget{Platform: "macos", Architecture: "universal"}, "0.5.0")
	if CodeOf(err) != "UPDATE_MANIFEST_SIGNATURE_INVALID" {
		t.Fatalf("CodeOf(err) = %q, want UPDATE_MANIFEST_SIGNATURE_INVALID (err=%v)", CodeOf(err), err)
	}
}

func TestVerifyReleaseManifestRejectsDifferentPlatform(t *testing.T) {
	privateKey := ed25519.NewKeyFromSeed([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"))
	manifest := ReleaseManifest{
		SchemaVersion:          1,
		Version:                "0.6.0",
		Platform:               "windows",
		Architecture:           "amd64",
		URL:                    "https://updates.example/releases/windows/GiantMaterialExecutor-0.6.0.zip",
		SHA256:                 "45a40dcea5c1a33d9bde6098917f9a452f215301c00e149cc3c8d2932640ab78",
		MinimumExecutorVersion: "0.5.0",
		PublishedAt:            "2026-10-01T00:00:00Z",
	}
	signed, err := manifest.SigningBytes()
	if err != nil {
		t.Fatalf("SigningBytes() error = %v", err)
	}
	manifest.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, signed))

	_, err = VerifyReleaseManifest(manifest, privateKey.Public().(ed25519.PublicKey), ReleaseTarget{Platform: "macos", Architecture: "universal"}, "0.5.0")
	if CodeOf(err) != "UPDATE_PLATFORM_MISMATCH" {
		t.Fatalf("CodeOf(err) = %q, want UPDATE_PLATFORM_MISMATCH (err=%v)", CodeOf(err), err)
	}
}

func TestPublicKeyFromBase64RejectsInvalidKey(t *testing.T) {
	_, err := PublicKeyFromBase64("not-a-public-key")
	if CodeOf(err) != "UPDATE_MANIFEST_SIGNATURE_INVALID" {
		t.Fatalf("CodeOf(err) = %q, want UPDATE_MANIFEST_SIGNATURE_INVALID (err=%v)", CodeOf(err), err)
	}
}
