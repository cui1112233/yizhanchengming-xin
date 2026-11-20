package novelcreation

import (
	"context"
	"encoding/json"
	"testing"
)

// Existing piece-protocol regression fixtures must keep their original request
// contracts. New whole-draft fixtures live in short_full_pipeline_test.go.
func freezeLegacyShortFullDraft(t *testing.T, service *Service, owner, runID string) {
	t.Helper()
	switch store := service.store.(type) {
	case *MemoryStore:
		store.mu.Lock()
		defer store.mu.Unlock()
		run := store.runs[memoryKey(owner, runID)]
		var snapshot map[string]any
		if err := json.Unmarshal(run.PromptSnapshot, &snapshot); err != nil {
			t.Fatal(err)
		}
		snapshot["short_full_draft_version"] = 0
		raw, err := json.Marshal(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		run.PromptSnapshot = raw
		store.runs[memoryKey(owner, runID)] = run
	case *MySQLStore:
		if _, err := store.db.ExecContext(context.Background(), `UPDATE novel_runs SET prompt_snapshot_json=JSON_SET(prompt_snapshot_json,'$.short_full_draft_version',0) WHERE owner_username=? AND run_id=?`, owner, runID); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatal("unsupported legacy short-draft test store")
	}
}
