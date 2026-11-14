package httpapi

import (
	"context"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestDeferredScheduleFollowsRetriedBooksThroughConfirmedUpload(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	now := time.Now().UTC()
	next := now.Add(time.Hour).Format(time.RFC3339Nano)
	item := map[string]any{"id": "original-schedule", "status": "waiting_retry", "job_id": "original-job", "book_ids": []any{"12345678901"}, "auto_submit": true, "run_at": now.Format(time.RFC3339Nano)}
	if err := store.PutConfig(ctx, "alice", map[string]any{"novel_fetch_schedules": []any{item}}); err != nil {
		t.Fatal(err)
	}
	doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"status": "waiting_retry", "error": "upstream unavailable", "originalFetchNextRetryAt": next}}
	for _, step := range []struct{ task, submit, want string }{
		{"waiting_retry", "", "waiting_retry"},
		{"processing", "", "running"},
		{"completed", "", "upload_pending"},
		{"completed", "submitted", "completed"},
	} {
		doc.Meta["status"], doc.Meta["siteSubmitStatus"] = step.task, step.submit
		if err := store.PutDocument(ctx, "alice", doc); err != nil {
			t.Fatal(err)
		}
		if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "alice", now); err != nil {
			t.Fatal(err)
		}
		settings, _ := store.GetConfig(ctx, "alice")
		row := mapOrEmpty(arrayOrEmpty(settings["novel_fetch_schedules"])[0])
		if row["status"] != step.want || row["job_id"] != "original-job" {
			t.Fatalf("step %s: %#v", step.task, row)
		}
		if step.task == "waiting_retry" && (row["next_retry_at"] != next || row["error"] != "upstream unavailable") {
			t.Fatalf("retry details missing: %#v", row)
		}
		if step.want == "completed" && (firstText(row["error"]) != "" || firstText(row["next_retry_at"]) != "") {
			t.Fatalf("stale failure after success: %#v", row)
		}
	}
	if _, err := queue.Claim(ctx, time.Millisecond); err == nil {
		t.Fatal("status reconciliation redispatched an existing schedule")
	}
}

func TestDeferredScheduleDoesNotUseAnotherOwnersCompletedBook(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	item := map[string]any{"id": "alice-schedule", "status": "waiting_retry", "book_ids": []any{"12345678901"}, "auto_submit": false}
	_ = store.PutConfig(ctx, "alice", map[string]any{"novel_fetch_schedules": []any{item}})
	_ = store.PutDocument(ctx, "bob", novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"status": "completed"}})
	if _, err := reconcileNovelFetchScheduleRetry(ctx, store, "alice", item, nil, time.Now()); err != nil {
		t.Fatal(err)
	}
	settings, _ := store.GetConfig(ctx, "alice")
	row := mapOrEmpty(arrayOrEmpty(settings["novel_fetch_schedules"])[0])
	if row["status"] != "failed" {
		t.Fatalf("another owner's book completed Alice's schedule: %#v", row)
	}
}

func TestDeferredScheduleTracksAutomaticFailureUntilManualBudgetExhaustion(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	settings := map[string]any{"novel_fetch_automation": map[string]any{"enabled": true}}
	for _, manualOnly := range []bool{false, true} {
		item := map[string]any{"id": "s1", "status": "waiting_retry", "book_ids": []any{"12345678901"}, "auto_submit": false}
		settings["novel_fetch_schedules"] = []any{item}
		_ = store.PutConfig(ctx, "alice", settings)
		_ = store.PutDocument(ctx, "alice", novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"status": "failed", "manualOnly": manualOnly}})
		if _, err := reconcileNovelFetchScheduleRetry(ctx, store, "alice", item, settings, time.Now()); err != nil {
			t.Fatal(err)
		}
		current, _ := store.GetConfig(ctx, "alice")
		want := "waiting_retry"
		if manualOnly {
			want = "failed"
		}
		if row := mapOrEmpty(arrayOrEmpty(current["novel_fetch_schedules"])[0]); row["status"] != want {
			t.Fatalf("manualOnly=%v row=%#v", manualOnly, row)
		}
	}
}

func TestDeferredScheduleMergeDoesNotOverwriteConcurrentCompletion(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	_ = store.PutConfig(ctx, "alice", map[string]any{"novel_fetch_schedules": []any{map[string]any{"id": "s1", "status": "completed"}}})
	if err := saveNovelFetchScheduleItem(ctx, store, "alice", map[string]any{"id": "s1", "status": "running", "_expected_status": "waiting_retry", "retry_tracking": true}); err != nil {
		t.Fatal(err)
	}
	settings, _ := store.GetConfig(ctx, "alice")
	if row := mapOrEmpty(arrayOrEmpty(settings["novel_fetch_schedules"])[0]); row["status"] != "completed" {
		t.Fatalf("completion was overwritten: %#v", row)
	}
}
