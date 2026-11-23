package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestRouterKeepsShuihuoVideoSubmissionsIndependent(t *testing.T) {
	for _, path := range []string{"/tasks", "/tasks/batch"} {
		t.Run(path, func(t *testing.T) {
			ctx := context.Background()
			store := batchfactoryv11.NewMemoryStore()
			batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "Project", InitialBatchSettings: batchfactoryv11.SettingsPatch{"shuihuoProject": mustJSON(true)}, Books: []batchfactoryv11.CreateBookInput{{Title: "book", SourceText: "源稿。"}}})
			if err != nil {
				t.Fatal(err)
			}
			director := &batchfactoryv11.DirectorService{Store: store, Provider: &directorHTTPProvider{output: directorHTTPJSON}}
			if _, err = director.RunDirector(ctx, "alice", batch.ID, batch.Books[0].ID); err != nil {
				t.Fatal(err)
			}
			adapter := &productionHTTPAdapter{}
			production := &batchfactoryv11.ProductionService{Store: store, Compiler: &batchfactoryv11.PromptCompilerService{Store: store}, Adapter: adapter, Enabled: true, Model: batchfactoryv11.FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}
			now := time.Unix(1700000000, 0)
			api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 4, Store: store, Director: director, Production: production})
			result := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/projects/"+batch.ID+path, map[string]any{"requestId": "shuihuo", "kind": "video"})
			if result.Code != 201 || adapter.calls != 1 || !strings.Contains(result.Body.String(), "provider-http-1") {
				t.Fatalf("status=%d calls=%d body=%s", result.Code, adapter.calls, result.Body.String())
			}
			ordinary, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "ordinary", Books: []batchfactoryv11.CreateBookInput{{Title: "book", SourceText: "正文。"}}})
			if err != nil {
				t.Fatal(err)
			}
			rejected := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/projects/"+ordinary.ID+path, map[string]any{"requestId": "ordinary", "kind": "video"})
			if rejected.Code != http.StatusNotFound || adapter.calls != 1 {
				t.Fatalf("ordinary batch entered Shuihuo production: %d calls=%d", rejected.Code, adapter.calls)
			}
			updated, err := store.GetBatch(ctx, "alice", batch.ID)
			if err != nil {
				t.Fatal(err)
			}
			if updated.Books[0].SourceMetadata["batchSensitiveStage"] != nil {
				t.Fatal("unexpected batch-factory sensitive record on Shuihuo project")
			}
		})
	}
}

// Keep a serialized job snapshot, as the real state store does. Signal only
// when the worker has persisted the terminal result of this retry.
type videoRetryStateStore struct {
	mu       sync.Mutex
	snapshot []byte
	finished chan map[string]any
}

func (s *videoRetryStateStore) Load(context.Context) (automationStateFile, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var state automationStateFile
	err := json.Unmarshal(s.snapshot, &state)
	return state, true, err
}
func (s *videoRetryStateStore) Save(_ context.Context, state automationStateFile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot = mustJSON(state)
	for _, job := range state.Jobs {
		if automationNumberValue(job["retryAttempt"], 0) > 0 && (job["state"] == "failed" || job["state"] == "completed" || job["state"] == "needs_attention") {
			var frozen map[string]any
			if err := json.Unmarshal(mustJSON(job), &frozen); err != nil {
				return err
			}
			select {
			case s.finished <- frozen:
			default:
			}
		}
	}
	return nil
}

func TestRouterAutomationVideoRetryVerifiesStoryboardVersion(t *testing.T) {
	for _, tc := range []struct {
		name            string
		stale, injected bool
	}{
		{"stale", true, false}, {"fresh", false, false}, {"injected-stale", true, true}, {"injected-fresh", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store := batchfactoryv11.NewMemoryStore()
			settings := batchfactoryv11.SettingsPatch{"textModelId": mustJSON("text-1"), "videoModelId": mustJSON("video-model-a"), "assetImagesEnabled": mustJSON(false)}
			batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "batch", InitialBatchSettings: settings, Books: []batchfactoryv11.CreateBookInput{{Title: "book", TxtText: "林晚推门进入客厅。", SourceText: "林晚推门进入客厅。", SourceMetadata: map[string]any{"rawSourceStatus": "verified"}}}})
			if err != nil {
				t.Fatal(err)
			}
			book := batch.Books[0]
			audio := batchfactoryv11.SettingsPatch{"audioPlanningEnabled": mustJSON(true), "audioLineDurations": mustJSON([]float64{9}), "audioDurationSeconds": mustJSON(9), "ttsAudioArtifactId": mustJSON("tts-test")}
			if _, err := store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBook, BatchID: batch.ID, BookID: book.ID}, batchfactoryv11.SettingsUpdate{Patch: audio, ExpectedRevision: book.Revision}); err != nil {
				t.Fatal(err)
			}
			files := localartifact.NewStore(t.TempDir(), 20<<20)
			if _, err := files.SaveMedia("tts-test", "audio/mpeg", "mp3", bytes.NewBufferString("ID3")); err != nil {
				t.Fatal(err)
			}
			catalog := novelfetchworkshop.NewMemoryStore()
			if err := catalog.PutConfig(ctx, "alice", map[string]any{"modelCatalog": []any{map[string]any{"id": "video-model-a", "modelId": "video-model-a", "kind": "video", "enabled": true, "baseUrl": "https://video.invalid", "credential": "fake"}}}); err != nil {
				t.Fatal(err)
			}
			coordinator := &BatchAutomationCoordinator{store: store, ttsStore: catalog}
			if _, err := coordinator.runBatchSensitiveStage(ctx, "alice", batch.ID, book.ID, false); err != nil {
				t.Fatal(err)
			}
			director := &batchfactoryv11.DirectorService{Store: store, Provider: &directorHTTPProvider{output: directorHTTPJSON}}
			if _, err = director.RunDirector(ctx, "alice", batch.ID, book.ID); err != nil {
				t.Fatal(err)
			}
			if tc.stale {
				if _, err = store.SaveDraft(ctx, "alice", batchfactoryv11.Draft{Key: "working-front:" + book.ID, Kind: "working-front-content", Scope: batch.ID, Content: "新的正文。"}); err != nil {
					t.Fatal(err)
				}
			}
			adapter := &productionHTTPAdapter{}
			production := &batchfactoryv11.ProductionService{Store: store, Compiler: &batchfactoryv11.PromptCompilerService{Store: store}, Adapter: adapter, Enabled: true, Model: batchfactoryv11.FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}
			jobID := "retry-video"
			job := map[string]any{"id": jobID, "owner": "alice", "batchId": batch.ID, "state": "failed", "runMode": "video_no_submit", "books": map[string]any{book.ID: map[string]any{"bookId": book.ID, "title": "book", "status": "failed", "stage": "video"}}}
			retryStore := &videoRetryStateStore{snapshot: mustJSON(automationStateFile{Version: 1, Jobs: map[string]map[string]any{jobID: job}}), finished: make(chan map[string]any, 1)}
			reader := NewFileAutomationStatusReader("")
			now := time.Unix(1700000000, 0)
			options := RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 4, Store: store, Director: director, Production: production, AutomationStatusReader: reader, AutomationStateStore: retryStore, LocalArtifacts: files, NovelFetchStore: catalog}
			if tc.injected {
				stages := &batchfactoryv11.BookStageService{Store: store, Director: director, Production: production, VideoReady: shuihuoVideoReadyChecker(store, catalog, production)}
				options.AutomationCoordinator = NewBatchAutomationCoordinator(store, nil, stages, nil, nil, catalog, nil, reader, options.Now)
				options.AutomationCoordinator.SetLocalFiles(files)
				options.Production = nil
			}
			api := NewRouter(options)
			resp := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/automation/retry", map[string]any{"bookIds": []string{book.ID}})
			if resp.Code < 200 || resp.Code >= 300 {
				t.Fatalf("retry %d %s", resp.Code, resp.Body.String())
			}
			var final map[string]any
			select {
			case final = <-retryStore.finished:
			case <-time.After(5 * time.Second):
				t.Fatal("retry did not finish")
			}
			jobs, err := store.ListProductionJobs(ctx, "alice", batch.ID)
			if err != nil {
				t.Fatal(err)
			}
			if tc.stale {
				if len(jobs) != 0 || adapter.calls != 0 || !strings.Contains(string(mustJSON(final)), "旧版生产正文") {
					t.Fatalf("jobs=%d submits=%d final=%s", len(jobs), adapter.calls, mustJSON(final))
				}
			} else if final["state"] != "completed" || len(jobs) != 1 || adapter.calls != 1 {
				t.Fatalf("jobs=%d submits=%d final=%s", len(jobs), adapter.calls, mustJSON(final))
			}
		})
	}
}
