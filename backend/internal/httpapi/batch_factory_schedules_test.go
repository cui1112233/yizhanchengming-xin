package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
)

func TestBatchFactorySchedulesCRUDUsesGoUserStorage(t *testing.T) {
	now := time.Date(2026, 10, 11, 10, 0, 0, 0, time.UTC)
	secret := "secret"
	root := t.TempDir()
	api := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, UserDataDir: filepath.Join(root, "users"), AuthStore: &authTestStore{account: auth.Account{Username: "tester", Active: true, IsOwner: true}}})
	create := signedJSONRequest(t, api, now, "tester", http.MethodPost, "/api/batch-factory/v12/schedules", map[string]any{
		"batchId": "batch-1", "requestId": "request-1", "runAt": now.Add(time.Hour).Format(time.RFC3339),
	})
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}
	var created struct {
		Schedule map[string]any `json:"schedule"`
	}
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	id, _ := created.Schedule["id"].(string)
	if id == "" || created.Schedule["status"] != "scheduled" {
		t.Fatalf("unexpected schedule=%#v", created.Schedule)
	}
	list := signedJSONRequest(t, api, now, "tester", http.MethodGet, "/api/batch-factory/v12/schedules", nil)
	if list.Code != http.StatusOK || !containsJSONID(list.Body.Bytes(), id) {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	patch := signedJSONRequest(t, api, now, "tester", http.MethodPatch, "/api/batch-factory/v12/schedules/"+id, map[string]any{"runAt": now.Add(2 * time.Hour).Format(time.RFC3339)})
	if patch.Code != http.StatusOK || !containsJSONText(patch.Body.Bytes(), now.Add(2*time.Hour).Format(time.RFC3339)) {
		t.Fatalf("patch status=%d body=%s", patch.Code, patch.Body.String())
	}
	remove := signedJSONRequest(t, api, now, "tester", http.MethodDelete, "/api/batch-factory/v12/schedules/"+id, nil)
	if remove.Code != http.StatusOK || !containsJSONText(remove.Body.Bytes(), `"deleted":true`) {
		t.Fatalf("remove status=%d body=%s", remove.Code, remove.Body.String())
	}
}

func containsJSONID(raw []byte, id string) bool     { return containsJSONText(raw, id) }
func containsJSONText(raw []byte, text string) bool { return bytes.Contains(raw, []byte(text)) }
