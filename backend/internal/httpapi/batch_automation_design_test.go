package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"qiantie/backend/internal/batchfactoryv11"
	"testing"
	"time"
)

func TestRuntimeSummaryMissingComicPagesHaveZeroActiveTasks(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "未启动的漫画", Books: []batchfactoryv11.CreateBookInput{{Title: "书", SourceText: "正文", Videos: []batchfactoryv11.CreateVideoInput{{Label: "页1", VideoPrompt: "画面1"}, {Label: "页2", VideoPrompt: "画面2"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Store: store})
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/runtime-summary", nil)
	var result struct {
		StageSummaries map[string]map[string]any `json:"stageSummaries"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	summary := result.StageSummaries[batch.Books[0].ID]
	if response.Code != 200 || summary["comicImageTotalCount"] != float64(2) || summary["comicImageGeneratingCount"] != float64(0) || summary["comicImageFailedCount"] != float64(0) {
		t.Fatalf("missing pages became active: status=%d summary=%+v", response.Code, summary)
	}
}

func TestComicActivityUsesActualTasksNotMissingImages(t *testing.T) {
	now := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	got := comicPageActivityCounts(nil, now)
	if got.Generating != 0 || got.Failed != 0 || got.Interrupted != 0 {
		t.Fatalf("missing images cannot imply a task: %+v", got)
	}
	got = comicPageActivityCounts([]batchfactoryv11.ComicPage{{Status: "ready"}, {Status: "generating", UpdatedAt: now.Add(-time.Minute)}, {Status: "generation_failed"}, {Status: "generating", UpdatedAt: now.Add(-comicPageGenerationStaleAfter - time.Minute)}}, now)
	if got.Generating != 1 || got.Failed != 1 || got.Interrupted != 1 {
		t.Fatalf("activity %+v", got)
	}
}

func TestEndAutomationStopsNeedsAttentionAndPreservesCompletedProgress(t *testing.T) {
	path := filepath.Join(t.TempDir(), "automation.json")
	raw := `{"jobs":{"job-1":{"id":"job-1","owner":"alice","batchId":"batch-1","state":"needs_attention","books":{"done":{"bookId":"done","status":"completed","stage":"published"},"pending":{"bookId":"pending","status":"queued","stage":"assets"},"working":{"bookId":"working","status":"running","stage":"comic_images"}}}}}`
	if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	coordinator := NewBatchAutomationCoordinator(batchfactoryv11.NewMemoryStore(), nil, nil, nil, nil, nil, nil, NewFileAutomationStatusReader(path), nil)
	if err := coordinator.Cancel("alice", "batch-1"); err != nil {
		t.Fatal(err)
	}
	state, err := coordinator.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	job := state.Jobs["job-1"]
	counts := mapOrEmpty(job["counts"])
	books := mapOrEmpty(job["books"])
	if job["state"] != "stopped" || automationNumberValue(counts["ready"], 0) != 1 || automationNumberValue(counts["total"], 0) != 3 || automationNumberValue(counts["running"], 0) != 0 || mapOrEmpty(books["working"])["status"] != "stopped" {
		t.Fatalf("end lost progress or stayed running: %+v", job)
	}
}

func TestPauseAlreadyPausedDoesNotPretendToPauseAgain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "automation.json")
	os.WriteFile(path, []byte(`{"jobs":{"job-1":{"id":"job-1","owner":"alice","batchId":"batch-1","state":"paused","books":{}}}}`), 0600)
	coordinator := NewBatchAutomationCoordinator(batchfactoryv11.NewMemoryStore(), nil, nil, nil, nil, nil, nil, NewFileAutomationStatusReader(path), nil)
	if _, err := coordinator.Pause("alice", "batch-1"); !errors.Is(err, batchfactoryv11.ErrConflict) {
		t.Fatalf("pause err=%v", err)
	}
}
