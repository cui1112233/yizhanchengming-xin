package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"qiantie/backend/internal/novelfetchworkshop"
)

func scheduleControlRequest(store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, owner, method, path string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	registerNovelFetchLegacyWorkflowRoutes(mux, store, queue, "", nil)
	req := httptest.NewRequest(method, path, nil)
	req = req.WithContext(context.WithValue(req.Context(), bridgeIdentityKey{}, BridgeIdentity{Username: owner}))
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	return res
}

func TestWorkshopScheduleCancelCleansOnlyExplicitOwnedCreatedBooks(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	store.PutConfig(ctx, "alice", map[string]any{"novel_fetch_schedules": []any{map[string]any{"id": "s1", "status": "scheduled", "created_ids": []string{"owned", "existing", "foreign"}}}})
	for _, row := range []struct{ owner, id, schedule string }{{"alice", "owned", "s1"}, {"alice", "existing", "other"}, {"bob", "foreign", "s1"}} {
		store.PutDocument(ctx, row.owner, novelfetchworkshop.Document{BookID: row.id, Meta: map[string]any{"schedule_id": row.schedule}, Original: "saved"})
	}
	res := scheduleControlRequest(store, queue, "alice", http.MethodPost, "/api/novel-fetch-workshop/schedules/s1/cancel")
	if res.Code != http.StatusOK {
		t.Fatalf("cancel: %d %s", res.Code, res.Body.String())
	}
	if _, err := store.GetDocument(ctx, "alice", "owned"); err == nil {
		t.Fatal("owned schedule book not cleaned")
	}
	for _, row := range []struct{ owner, id string }{{"alice", "existing"}, {"bob", "foreign"}} {
		if _, err := store.GetDocument(ctx, row.owner, row.id); err != nil {
			t.Fatalf("unrelated book removed: %s/%s", row.owner, row.id)
		}
	}
	cfg, _ := store.GetConfig(ctx, "alice")
	if cfg["novel_fetch_schedules"].([]any)[0].(map[string]any)["status"] != "cancelled" {
		t.Fatal("cancel audit record missing")
	}
}

func TestWorkshopScheduleDeleteRequiresTerminalAndPreservesBooks(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	store.PutConfig(ctx, "alice", map[string]any{"novel_fetch_schedules": []any{map[string]any{"id": "s1", "status": "scheduled"}}})
	store.PutDocument(ctx, "alice", novelfetchworkshop.Document{BookID: "saved", Original: "body"})
	res := scheduleControlRequest(store, queue, "alice", http.MethodDelete, "/api/novel-fetch-workshop/schedules/s1")
	if res.Code != http.StatusConflict {
		t.Fatalf("active schedule deleted: %d", res.Code)
	}
	scheduleControlRequest(store, queue, "alice", http.MethodPost, "/api/novel-fetch-workshop/schedules/s1/cancel")
	res = scheduleControlRequest(store, queue, "alice", http.MethodDelete, "/api/novel-fetch-workshop/schedules/s1")
	if res.Code != http.StatusOK {
		t.Fatalf("delete cancelled record: %d %s", res.Code, res.Body.String())
	}
	if _, err := store.GetDocument(ctx, "alice", "saved"); err != nil {
		t.Fatal("delete record deleted body")
	}
	res = scheduleControlRequest(store, queue, "bob", http.MethodPost, "/api/novel-fetch-workshop/schedules/s1/cancel")
	if res.Code != http.StatusNotFound {
		t.Fatalf("cross-owner schedule accessible: %d", res.Code)
	}
}

func TestWorkshopScheduleRunningCancelSignalsQueueWithoutDeletingOutput(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	store.PutConfig(ctx, "alice", map[string]any{"novel_fetch_schedules": []any{map[string]any{"id": "s1", "status": "running", "book_ids": []string{"running"}, "created_ids": []string{"running"}}}})
	store.PutDocument(ctx, "alice", novelfetchworkshop.Document{BookID: "running", Original: "completed output", Meta: map[string]any{"schedule_id": "s1"}})
	res := scheduleControlRequest(store, queue, "alice", http.MethodPost, "/api/novel-fetch-workshop/schedules/s1/cancel")
	if res.Code != http.StatusOK {
		t.Fatalf("cancel: %d %s", res.Code, res.Body.String())
	}
	cancelled, _ := queue.IsBookCancelled(ctx, "alice", "running")
	if !cancelled {
		t.Fatal("queue not signalled")
	}
	if _, err := store.GetDocument(ctx, "alice", "running"); err != nil {
		t.Fatal("active output deleted")
	}
}
