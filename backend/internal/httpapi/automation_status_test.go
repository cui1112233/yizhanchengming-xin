package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileAutomationStatusReaderMatchesNodePublicShape(t *testing.T) {
	path := filepath.Join(t.TempDir(), "automation.json")
	raw := `{"version":1,"jobs":{"owner%3Aalice":{"owner":"alice","batchId":"batch-1","id":"job-1","state":"running","runMode":"video_no_submit","counts":{"total":1,"running":1},"books":{"book-1":{"bookId":"book-1","title":"第一本","status":"running","stage":"video","message":"执行中","error":"","attempts":1,"retryCount":0,"retryRequested":false},"secret":{"apiKey":"must-not-leak"}}}}}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	reader := NewFileAutomationStatusReader(path)
	status, err := reader.Status("alice", "batch-1")
	if err != nil || status["id"] != "job-1" {
		t.Fatalf("status=%+v err=%v", status, err)
	}
	books, ok := status["books"].([]any)
	if !ok || len(books) != 2 {
		t.Fatalf("books=%#v", status["books"])
	}
	if strings.Contains(string(mustJSON(status)), "apiKey") {
		t.Fatal("private runtime fields leaked into public automation status")
	}
}

func TestV12AutomationStatusRouteUsesPersistedState(t *testing.T) {
	now := time.Unix(1700000000, 0)
	path := filepath.Join(t.TempDir(), "automation.json")
	if err := os.WriteFile(path, []byte(`{"jobs":{"job-key":{"owner":"alice","batchId":"batch-1","id":"job-1","state":"completed","books":{}}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AutomationStatusReader: NewFileAutomationStatusReader(path)})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/batch-1/automation", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"job-1"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func mustJSON(value any) []byte {
	encoded, _ := json.Marshal(value)
	return encoded
}
