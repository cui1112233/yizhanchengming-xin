package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
)

func TestSignedSelectedTextAutomationUsesOnlyRequestedBook(t *testing.T) {
	now := func() time.Time { return time.Unix(1700000000, 0) }
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 2, 0)
	batch = savePreflightBatchSettings(t, store, batch, batchfactoryv11.SettingsPatch{"productionMode": json.RawMessage(`"text"`)})
	presets, err := NewFileAutomationPresetStore(filepath.Join(t.TempDir(), "presets.json"), now)
	if err != nil {
		t.Fatal(err)
	}
	preset, err := presets.Create("alice", "文本合集", map[string]any{"productionMode": "text", "sensitiveMode": "off"})
	if err != nil {
		t.Fatal(err)
	}
	selected := batch.Books[1].ID
	coord := NewBatchAutomationCoordinator(store, presets, &batchfactoryv11.BookStageService{Store: store}, nil, nil, nil, nil, NewFileAutomationStatusReader(filepath.Join(t.TempDir(), "state.json")), now)
	requests := make(chan txt.EnsureRequest, 2)
	coord.textEnsure = func(_ context.Context, req txt.EnsureRequest) (txt.EnsureResult, error) {
		requests <- req
		return txt.EnsureResult{Stage: "text_collection", Collection: txt.Collection{ID: "selected-collection"}}, nil
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: now, Slice: 2, Store: store, AutomationCoordinator: coord})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/automation/"
	payload := map[string]any{"presetId": preset.ID, "bookIds": []string{selected}, "productionMode": "text", "runMode": "video_no_submit", "scheduleMode": "execute", "autoPublish": false}
	pre := signedJSONRequest(t, api, now(), "alice", http.MethodPost, path+"preflight", payload)
	if pre.Code != http.StatusOK {
		t.Fatalf("preflight %d %s", pre.Code, pre.Body.String())
	}
	check := decodeBody[automationPreflight](t, pre)
	if len(check.Books) != 1 || check.Books[0].BookID != selected || !check.Books[0].Passed {
		t.Fatalf("preflight selection %s", pre.Body.String())
	}
	start := signedJSONRequest(t, api, now(), "alice", http.MethodPost, path+"start", payload)
	if start.Code != http.StatusCreated {
		t.Fatalf("start %d %s", start.Code, start.Body.String())
	}
	job := waitForTerminalAutomationJob(t, coord, "alice", batch.ID)
	books := mapOrEmpty(job["books"])
	if len(books) != 1 || mapOrEmpty(books[selected])["textCollectionId"] != "selected-collection" {
		t.Fatalf("selection %+v", job)
	}
	select {
	case req := <-requests:
		if req.Owner != "alice" || req.BookID != selected || req.Target != "collection" {
			t.Fatalf("text request %+v", req)
		}
	default:
		t.Fatal("text workflow did not receive selected book")
	}
	if len(requests) > 0 {
		t.Fatal("unselected book was dispatched")
	}
}
