package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestSliceThreeExposesReadOnlyEffectiveSettingsAndFinalPrompt(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, _ := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "林晚推门进入客厅。"}}})
	provider := &directorHTTPProvider{output: directorHTTPJSON}
	director := &batchfactoryv11.DirectorService{Store: store, Provider: provider}
	revision, err := director.RunDirector(context.Background(), "alice", batch.ID, batch.Books[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	compiler := &batchfactoryv11.PromptCompilerService{Store: store}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: store, Director: director, Compiler: compiler})
	base := "/api/batch-factory/v11/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/videos/" + revision.Videos[0].ID
	for _, suffix := range []string{"/effective-settings", "/final-prompt"} {
		rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, base+suffix, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", suffix, rec.Code, rec.Body.String())
		}
	}
	v12Base := strings.Replace(base, "/v11/", "/v12/", 1)
	for _, suffix := range []string{"/effective-settings", "/final-prompt"} {
		rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, v12Base+suffix, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("v12 %s status=%d body=%s", suffix, rec.Code, rec.Body.String())
		}
	}
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, base+"/final-prompt", nil)
	if !strings.Contains(rec.Body.String(), "compiledPrompt") || !strings.Contains(rec.Body.String(), revision.ID) {
		t.Fatalf("body=%s", rec.Body.String())
	}
}

func TestSliceTwoDoesNotExposeCompilerPreview(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 2, Store: store, Compiler: &batchfactoryv11.PromptCompilerService{Store: store}})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v11/batches/x/books/y/videos/z/final-prompt", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
