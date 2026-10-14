package runtimeaudit

import "testing"

func TestRegistryCountsLegacyBridgeHits(t *testing.T) {
	registry := NewRegistry()
	registry.RecordBridge("/api/script-video")
	registry.RecordBridge("/api/script-video")
	registry.RecordBridge("/api/login")

	got := registry.Snapshot()
	if got.BridgeHits != 3 {
		t.Fatalf("BridgeHits = %d, want 3", got.BridgeHits)
	}
	if got.Paths["/api/script-video"] != 2 || got.Paths["/api/login"] != 1 {
		t.Fatalf("Paths = %#v", got.Paths)
	}
}

func TestRegistrySnapshotDoesNotExposeMutableState(t *testing.T) {
	registry := NewRegistry()
	registry.RecordBridge("/api/login")
	first := registry.Snapshot()
	first.Paths["/api/login"] = 99
	second := registry.Snapshot()
	if second.Paths["/api/login"] != 1 {
		t.Fatalf("snapshot mutated registry: %#v", second.Paths)
	}
}
