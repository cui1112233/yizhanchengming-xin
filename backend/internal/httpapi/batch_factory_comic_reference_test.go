package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
)

func TestBatchComicReferenceUploadCanBeReadByOwnersExecutor(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本"}}})
	if err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	executors := localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })
	pairing, err := executors.CreatePairing(context.Background(), "alice", localexecutor.PlatformDoubao)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := executors.Pair(context.Background(), localexecutor.PairInput{Code: pairing.Code, Platform: localexecutor.PlatformDoubao, DeviceName: "comic-reference", OS: "windows", Version: "test"})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterBatchComicReferenceRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, files, executors)
	png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	upload := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/comic-reference/upload", map[string]any{"dataUrl": "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)})
	if upload.Code != http.StatusCreated {
		t.Fatalf("upload status=%d body=%s", upload.Code, upload.Body.String())
	}
	var result struct {
		Reference batchfactoryv11.ComicGlobalReferenceImage `json:"reference"`
	}
	if err := json.NewDecoder(upload.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	_, err = store.SaveSettings(context.Background(), "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: batchfactoryv11.SettingsPatch{
		"productionMode": comicTestRaw(t, "comic"),
		"aiPromptConfig": comicTestRaw(t, map[string]any{"comic": map[string]any{"enabled": true, "globalReferenceImage": result.Reference}}),
	}, ExpectedRevision: batch.Revision})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, result.Reference.URL, nil)
	request.Header.Set("Authorization", "Bearer "+paired.Token)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("executor read status=%d type=%q body=%s", response.Code, response.Header().Get("Content-Type"), response.Body.String())
	}
	if response.Body.Len() != len(png) {
		t.Fatalf("executor read bytes=%d want=%d", response.Body.Len(), len(png))
	}
}

func TestBatchComicStylePreviewCanBeReadByPairedExecutor(t *testing.T) {
	now := time.Unix(1700000000, 0)
	executors := localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })
	pairing, _ := executors.CreatePairing(context.Background(), "alice", localexecutor.PlatformDoubao)
	paired, err := executors.Pair(context.Background(), localexecutor.PairInput{Code: pairing.Code, Platform: localexecutor.PlatformDoubao, DeviceName: "comic-style", OS: "windows", Version: "test"})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterBatchComicReferenceRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, batchfactoryv11.NewMemoryStore(), localartifact.NewStore(t.TempDir(), 1<<20), executors)
	request := httptest.NewRequest(http.MethodGet, "/api/local-executor/v1/batch-comic-style-previews/batch-comic-style-01", nil)
	request.Header.Set("Authorization", "Bearer "+paired.Token)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "image/jpeg" || response.Body.Len() < 1000 {
		t.Fatalf("style preview status=%d type=%q bytes=%d", response.Code, response.Header().Get("Content-Type"), response.Body.Len())
	}
}
