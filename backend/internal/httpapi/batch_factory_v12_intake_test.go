package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestV12NovelFetchIntakeUsesGoBatchStore(t *testing.T) {
	const secret = "secret"
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	api := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, Slice: 1, Store: store})
	input := map[string]any{
		"books":    []map[string]any{{"id": "book-1", "bookId": "book-1", "sourceTaskId": "task-1", "title": "测试小说", "sourceText": "正文"}},
		"metadata": map[string]any{"source": "novel-fetch"},
	}
	created := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/intakes/novel-fetch", input)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	var payload struct {
		Intake batchfactoryv11.Intake `json:"intake"`
	}
	if err := json.NewDecoder(bytes.NewReader(created.Body.Bytes())).Decode(&payload); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if payload.Intake.ID == "" {
		t.Fatal("Go V12 intake response did not include an id")
	}
	read := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/intakes/"+payload.Intake.ID, nil)
	if read.Code != http.StatusOK {
		t.Fatalf("read status=%d body=%s", read.Code, read.Body.String())
	}
}
