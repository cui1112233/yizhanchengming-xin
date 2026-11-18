package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTestManifest(t *testing.T, dir, name string, data map[string]any, modified time.Time) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := writeWebSubmitManifest(path, data); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRecoverStaleWebSubmitManifestsMarksExecutingUnknown(t *testing.T) {
	usersDir := t.TempDir()
	dir := filepath.Join(usersDir, "tester", "web-submit", "manifests")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	stale := time.Now().UTC().Add(-10 * time.Minute)
	staleAt := stale.Format(time.RFC3339Nano)
	writeTestManifest(t, dir, "submit-stale.json", map[string]any{
		"manifest_version":  1,
		"submission_run_id": "submit-stale",
		"status":            "executing",
		"created_at":        staleAt,
		"updated_at":        staleAt,
		"items": []any{
			map[string]any{"id": "a", "version": "ai1", "state": "executing"},
			map[string]any{"id": "b", "version": "ai1", "state": "submitted"},
		},
	}, stale)
	recentAt := time.Now().UTC().Format(time.RFC3339Nano)
	writeTestManifest(t, dir, "submit-recent.json", map[string]any{
		"manifest_version":  1,
		"submission_run_id": "submit-recent",
		"status":            "planned",
		"created_at":        recentAt,
		"updated_at":        recentAt,
		"items":             []any{},
	}, time.Now().UTC())

	if got := recoverStaleWebSubmitManifests(usersDir, "tester"); got != 1 {
		t.Fatalf("recovered=%d, want 1", got)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "submit-stale.json"))
	if err != nil {
		t.Fatal(err)
	}
	var recovered struct {
		Status string           `json:"status"`
		Items  []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(raw, &recovered); err != nil {
		t.Fatal(err)
	}
	if recovered.Status != "recovered" || recovered.Items[0]["state"] != "unknown" || recovered.Items[0]["reason"] != "recovered_after_crash" {
		t.Fatalf("stale manifest was not recovered: %s", string(raw))
	}
	if recovered.Items[1]["state"] != "submitted" {
		t.Fatalf("confirmed item must stay submitted: %#v", recovered.Items[1])
	}
	rawRecent, err := os.ReadFile(filepath.Join(dir, "submit-recent.json"))
	if err != nil {
		t.Fatal(err)
	}
	var recent struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rawRecent, &recent); err != nil {
		t.Fatal(err)
	}
	if recent.Status != "planned" {
		t.Fatalf("recent manifest must not be touched: %s", string(rawRecent))
	}
}

func TestListWebSubmitManifestsReturnsLatestFirst(t *testing.T) {
	usersDir := t.TempDir()
	dir := filepath.Join(usersDir, "tester", "web-submit", "manifests")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	old := time.Now().UTC().Add(-2 * time.Hour)
	writeTestManifest(t, dir, "submit-old.json", map[string]any{"submission_run_id": "submit-old", "status": "completed"}, old)
	writeTestManifest(t, dir, "submit-new.json", map[string]any{"submission_run_id": "submit-new", "status": "completed"}, time.Now().UTC())

	runs := listWebSubmitManifests(usersDir, "tester", 10)
	if len(runs) != 2 {
		t.Fatalf("runs=%#v", runs)
	}
	if first, _ := runs[0].(map[string]any); first["submission_run_id"] != "submit-new" {
		t.Fatalf("latest run must be first: %#v", runs)
	}
	if limited := listWebSubmitManifests(usersDir, "tester", 1); len(limited) != 1 {
		t.Fatalf("limit not applied: %#v", limited)
	}
}
