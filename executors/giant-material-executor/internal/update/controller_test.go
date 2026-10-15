package update

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"testing"
)

func TestControllerCheckReportsAvailableWithoutStagingRelease(t *testing.T) {
	privateKey := ed25519.NewKeyFromSeed([]byte("01234567890123456789012345678901"))
	manifest := signedTestManifest(t, privateKey, "windows", "amd64")
	stageCalls := 0
	controller := NewController(ControllerConfig{
		CurrentVersion: "0.5.0",
		Target:         ReleaseTarget{Platform: "windows", Architecture: "amd64"},
		PublicKey:      privateKey.Public().(ed25519.PublicKey),
		Fetch: func(context.Context) (ReleaseManifest, error) {
			return manifest, nil
		},
		Stage: func(context.Context, ReleaseManifest) error {
			stageCalls++
			return nil
		},
	})

	status, err := controller.Check(context.Background())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if status.State != StateAvailable || status.AvailableVersion != "0.6.0" || !status.CanApply {
		t.Fatalf("status = %#v, want available 0.6.0", status)
	}
	if stageCalls != 0 {
		t.Fatalf("Check() staged %d releases, want 0", stageCalls)
	}
}

func TestControllerApplyLeavesBusyExecutorUntouched(t *testing.T) {
	privateKey := ed25519.NewKeyFromSeed([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	manifest := signedTestManifest(t, privateKey, "macos", "universal")
	stageCalls := 0
	controller := NewController(ControllerConfig{
		CurrentVersion: "0.5.0",
		Target:         ReleaseTarget{Platform: "macos", Architecture: "universal"},
		PublicKey:      privateKey.Public().(ed25519.PublicKey),
		Idle:           func() bool { return false },
		Fetch: func(context.Context) (ReleaseManifest, error) {
			return manifest, nil
		},
		Stage: func(context.Context, ReleaseManifest) error {
			stageCalls++
			return nil
		},
	})

	status, err := controller.Apply(context.Background())
	if CodeOf(err) != "UPDATE_EXECUTOR_BUSY" {
		t.Fatalf("CodeOf(err) = %q, want UPDATE_EXECUTOR_BUSY (err=%v)", CodeOf(err), err)
	}
	if status.State != StateBusy || status.CanApply {
		t.Fatalf("status = %#v, want busy and not applicable", status)
	}
	if stageCalls != 0 {
		t.Fatalf("Apply() staged %d releases while busy, want 0", stageCalls)
	}
}

func TestControllerApplyRechecksIdlenessImmediatelyBeforeStaging(t *testing.T) {
	privateKey := ed25519.NewKeyFromSeed([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"))
	manifest := signedTestManifest(t, privateKey, "windows", "amd64")
	idleChecks := 0
	stageCalls := 0
	controller := NewController(ControllerConfig{
		CurrentVersion: "0.5.0",
		Target:         ReleaseTarget{Platform: "windows", Architecture: "amd64"},
		PublicKey:      privateKey.Public().(ed25519.PublicKey),
		Idle: func() bool {
			idleChecks++
			return idleChecks < 3
		},
		Fetch: func(context.Context) (ReleaseManifest, error) {
			return manifest, nil
		},
		Stage: func(context.Context, ReleaseManifest) error {
			stageCalls++
			return nil
		},
	})

	status, err := controller.Apply(context.Background())
	if CodeOf(err) != "UPDATE_EXECUTOR_BUSY" {
		t.Fatalf("CodeOf(err) = %q, want UPDATE_EXECUTOR_BUSY (err=%v)", CodeOf(err), err)
	}
	if status.State != StateBusy || stageCalls != 0 {
		t.Fatalf("status=%#v stageCalls=%d, want busy without staging", status, stageCalls)
	}
}

func signedTestManifest(t *testing.T, privateKey ed25519.PrivateKey, platform, architecture string) ReleaseManifest {
	t.Helper()
	manifest := ReleaseManifest{
		SchemaVersion:          1,
		Version:                "0.6.0",
		Platform:               platform,
		Architecture:           architecture,
		URL:                    "https://updates.example/releases/test/GiantMaterialExecutor-0.6.0.zip",
		SHA256:                 "45a40dcea5c1a33d9bde6098917f9a452f215301c00e149cc3c8d2932640ab78",
		MinimumExecutorVersion: "0.5.0",
		PublishedAt:            "2026-10-01T00:00:00Z",
	}
	payload, err := manifest.SigningBytes()
	if err != nil {
		t.Fatalf("SigningBytes() error = %v", err)
	}
	manifest.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, payload))
	return manifest
}
