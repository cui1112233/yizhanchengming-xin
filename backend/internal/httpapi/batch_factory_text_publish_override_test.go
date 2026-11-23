package httpapi

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
)

func TestTextPublishOverridesRejectForeignCollectionBeforeStarting(t *testing.T) {
	now := time.Now()
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 2, 0)
	c := preflightCoordinator(t, store, &batchfactoryv11.BookStageService{Store: store})
	allowed := batch.Books[0].ID
	c.textPublishCollection = func(_ context.Context, s txt.Scope, id string) (txt.Collection, error) {
		if s.Owner == "alice" && s.BatchID == batch.ID && s.BookID == allowed && id == "mine" {
			return txt.Collection{ID: id, Status: "ready"}, nil
		}
		return txt.Collection{}, txt.ErrNotFound
	}
	api := http.NewServeMux()
	registerBatchAutomationRoutes(api, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, c, "v12")
	for _, route := range []string{"start", "retry"} {
		for _, example := range []struct{ overrideBook, collection string }{{allowed, "foreign"}, {batch.Books[1].ID, "mine"}} {
			body := map[string]any{"bookIds": []string{allowed}, "textPublishOverrides": map[string]any{example.overrideBook: map[string]any{"collectionId": example.collection, "force": true, "requestKey": "test-confirmation"}}}
			resp := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/automation/"+route, body)
			if resp.Code != http.StatusBadRequest && resp.Code != http.StatusNotFound {
				t.Fatalf("%s: %d %s", route, resp.Code, resp.Body.String())
			}
			if c.activeAutomationJob("alice", batch.ID) != nil {
				t.Fatalf("invalid override started %s", route)
			}
		}
	}
}

func TestSignedTextStartPersistsOnlySelectedUploadOverrides(t *testing.T) {
	now := time.Now
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 3, 0)
	dir := t.TempDir()
	presets, err := NewFileAutomationPresetStore(filepath.Join(dir, "presets.json"), now)
	if err != nil {
		t.Fatal(err)
	}
	preset, err := presets.Create("alice", "文本", map[string]any{"productionMode": "text", "sensitiveMode": "off"})
	if err != nil {
		t.Fatal(err)
	}
	reader := NewFileAutomationStatusReader(filepath.Join(dir, "state.json"))
	c := NewBatchAutomationCoordinator(store, presets, &batchfactoryv11.BookStageService{Store: store}, nil, nil, nil, nil, reader, now)
	c.textPublishCollection = func(_ context.Context, s txt.Scope, id string) (txt.Collection, error) {
		if s.Owner != "alice" || s.BatchID != batch.ID || id != "ready-"+s.BookID {
			return txt.Collection{}, txt.ErrNotFound
		}
		return txt.Collection{ID: id, Status: "ready"}, nil
	}
	api := http.NewServeMux()
	registerBatchAutomationRoutes(api, BridgeAuth{Secret: "secret", Now: now}, c, "v12")
	first, second := batch.Books[0].ID, batch.Books[1].ID
	body := map[string]any{"bookIds": []string{first, second}, "presetId": preset.ID, "productionMode": "text", "runMode": "video_no_submit", "scheduledAt": now().Add(time.Hour).Format(time.RFC3339),
		"textPublishOverrides": map[string]textPublishOverride{first: {CollectionID: "ready-" + first, Force: true, RequestKey: "confirmation-1"}, second: {CollectionID: "ready-" + second}}}
	resp := signedJSONRequest(t, api, now(), "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/automation/start", body)
	if resp.Code != http.StatusCreated {
		t.Fatalf("start %d %s", resp.Code, resp.Body.String())
	}
	replay := signedJSONRequest(t, api, now(), "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/automation/start", body)
	if replay.Code != http.StatusCreated {
		t.Fatalf("repeat same start: %d %s", replay.Code, replay.Body.String())
	}
	recovered := NewBatchAutomationCoordinator(store, presets, &batchfactoryv11.BookStageService{Store: store}, nil, nil, nil, nil, reader, now)
	entries := mapOrEmpty(recovered.activeAutomationJob("alice", batch.ID)["books"])
	if len(entries) != 2 {
		t.Fatalf("selected %d books", len(entries))
	}
	for id, raw := range entries {
		e := mapOrEmpty(raw)
		if e["textPendingPublishCollectionId"] != "ready-"+id || (automationStringValue(e["textPublishIntentId"], "") != "") != (id == first) {
			t.Fatalf("incorrect saved override for %s: %+v", id, e)
		}
	}
}

func TestTextPublishOverridesValidateAllBeforeMutation(t *testing.T) {
	c := &BatchAutomationCoordinator{store: batchfactoryv11.NewMemoryStore(), textPublishCollection: func(_ context.Context, s txt.Scope, id string) (txt.Collection, error) {
		if id != "ready-"+s.BookID {
			return txt.Collection{}, txt.ErrNotFound
		}
		return txt.Collection{ID: id, Status: "ready"}, nil
	}}
	overrides := map[string]textPublishOverride{"a": {CollectionID: "ready-a", Force: true, RequestKey: "confirmation-1"}, "b": {CollectionID: "ready-b"}}
	for _, ids := range [][]string{nil, {"a"}, {"b"}} {
		if _, err := c.textPublishOverrides(context.Background(), "alice", "batch", ids, overrides); !errors.Is(err, batchfactoryv11.ErrInvalid) {
			t.Fatalf("selection %v: %v", ids, err)
		}
	}
	for range 2 {
		if _, err := c.textPublishOverrides(context.Background(), "alice", "batch", []string{"a", "b"}, overrides); err != nil {
			t.Fatal(err)
		}
	}
	for _, change := range []struct {
		owner, book, key string
		want             error
	}{
		{"bob", "a", "confirmation-1", batchfactoryv11.ErrConflict}, {"alice", "b", "confirmation-1", batchfactoryv11.ErrConflict},
		{"alice", "a", "", batchfactoryv11.ErrInvalid}, {"alice", "a", "bad key!", batchfactoryv11.ErrInvalid},
	} {
		o := map[string]textPublishOverride{change.book: {CollectionID: "ready-" + change.book, Force: true, RequestKey: change.key}}
		if _, err := c.textPublishOverrides(context.Background(), change.owner, "batch", []string{change.book}, o); !errors.Is(err, change.want) {
			t.Fatalf("scope %+v: %v", change, err)
		}
	}
}
