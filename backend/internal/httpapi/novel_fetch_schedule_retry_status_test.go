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
		if step.submit == "submitted" {
			doc.Meta["siteSubmitDoneVersions"] = []string{"original"}
		}
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

func TestDeferredGenerateOnlyScheduleCompletesAcrossStaleUploadStates(t *testing.T) {
	for _, status := range []string{"upload_pending", "pending_upload", "pending_verification", "pending_confirmation", "submitted"} {
		t.Run(status, func(t *testing.T) {
			ctx := context.Background()
			store := novelfetchworkshop.NewMemoryStore()
			doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{
				"status": status, "siteSubmitStatus": "pending_confirmation",
				"processSnapshot": map[string]any{"auto_submit": false, "selected_versions": []string{"ai1", "ai2"}},
			}, Versions: map[string]any{"ai1": "第一版已生成", "ai2": "第二版已生成"}}
			item := saveSubmissionEvidenceSchedule(t, store, doc, false)
			item["status"], item["pending_book_ids"] = status, nil
			if err := saveNovelFetchScheduleItem(ctx, store, "tester", item); err != nil {
				t.Fatal(err)
			}
			queue := &uniqueRecordingNovelFetchQueue{}
			if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", time.Now()); err != nil {
				t.Fatal(err)
			}
			if row := storedSubmissionEvidenceSchedule(t, store); row["status"] != "completed" || len(queue.jobs) != 0 {
				t.Fatalf("old upload state kept a generate-only schedule open: row=%+v jobs=%+v", row, queue.jobs)
			}
		})
	}
}

func TestDeferredGenerateOnlySchedulePreservesCurrentWorkWithOldProse(t *testing.T) {
	for _, scenario := range []struct {
		name, status, aiStatus, want string
		manual, held                 bool
	}{
		{name: "failed with validation hold", status: "failed", want: "failed", held: true},
		{name: "manual hold over stale upload", status: "upload_pending", want: "failed", manual: true},
		{name: "failed ai stage over stale upload", status: "upload_pending", aiStatus: "failed", want: "failed"},
		{name: "processing", status: "processing", want: "running"},
		{name: "waiting retry", status: "waiting_retry", want: "waiting_retry"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			ctx := context.Background()
			store := novelfetchworkshop.NewMemoryStore()
			doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{
				"status": scenario.status, "siteSubmitStatus": "pending_confirmation",
				"manualOnly": scenario.manual, "rewriteValidationHold": scenario.held, "aiStatus": scenario.aiStatus,
			}, Versions: map[string]any{"ai1": "第一版旧正文", "ai2": "第二版旧正文"}}
			item := saveSubmissionEvidenceSchedule(t, store, doc, false)
			item["pending_book_ids"] = nil
			if err := saveNovelFetchScheduleItem(ctx, store, "tester", item); err != nil {
				t.Fatal(err)
			}
			queue := &uniqueRecordingNovelFetchQueue{}
			if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", time.Now()); err != nil {
				t.Fatal(err)
			}
			if row := storedSubmissionEvidenceSchedule(t, store); row["status"] != scenario.want || len(queue.jobs) != 0 {
				t.Fatalf("old prose overrode current work: row=%+v want=%s jobs=%+v", row, scenario.want, queue.jobs)
			}
		})
	}
}

func TestDeferredGenerateOnlyScheduleUsesActiveQueueStateOverStaleCompletion(t *testing.T) {
	for _, previous := range []string{"completed", "upload_pending"} {
		t.Run(previous, func(t *testing.T) {
			ctx := context.Background()
			store := novelfetchworkshop.NewMemoryStore()
			queue := &uniqueRecordingNovelFetchQueue{}
			doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{
				"status": previous, "queueState": "running", "aiStatus": "done",
			}, Versions: map[string]any{"ai1": "第一版旧正文", "ai2": "第二版旧正文"}}
			saveSubmissionEvidenceSchedule(t, store, doc, false)
			now := time.Now()
			if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now); err != nil {
				t.Fatal(err)
			}
			if row := storedSubmissionEvidenceSchedule(t, store); row["status"] != "running" || len(queue.jobs) != 0 {
				t.Fatalf("stale completion masked the active queue state: row=%+v jobs=%+v", row, queue.jobs)
			}
			doc.Meta["status"], doc.Meta["queueState"] = "upload_pending", "upload_pending"
			if err := store.PutDocument(ctx, "tester", doc); err != nil {
				t.Fatal(err)
			}
			if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now.Add(time.Minute)); err != nil {
				t.Fatal(err)
			}
			if row := storedSubmissionEvidenceSchedule(t, store); row["status"] != "completed" || len(queue.jobs) != 0 {
				t.Fatalf("stale upload state after generation did not settle: row=%+v jobs=%+v", row, queue.jobs)
			}
		})
	}
}

func TestDueGenerateOnlyScheduleDoesNotSkipAnActiveBookWithOldProse(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := &uniqueRecordingNovelFetchQueue{}
	doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{
		"status": "completed", "queueState": "running", "aiStatus": "done", "scheduleId": "evidence-schedule",
	}, Versions: map[string]any{"ai1": "第一版旧正文", "ai2": "第二版旧正文"}}
	item := saveSubmissionEvidenceSchedule(t, store, doc, false)
	now := time.Now()
	item["status"], item["pending_book_ids"], item["retry_tracking"] = "scheduled", nil, false
	item["tasks_prepared"], item["run_at"] = true, now.Add(-time.Minute).Format(time.RFC3339Nano)
	cfg, err := store.GetConfig(ctx, "tester")
	if err != nil {
		t.Fatal(err)
	}
	cfg["novel_fetch_schedules"] = []any{item}
	if err := store.PutConfig(ctx, "tester", cfg); err != nil {
		t.Fatal(err)
	}
	active := novelfetchworkshop.ProcessJob{ID: "active-generation", Owner: "tester", PlatformID: "2", BookIDs: []string{doc.BookID}}
	if queued, _, err := queue.EnqueueUniqueBooks(ctx, active, time.Hour); err != nil || !queued {
		t.Fatalf("could not set up active generation: %v", err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now); err != nil {
		t.Fatal(err)
	}
	if row := storedSubmissionEvidenceSchedule(t, store); row["status"] != "running" || len(queue.jobs) != 1 {
		t.Fatalf("initial dispatch hid or duplicated active generation: row=%+v jobs=%+v", row, queue.jobs)
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
