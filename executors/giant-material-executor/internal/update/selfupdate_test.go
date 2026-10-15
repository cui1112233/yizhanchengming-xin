package update

import (
	"context"
	"crypto/ed25519"
	"errors"
	"testing"
)

func TestSelfUpdaterApplyNowReturnsHelperLaunchFailure(t *testing.T) {
	privateKey := ed25519.NewKeyFromSeed([]byte("01234567890123456789012345678901"))
	manifest := signedTestManifest(t, privateKey, "windows", "amd64")
	want := errors.New("update helper is missing")
	updater := &SelfUpdater{
		CurrentVersion: "0.5.0",
		PublicKey:      privateKey.Public().(ed25519.PublicKey),
		Target:         ReleaseTarget{Platform: "windows", Architecture: "amd64"},
		Apply:          func(string) error { return want },
		controller: NewController(ControllerConfig{
			CurrentVersion: "0.5.0",
			Target:         ReleaseTarget{Platform: "windows", Architecture: "amd64"},
			PublicKey:      privateKey.Public().(ed25519.PublicKey),
			Fetch:          func(context.Context) (ReleaseManifest, error) { return manifest, nil },
			Stage:          func(context.Context, ReleaseManifest) error { return nil },
		}),
	}
	updater.once.Do(func() {})

	if _, err := updater.ApplyNow(context.Background()); !errors.Is(err, want) {
		t.Fatalf("ApplyNow() error = %v, want helper launch failure", err)
	}
}
