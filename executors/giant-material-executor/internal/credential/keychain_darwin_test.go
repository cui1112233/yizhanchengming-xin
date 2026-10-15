//go:build darwin

package credential

import "testing"

func TestDeviceStoreUsesCallerCredentialPathOnDarwin(t *testing.T) {
	store := NewDeviceStore(t.TempDir() + "/executor.json")
	if _, ok := store.(*FileStore); !ok {
		t.Fatalf("device store type = %T, want *FileStore so pairing survives a background restart", store)
	}
}
