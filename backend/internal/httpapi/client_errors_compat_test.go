package httpapi

import (
	"testing"
)

func TestClientErrorLogIsBoundedAndReadable(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 505; i++ {
		if err := appendClientError(dir, map[string]any{"id": i, "username": "tester"}); err != nil {
			t.Fatalf("append error log: %v", err)
		}
	}
	entries := readClientErrors(dir)
	if len(entries) != 500 {
		t.Fatalf("entries = %d, want 500", len(entries))
	}
	if entries[0]["id"] != float64(5) && entries[0]["id"] != 5 {
		t.Fatalf("first retained id = %#v", entries[0]["id"])
	}
}

func TestClientErrorHelpersClipAndLimit(t *testing.T) {
	if got := clipClientError("  abc  ", 3); got != "abc" {
		t.Fatalf("clip = %q", got)
	}
	entries := []map[string]any{{"id": 1}, {"id": 2}}
	if got := limitClientErrors(entries, "1"); len(got) != 1 || got[0]["id"] != 2 {
		t.Fatalf("limited entries = %#v", got)
	}
}
