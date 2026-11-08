package productionmedia

import "testing"

func TestObjectKeyKeepsStableLogicalReference(t *testing.T) {
	got, err := ObjectKey("production-media", KindMergeVideo, "merge_123.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if got != "production-media/video/merged/merge_123.mp4" {
		t.Fatalf("object key=%q", got)
	}
	if _, err := ObjectKey("production-media", KindVideo, "../secret"); err == nil {
		t.Fatal("expected traversal reference to be rejected")
	}
}

func TestNewTOSStoreRejectsIncompleteConfiguration(t *testing.T) {
	if _, err := NewTOSStore(TOSConfig{}); err == nil {
		t.Fatal("expected incomplete TOS configuration error")
	}
}
