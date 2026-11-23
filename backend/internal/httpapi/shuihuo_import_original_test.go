package httpapi

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestUploadedSourceSurvivesProductionEdit(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	mux := http.NewServeMux()
	now := time.Unix(100, 0)
	registerShuihuoProjectCompatRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, nil, nil, nil, nil, nil, nil, "", nil, nil)
	const full = "  第一段\n第二段\n"
	created := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/shuihuo-production/projects/import", map[string]any{"name": "测试", "filename": "story.txt", "dataUrl": "data:text/plain;base64," + base64.StdEncoding.EncodeToString([]byte(full))})
	if created.Code != http.StatusCreated {
		t.Fatalf("import: %d %s", created.Code, created.Body.String())
	}
	batches, err := store.ListBatches(context.Background(), "alice")
	if err != nil || len(batches) != 1 {
		t.Fatalf("batches: %v %v", batches, err)
	}
	id := batches[0].ID
	edited := signedJSONRequest(t, mux, now, "alice", http.MethodPut, "/api/shuihuo-production/projects/"+id+"/source", map[string]any{"sourceText": "生产稿"})
	if edited.Code != http.StatusOK {
		t.Fatalf("edit: %d %s", edited.Code, edited.Body.String())
	}
	batch, err := store.GetBatch(context.Background(), "alice", id)
	if err != nil || len(batch.Books) != 1 {
		t.Fatalf("book: %v %v", batch.Books, err)
	}
	book := batch.Books[0]
	if book.TxtText != full || book.SourceText != "生产稿" || book.SourceMetadata["rawSourceOrigin"] != "uploaded_file" || book.SourceMetadata["rawSourceStatus"] != "verified" {
		t.Fatalf("edited book: %+v", book)
	}
}
