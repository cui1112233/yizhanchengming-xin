package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/mergeworker"
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

func TestGenerateOnlyScheduleFirstDispatchAndRetryAgree(t *testing.T) {
	for _, scenario := range []struct {
		name, status, queueState, aiStatus, hold, want string
	}{
		{"validation hold during rewrite", "upload_pending", "running", "done", "rewriteValidationHold", "failed"},
		{"manual original during rewrite", "running", "queued", "", "manualOriginalFetchRequired", "failed"},
		{"manual only during rewrite", "completed", "processing", "", "manualOnly", "failed"},
		{"active queue after old completion", "completed", "running", "done", "", "running"},
		{"active retry after old upload", "upload_pending", "waiting_retry", "done", "", "waiting_retry"},
		{"waiting for classification after old upload", "upload_pending", "waiting_classification", "done", "", "waiting_classification"},
		{"failed rewrite with old prose", "upload_pending", "", "failed", "", "failed"},
		{"cancelled queue after old completion", "completed", "cancelled", "done", "", "stopped"},
		{"failed queue with active status", "running", "failed", "done", "", "waiting_retry"},
		{"legacy pending queue", "completed", "pending", "done", "", "queued"},
		{"legacy retry queue", "upload_pending", "retrying", "done", "", "waiting_retry"},
		{"legacy ai service wait", "waiting_ai_service", "", "", "", "waiting_sensitive_ai"},
		{"login wait", "waiting_login", "", "", "", "waiting_login"},
		{"configuration wait", "waiting_configuration", "", "", "", "waiting_configuration"},
		{"legacy pending submit", "pending_submit", "", "", "", "completed"},
		{"legacy uploaded", "uploaded", "", "", "", "completed"},
		{"legacy restored", "restored", "", "", "", "completed"},
		{"legacy complete", "complete", "", "", "", "completed"},
		{"completed selected versions", "upload_pending", "", "done", "", "completed"},
	} {
		for _, entry := range []string{"first dispatch", "interrupted preparation", "dispatch recovery", "recovery with completed run", "retry", "pending retry"} {
			t.Run(scenario.name+"/"+entry, func(t *testing.T) {
				ctx := context.Background()
				now := time.Now()
				store := novelfetchworkshop.NewMemoryStore()
				queue := &uniqueRecordingNovelFetchQueue{}
				doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{
					"status": scenario.status, "queueState": scenario.queueState, "aiStatus": scenario.aiStatus, "scheduleId": "evidence-schedule",
				}, Versions: map[string]any{"ai1": "第一版旧正文", "ai2": "第二版旧正文"}}
				if scenario.hold != "" {
					doc.Meta[scenario.hold] = true
				}
				// Persist the dispatch checkpoint first. Recovery must see the later book change.
				initial := doc
				initial.Meta = map[string]any{"scheduleId": "evidence-schedule", "scheduledHold": true, "status": "scheduled_waiting"}
				item := saveSubmissionEvidenceSchedule(t, store, initial, false)
				if entry != "pending retry" {
					item["pending_book_ids"] = nil
				}
				if entry == "first dispatch" || entry == "interrupted preparation" {
					item["status"], item["retry_tracking"], item["tasks_prepared"] = "scheduled", false, entry == "first dispatch"
					item["input_text"] = doc.BookID + " 测试"
					item["run_at"] = now.Add(-time.Minute).Format(time.RFC3339Nano)
				}
				if entry == "dispatch recovery" || entry == "recovery with completed run" {
					item["status"], item["tasks_prepared"], item["retry_tracking"] = "dispatching", true, false
					item["updatedAt"], item["run_at"] = now.Add(-time.Hour).Format(time.RFC3339Nano), now.Add(-2*time.Hour).Format(time.RFC3339Nano)
					item["dispatch_id"], item["input_text"] = "checkpoint", doc.BookID
					if entry == "recovery with completed run" {
						if err := store.PutRun(ctx, "tester", novelfetchworkshop.RunRecord{RunID: "evidence-job", BookID: "evidence-batch", Stage: "novel_fetch_process", ModelID: "novel-fetch-workshop", Status: "completed"}); err != nil {
							t.Fatal(err)
						}
					}
				}
				if err := store.PutConfig(ctx, "tester", map[string]any{
					"novel_fetch_schedules": []any{item}, novelFetchAutomationConfigKey: map[string]any{"enabled": true},
				}); err != nil {
					t.Fatal(err)
				}
				if err := store.PutDocument(ctx, "tester", doc); err != nil {
					t.Fatal(err)
				}
				if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now); err != nil {
					t.Fatal(err)
				}
				row := storedSubmissionEvidenceSchedule(t, store)
				if row["status"] != scenario.want || len(queue.jobs) != 0 {
					t.Fatalf("got status=%v jobs=%d, want %s and no new jobs", row["status"], len(queue.jobs), scenario.want)
				}
				if scenario.want != "completed" && !seenString(scheduleStringSlice(row["book_ids"]), doc.BookID) {
					t.Fatal("schedule lost the book that still needs attention")
				}
			})
		}
	}
}

func TestGenerateOnlyScheduleDispatchesUnfinishedBookAndTracksHeldBook(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	due := now.Add(time.Hour)
	item := createPreparedSchedule(t, store, map[string]any{
		"input_text": "12345678901 暂停\n12345678902 待改文\n12345678903 已被占用", "run_at": due.Format(time.RFC3339Nano),
		"selected_versions": []string{"ai1"}, "auto_submit": false,
	})
	ids := scheduleStringSlice(item["book_ids"])
	for i, id := range ids {
		doc, err := store.GetDocument(ctx, "tester", id)
		if err != nil {
			t.Fatal(err)
		}
		doc.Versions = map[string]any{"ai1": "旧正文"}
		if i == 0 {
			doc.Meta["queueState"], doc.Meta["rewriteValidationHold"] = "running", true
		}
		if err := store.PutDocument(ctx, "tester", doc); err != nil {
			t.Fatal(err)
		}
	}
	occupied := novelfetchworkshop.ProcessJob{ID: "existing", BatchID: "existing-batch", Owner: "tester", PlatformID: "2", BookIDs: []string{"12345678903"}}
	if ok, _, err := queue.EnqueueUniqueBooks(ctx, occupied, time.Hour); err != nil || !ok {
		t.Fatalf("set up occupied book: %v", err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due); err != nil {
		t.Fatal(err)
	}
	row := storedSubmissionEvidenceSchedule(t, store)
	busy, err := queue.Claim(ctx, time.Millisecond)
	if err != nil || busy.ID != occupied.ID {
		t.Fatalf("existing job changed: %+v %v", busy, err)
	}
	free, err := queue.Claim(ctx, time.Millisecond)
	if err != nil || len(free.BookIDs) != 1 || free.BookIDs[0] != "12345678902" || len(scheduleStringSlice(row["book_ids"])) != 3 {
		t.Fatalf("expected only free book queued and all books tracked: job=%+v row=%+v err=%v", free, row, err)
	}
	doc, err := store.GetDocument(ctx, "tester", "12345678902")
	if err != nil {
		t.Fatal(err)
	}
	doc.Meta["status"], doc.Meta["queueState"] = "completed", "completed"
	if err := store.PutDocument(ctx, "tester", doc); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	row = storedSubmissionEvidenceSchedule(t, store)
	if pending := scheduleStringSlice(row["pending_book_ids"]); row["status"] != "waiting_retry" || len(pending) != 1 || pending[0] != "12345678903" {
		t.Fatalf("occupied book lost while waiting: %+v", row)
	}
	if err := queue.Ack(ctx, busy); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	resumed, err := queue.Claim(ctx, time.Millisecond)
	if err != nil || len(resumed.BookIDs) != 1 || resumed.BookIDs[0] != "12345678903" {
		t.Fatalf("wrong book resumed after lock released: %+v %v", resumed, err)
	}
	doc, err = store.GetDocument(ctx, "tester", "12345678903")
	if err != nil {
		t.Fatal(err)
	}
	doc.Meta["status"], doc.Meta["queueState"] = "completed", "completed"
	if err := store.PutDocument(ctx, "tester", doc); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	if row := storedSubmissionEvidenceSchedule(t, store); row["status"] != "failed" {
		t.Fatalf("held book lost after other books completed: %+v", row)
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

func TestGenerateOnlyScheduleQueuesOnlyMissingVersionAfterActiveBook(t *testing.T) {
	for _, entry := range []string{"pending retry", "legacy retry without pending ids", "first dispatch observing active book"} {
		t.Run(entry, func(t *testing.T) {
			ctx := context.Background()
			now := time.Now().UTC()
			store := novelfetchworkshop.NewMemoryStore()
			queue := novelfetchworkshop.NewMemoryQueue()
			doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"scheduleId": "evidence-schedule"}}
			item := saveSubmissionEvidenceSchedule(t, store, doc, false)
			if entry != "pending retry" {
				item["pending_book_ids"] = nil
			}
			if entry == "legacy retry without pending ids" {
				// Older schedules observing an entire batch never saved the parent job ID.
				delete(item, "job_id")
				item["run_at"] = now.Add(-time.Hour).Format(time.RFC3339Nano)
			}
			if entry == "first dispatch observing active book" {
				delete(item, "job_id")
				item["status"], item["retry_tracking"], item["tasks_prepared"] = "scheduled", false, true
				item["run_at"] = now.Add(-time.Hour).Format(time.RFC3339Nano)
			}
			if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{item}}); err != nil {
				t.Fatal(err)
			}
			occupied := novelfetchworkshop.ProcessJob{ID: "other-job", BatchID: "other-batch", Owner: "tester", PlatformID: "2", BookIDs: []string{doc.BookID}, SelectedVersions: []string{"ai1"}}
			if ok, _, err := queue.EnqueueUniqueBooks(ctx, occupied, time.Hour); err != nil || !ok {
				t.Fatalf("occupy book: %v", err)
			}
			active, err := queue.Claim(ctx, time.Millisecond)
			if err != nil {
				t.Fatal(err)
			}
			if err := updateWorkshopTaskRuntimeState(ctx, store, "tester", doc.BookID, "processing", ""); err != nil {
				t.Fatal(err)
			}
			tick := func() map[string]any {
				t.Helper()
				now = now.Add(time.Second)
				if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now); err != nil {
					t.Fatal(err)
				}
				return storedSubmissionEvidenceSchedule(t, store)
			}
			row := tick()
			if row["status"] != "running" || row["retry_tracking"] != true {
				t.Fatalf("active job lost tracking: %+v", row)
			}
			if entry == "pending retry" && !slices.Contains(scheduleStringSlice(row["pending_book_ids"]), doc.BookID) {
				t.Errorf("active book removed from pending list")
			}
			if _, err := queue.Claim(ctx, time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
				t.Fatalf("unexpected job during active work: %v", err)
			}
			saveBody := func(version, body string) {
				t.Helper()
				doc, err := store.GetDocument(ctx, "tester", doc.BookID)
				if err != nil {
					t.Fatal(err)
				}
				if doc.Versions == nil {
					doc.Versions = map[string]any{}
				}
				doc.Versions[version] = body
				if err := store.PutDocument(ctx, "tester", doc); err != nil {
					t.Fatal(err)
				}
				if err := updateWorkshopTaskRuntimeState(ctx, store, "tester", doc.BookID, "completed", ""); err != nil {
					t.Fatal(err)
				}
			}
			saveBody("ai1", "第一版")
			if err := queue.Ack(ctx, active); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 3; i++ {
				tick()
			}
			missing, err := queue.Claim(ctx, time.Millisecond)
			if err != nil || !slices.Equal(missing.BookIDs, []string{doc.BookID}) || !slices.Equal(missing.SelectedVersions, []string{"ai2"}) || missing.AutoSubmit {
				t.Fatalf("expected only ai2 after other job finished: job=%+v err=%v", missing, err)
			}
			row = storedSubmissionEvidenceSchedule(t, store)
			if firstText(row["job_id"]) == "" || !slices.Equal(scheduleStringSlice(row["selected_versions"]), []string{"ai1", "ai2"}) {
				t.Fatalf("lost schedule identity or original selection: %+v", row)
			}
			saveBody("ai2", "第二版")
			if err := queue.Ack(ctx, missing); err != nil {
				t.Fatal(err)
			}
			if row := tick(); row["status"] != "completed" || row["retry_tracking"] == true {
				t.Fatalf("parent did not finish with both versions: %+v", row)
			}
			if _, err := queue.Claim(ctx, time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
				t.Fatalf("unexpected duplicate after completion: %v", err)
			}
		})
	}
}

func TestGenerateOnlyScheduleSettlesAfterShortOriginal(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	if err := store.PutConfig(ctx, "tester", map[string]any{
		"workflow":                    map[string]any{"skip_short_original": true, "min_original_chars": 10},
		novelFetchAutomationConfigKey: map[string]any{"enabled": true},
	}); err != nil {
		t.Fatal(err)
	}
	doc := novelfetchworkshop.Document{BookID: "12345678901", OriginalRaw: "短文"}
	saveSubmissionEvidenceSchedule(t, store, doc, false)
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now); err != nil {
		t.Fatal(err)
	}
	first, err := queue.Claim(ctx, time.Second)
	if err != nil || !slices.Equal(first.BookIDs, []string{doc.BookID}) {
		t.Fatalf("first dispatch: job=%+v schedule=%+v err=%v", first, storedSubmissionEvidenceSchedule(t, store), err)
	}
	worker := workshopProcessWorker{store: store, queue: queue, registry: newWorkshopProcessRegistry()}
	if err := worker.process(ctx, first); err != nil {
		t.Fatal(err)
	}
	skipped, err := store.GetDocument(ctx, "tester", doc.BookID)
	if err != nil || skipped.Meta["originalStatus"] != "too_short" || skipped.Meta["aiStatus"] != "skipped_short_original" || len(skipped.Logs) == 0 {
		t.Fatalf("worker did not record the short skip: doc=%+v err=%v", skipped, err)
	}
	for cycle := 1; cycle <= 3; cycle++ {
		if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now.Add(time.Duration(cycle)*time.Second)); err != nil {
			t.Fatal(err)
		}
		row := storedSubmissionEvidenceSchedule(t, store)
		if pending := scheduleStringSlice(row["pending_book_ids"]); row["status"] != "completed" || len(pending) != 0 {
			t.Errorf("cycle %d after short skip: status=%v pending=%v", cycle, row["status"], pending)
		}
		if duplicate, err := queue.Claim(ctx, time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
			t.Fatalf("cycle %d dispatched short book again: job=%s books=%v err=%v", cycle, duplicate.ID, duplicate.BookIDs, err)
		}
	}
	after, err := store.GetDocument(ctx, "tester", doc.BookID)
	if err != nil || after.Meta["originalStatus"] != "too_short" || after.Meta["aiStatus"] != "skipped_short_original" || !reflect.DeepEqual(after.Logs, skipped.Logs) {
		t.Fatalf("short skip reason or processing history changed: doc=%+v err=%v", after, err)
	}
}

func TestGenerateOnlyScheduleShortSkipYieldsToCurrentState(t *testing.T) {
	for _, tc := range []struct {
		key      string
		value    any
		want     string
		tracking bool
	}{
		{"manualOriginalFetchRequired", true, "failed", false},
		{"manualOnly", true, "failed", false},
		{"rewriteValidationHold", true, "failed", false},
		{"queueState", "processing", "running", true},
		{"aiStatus", "failed", "failed", false},
		{"status", "failed", "waiting_retry", true},
	} {
		t.Run(tc.key, func(t *testing.T) {
			ctx := context.Background()
			store := novelfetchworkshop.NewMemoryStore()
			queue := &uniqueRecordingNovelFetchQueue{}
			doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{
				"status": "completed", "originalStatus": "too_short", "aiStatus": "skipped_short_original",
			}}
			doc.Meta[tc.key] = tc.value
			item := saveSubmissionEvidenceSchedule(t, store, doc, false)
			item["pending_book_ids"] = nil
			if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{item}, novelFetchAutomationConfigKey: map[string]any{"enabled": true}}); err != nil {
				t.Fatal(err)
			}
			if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", time.Now()); err != nil {
				t.Fatal(err)
			}
			row := storedSubmissionEvidenceSchedule(t, store)
			if row["status"] != tc.want || row["retry_tracking"] != tc.tracking || len(queue.jobs) != 0 {
				t.Fatalf("new %s=%v: status=%v tracking=%v jobs=%d", tc.key, tc.value, row["status"], row["retry_tracking"], len(queue.jobs))
			}
		})
	}
}

func TestGenerateOnlyScheduleKeepsShortAndCancelledBooksAtFirstDispatchAndRetry(t *testing.T) {
	for _, entry := range []string{"first", "retry"} {
		t.Run(entry, func(t *testing.T) {
			ctx := context.Background()
			store := novelfetchworkshop.NewMemoryStore()
			queue := &uniqueRecordingNovelFetchQueue{}
			now := time.Now().UTC()
			item := createPreparedSchedule(t, store, map[string]any{"run_at": now.Add(time.Hour).Format(time.RFC3339Nano), "input_text": "12345678901 短书\n12345678902 取消", "selected_versions": []string{"ai1"}, "auto_submit": false})
			ids := scheduleStringSlice(item["book_ids"])
			for i, id := range ids {
				doc, err := store.GetDocument(ctx, "tester", id)
				if err != nil {
					t.Fatal(err)
				}
				delete(doc.Meta, "scheduledHold")
				doc.Meta["status"], doc.Meta["queueState"] = "completed", "completed"
				if i == 0 {
					doc.Meta["originalStatus"] = "too_short"
				} else {
					doc.Meta["queueState"] = "cancelled"
				}
				if err := store.PutDocument(ctx, "tester", doc); err != nil {
					t.Fatal(err)
				}
			}
			if entry == "retry" {
				item["status"], item["retry_tracking"] = "waiting_retry", true
				if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{item}}); err != nil {
					t.Fatal(err)
				}
			}
			if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now.Add(2*time.Hour)); err != nil {
				t.Fatal(err)
			}
			row := storedSubmissionEvidenceSchedule(t, store)
			if row["status"] != "completed" || !slices.Contains(scheduleStringSlice(row["book_ids"]), ids[0]) || !slices.Contains(scheduleStringSlice(row["book_ids"]), ids[1]) || len(scheduleStringSlice(row["pending_book_ids"])) != 0 || len(queue.jobs) != 0 {
				t.Fatalf("status=%v books=%v pending=%v jobs=%d", row["status"], row["book_ids"], row["pending_book_ids"], len(queue.jobs))
			}
		})
	}
}

type manualImportTestQueue interface {
	novelfetchworkshop.ProcessQueue
	novelfetchworkshop.ProcessCancellationQueue
	novelfetchworkshop.ProcessBookDedupQueue
	novelfetchworkshop.ProcessBookConflictQueue
}

type manualImportStopQueue struct {
	manualImportTestQueue
	afterPreflight func()
	beforeEnqueue  func()
	afterEnqueue   func()
	onAccepted     func(novelfetchworkshop.ProcessJob)
	enqueueErr     error
}

func (q *manualImportStopQueue) ConflictingBookIDs(ctx context.Context, job novelfetchworkshop.ProcessJob) (map[string]string, error) {
	conflicts, err := q.manualImportTestQueue.ConflictingBookIDs(ctx, job)
	if q.afterPreflight != nil {
		q.afterPreflight()
	}
	return conflicts, err
}

func (q *manualImportStopQueue) EnqueueUniqueBooks(ctx context.Context, job novelfetchworkshop.ProcessJob, ttl time.Duration) (bool, string, error) {
	if q.beforeEnqueue != nil {
		q.beforeEnqueue()
	}
	if q.enqueueErr != nil {
		return false, "", q.enqueueErr
	}
	ok, id, err := q.manualImportTestQueue.EnqueueUniqueBooks(ctx, job, ttl)
	if ok && err == nil && q.onAccepted != nil {
		q.onAccepted(job)
	}
	if q.afterEnqueue != nil {
		q.afterEnqueue()
	}
	return ok, id, err
}

func (q *manualImportStopQueue) Claim(ctx context.Context, _ time.Duration) (novelfetchworkshop.ProcessJob, error) {
	<-ctx.Done()
	return novelfetchworkshop.ProcessJob{}, ctx.Err()
}

type manualImportStopStore struct {
	*novelfetchworkshop.MemoryStore
	beforeQueuedWrite func()
}

func (s *manualImportStopStore) CompareAndSwapDocument(ctx context.Context, owner string, doc novelfetchworkshop.Document) (string, error) {
	if s.beforeQueuedWrite != nil && doc.Meta["status"] == "queued" {
		s.beforeQueuedWrite()
	}
	return s.MemoryStore.CompareAndSwapDocument(ctx, owner, doc)
}

func TestManualProcessStartStopAfterPreflight(t *testing.T) {
	for _, timing := range []string{"after preflight", "between read and queued write", "after queued state", "during enqueue error", "after successful enqueue"} {
		t.Run(timing, func(t *testing.T) {
			ctx := context.Background()
			store := &manualImportStopStore{MemoryStore: novelfetchworkshop.NewMemoryStore()}
			queue := &manualImportStopQueue{manualImportTestQueue: novelfetchworkshop.NewMemoryQueue()}
			if url := os.Getenv("V88_NOVEL_FETCH_TEST_REDIS_URL"); url != "" {
				prefix := fmt.Sprintf("manual-stop-test:%d", time.Now().UnixNano())
				transport, err := mergeworker.NewRedisStore(url, prefix)
				if err != nil {
					t.Fatal(err)
				}
				queue.manualImportTestQueue = novelfetchworkshop.NewRedisQueueWithDo(prefix, novelfetchworkshop.RedisDoFunc(transport.Do))
				t.Cleanup(func() { _ = queue.ClearCancelledBooks(ctx, "tester", []string{"12345678901"}) })
			}
			now := time.Now()
			const id = "12345678901"
			if err := store.PutConfig(ctx, "tester", map[string]any{novelFetchAutomationConfigKey: map[string]any{"enabled": false}}); err != nil {
				t.Fatal(err)
			}
			if err := store.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: id, Meta: map[string]any{"status": "failed"}}); err != nil {
				t.Fatal(err)
			}
			h := NewRouter(RouterOptions{BridgeSecret: "manual-race", Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue})
			stop := func() {
				resp := doNovelFetchRequest(t, h, http.MethodPost, "/api/novel-fetch-workshop/process/stop", map[string]any{"ids": []string{id}}, now, "manual-race")
				if resp.Code != http.StatusOK {
					t.Fatalf("stop returned %d", resp.Code)
				}
			}
			want := http.StatusConflict
			switch timing {
			case "after preflight":
				queue.afterPreflight = stop
			case "between read and queued write":
				store.beforeQueuedWrite = stop
			case "after queued state":
				queue.beforeEnqueue = stop
			case "during enqueue error":
				queue.beforeEnqueue = stop
				queue.enqueueErr = errors.New("injected queue failure")
				want = http.StatusServiceUnavailable
			case "after successful enqueue":
				queue.afterEnqueue = stop
				want = http.StatusAccepted
			}
			resp := doNovelFetchRequest(t, h, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"input_text": id, "selected_versions": []string{"ai1"}}, now, "manual-race")
			cancelled, err := queue.IsBookCancelled(ctx, "tester", id)
			if err != nil {
				t.Fatal(err)
			}
			doc, err := store.GetDocument(ctx, "tester", id)
			if err != nil {
				t.Fatal(err)
			}
			job, err := queue.manualImportTestQueue.Claim(ctx, 5*time.Millisecond)
			if timing == "after successful enqueue" {
				if err != nil {
					t.Fatal(err)
				}
				w := workshopProcessWorker{store: store, queue: queue, registry: newWorkshopProcessRegistry()}
				if err := w.process(ctx, job); err != nil {
					t.Fatal(err)
				}
				if done, _ := w.registry.get("tester", job.ID); done["status"] != "stopped" {
					t.Fatalf("worker ran cancelled job: %v", done["status"])
				}
				_, err = queue.manualImportTestQueue.Claim(ctx, 5*time.Millisecond)
			}
			if resp.Code != want || !cancelled || doc.Meta["status"] != "stopped" || doc.Meta["queueState"] != "stopped" || !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
				t.Fatalf("start=%d cancelled=%v status=%v queueState=%v queued job=%v claim=%v", resp.Code, cancelled, doc.Meta["status"], doc.Meta["queueState"], job.ID, err)
			}
		})
	}
}

func TestGenerateOnlyScheduleRecognizesPersistedShortSkips(t *testing.T) {
	ctx := context.Background()
	for _, marker := range []struct{ key, value string }{{"original_status", "too_short"}, {"ai_status", "skipped_short_original"}} {
		t.Run(marker.key, func(t *testing.T) {
			store := novelfetchworkshop.NewMemoryStore()
			doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"status": "completed", marker.key: marker.value}}
			if err := store.PutDocument(ctx, "tester", doc); err != nil {
				t.Fatal(err)
			}
			item := createPreparedSchedule(t, store, map[string]any{
				"input_text": doc.BookID, "auto_submit": false, "selected_versions": []string{"ai1"}, "run_at": time.Now().Add(time.Hour).Format(time.RFC3339Nano),
			})
			skipped := arrayOrEmpty(item["skipped_books"])
			if item["status"] != "completed" || len(skipped) != 1 || !strings.Contains(firstText(mapOrEmpty(skipped[0])["reason"]), "原文过短") {
				t.Fatalf("prepared schedule: status=%v skipped=%v", item["status"], skipped)
			}
			saveSubmissionEvidenceSchedule(t, store, doc, false)
			queue := queueWithAbsentRedisCancellation(t)
			if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", time.Now()); err != nil {
				t.Fatalf("settle persisted pending book: %v", err)
			}
			row := storedSubmissionEvidenceSchedule(t, store)
			if row["status"] != "completed" || len(scheduleStringSlice(row["pending_book_ids"])) != 0 || len(queue.jobs) != 0 {
				t.Fatalf("short book remained pending: status=%v pending=%v jobs=%d", row["status"], row["pending_book_ids"], len(queue.jobs))
			}
		})
	}
}

func TestGenerateOnlyScheduleSettlesMixedTerminalBooks(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"status": "completed", "aiStatus": "skipped_short_original"}}
	item := saveSubmissionEvidenceSchedule(t, store, doc, false)
	for _, other := range []novelfetchworkshop.Document{
		{BookID: "12345678902", Meta: map[string]any{"status": "completed", "queueState": "cancelled"}},
		{BookID: "12345678903", Meta: map[string]any{"status": "completed"}, Versions: map[string]any{"ai1": "第一版", "ai2": "第二版"}},
	} {
		if err := store.PutDocument(ctx, "tester", other); err != nil {
			t.Fatal(err)
		}
	}
	item["book_ids"] = []string{doc.BookID, "12345678902", "12345678903"}
	item["pending_book_ids"] = nil
	if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{item}}); err != nil {
		t.Fatal(err)
	}
	queue := &uniqueRecordingNovelFetchQueue{}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", time.Now()); err != nil {
		t.Fatal(err)
	}
	if row := storedSubmissionEvidenceSchedule(t, store); row["status"] != "completed" || len(scheduleStringSlice(row["pending_book_ids"])) != 0 || len(queue.jobs) != 0 {
		t.Fatalf("mixed terminal books: status=%v pending=%v jobs=%d", row["status"], row["pending_book_ids"], len(queue.jobs))
	}
}

func TestManualProcessStartPreservesCancelledBook(t *testing.T) {
	for _, tc := range []struct {
		name, queueState string
		marker           bool
	}{
		{"stale completion and cancelled queue", "cancelled", true},
		{"cancelled queue only", "cancelled", false},
		{"cancellation marker only", "completed", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			now := time.Now()
			store := novelfetchworkshop.NewMemoryStore()
			queue := &recordingCancellationQueue{ProcessCancellationQueue: novelfetchworkshop.NewMemoryQueue()}
			if err := store.PutConfig(ctx, "tester", map[string]any{novelFetchAutomationConfigKey: map[string]any{"enabled": false}}); err != nil {
				t.Fatal(err)
			}
			doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"status": "completed", "queueState": tc.queueState}}
			if err := store.PutDocument(ctx, "tester", doc); err != nil {
				t.Fatal(err)
			}
			if tc.marker {
				if err := queue.CancelBooks(ctx, "tester", []string{doc.BookID}); err != nil {
					t.Fatal(err)
				}
			}
			handler := NewRouter(RouterOptions{BridgeSecret: "manual-import", Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue})
			response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{
				"input_text": doc.BookID, "selected_versions": []string{"ai1", "ai2"},
			}, now, "manual-import")
			var body struct {
				Result struct {
					Skipped []string `json:"skipped_existing_book_ids"`
				} `json:"result"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if response.Code != http.StatusOK || !slices.Equal(body.Result.Skipped, []string{doc.BookID}) || len(queue.jobs) != 0 {
				t.Errorf("import: HTTP %d skipped=%v jobs=%d", response.Code, body.Result.Skipped, len(queue.jobs))
			}
			if cancelled, err := queue.IsBookCancelled(ctx, "tester", doc.BookID); err != nil || cancelled != tc.marker {
				t.Errorf("cancellation marker: got %v want %v err=%v", cancelled, tc.marker, err)
			}
			if saved, err := store.GetDocument(ctx, "tester", doc.BookID); err != nil || saved.Meta["status"] != "completed" || saved.Meta["queueState"] != tc.queueState {
				t.Errorf("import changed cancelled document: status=%v queueState=%v err=%v", saved.Meta["status"], saved.Meta["queueState"], err)
			}
		})
	}
}

type recordingCancellationQueue struct {
	uniqueRecordingNovelFetchQueue
	novelfetchworkshop.ProcessCancellationQueue
}

// Production Redis represents an absent cancellation key with a RESP nil.
func queueWithAbsentRedisCancellation(t *testing.T) *recordingCancellationQueue {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		const getAbsent = "*2\r\n$3\r\nGET\r\n$6\r\nabsent\r\n"
		if _, err := io.ReadFull(conn, make([]byte, len(getAbsent))); err == nil {
			_, _ = io.WriteString(conn, "$-1\r\n")
		}
	}()
	transport, err := mergeworker.NewRedisStore("redis://"+listener.Addr().String(), "import-test")
	if err != nil {
		t.Fatal(err)
	}
	_, absent := transport.Do(context.Background(), "GET", "absent")
	if !mergeworker.IsRedisNil(absent) {
		t.Fatalf("fixture did not return a missing Redis key: %v", absent)
	}
	return &recordingCancellationQueue{ProcessCancellationQueue: novelfetchworkshop.NewRedisQueueWithDo("import-test", func(context.Context, ...string) (any, error) { return nil, absent })}
}

func TestManualProcessStartImportsFailedBookWithMissingBodyAndAutomationOff(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	store := novelfetchworkshop.NewMemoryStore()
	queue := queueWithAbsentRedisCancellation(t)
	if err := store.PutConfig(ctx, "tester", map[string]any{novelFetchAutomationConfigKey: map[string]any{"enabled": false}}); err != nil {
		t.Fatal(err)
	}
	doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"status": "failed"}, Versions: map[string]any{"ai1": "第一版"}}
	if err := store.PutDocument(ctx, "tester", doc); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: "manual-import", Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{
		"input_text": doc.BookID + " 测试", "selected_versions": []string{"ai1", "ai2"},
	}, now, "manual-import")
	if response.Code != http.StatusAccepted || len(queue.jobs) != 1 || !slices.Equal(queue.jobs[0].BookIDs, []string{doc.BookID}) {
		t.Fatalf("manual import got HTTP %d and %d jobs, want 202 and the failed book queued", response.Code, len(queue.jobs))
	}
}
func TestGenerateOnlyScheduleQueuesMissingBodyFromCompletedBook(t *testing.T) {
	for _, entry := range []string{"first dispatch", "dispatch recovery", "recovery with completed run"} {
		t.Run(entry, func(t *testing.T) {
			ctx := context.Background()
			now := time.Now().UTC()
			store := novelfetchworkshop.NewMemoryStore()
			queue := &uniqueRecordingNovelFetchQueue{}
			doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"scheduleId": "evidence-schedule", "status": "completed"}, Versions: map[string]any{"ai1": "第一版"}}
			item := saveSubmissionEvidenceSchedule(t, store, doc, false)
			item["maxTxt"], item["textModelId"], item["platformId"] = 2500, "saved-model", "7"
			item["status"], item["tasks_prepared"], item["retry_tracking"], item["pending_book_ids"] = "scheduled", true, false, nil
			item["run_at"] = now.Add(-2 * time.Hour).Format(time.RFC3339Nano)
			if entry != "first dispatch" {
				item["status"], item["dispatch_id"], item["updatedAt"] = "dispatching", "checkpoint", now.Add(-time.Hour).Format(time.RFC3339Nano)
				if entry == "recovery with completed run" {
					if err := store.PutRun(ctx, "tester", novelfetchworkshop.RunRecord{RunID: "evidence-job", BookID: "evidence-batch", Stage: "novel_fetch_process", ModelID: "novel-fetch-workshop", Status: "completed"}); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{item}}); err != nil {
				t.Fatal(err)
			}
			if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now); err != nil {
				t.Fatal(err)
			}
			if row := storedSubmissionEvidenceSchedule(t, store); row["status"] != "queued" || len(queue.jobs) != 1 || !slices.Equal(queue.jobs[0].SelectedVersions, []string{"ai2"}) {
				t.Fatalf("completion with missing ai2: status=%v jobs=%+v", row["status"], queue.jobs)
			}
			if job := queue.jobs[0]; job.MaxTxt != 2500 || job.TextModelID != "saved-model" || job.PlatformID != "7" || !job.SensitiveAIEnabled {
				t.Fatalf("narrowing to ai2 lost the saved processing options: %+v", job)
			}
		})
	}
}
func TestGenerateOnlyScheduleRecoveryHonorsHoldInLegacyBatch(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	store := novelfetchworkshop.NewMemoryStore()
	queue := &uniqueRecordingNovelFetchQueue{}
	doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"currentBatchId": "evidence-batch", "rewriteValidationHold": true}}
	item := saveSubmissionEvidenceSchedule(t, store, doc, false)
	item["status"], item["tasks_prepared"], item["retry_tracking"], item["pending_book_ids"] = "dispatching", true, false, nil
	item["run_at"], item["dispatch_id"], item["updatedAt"] = now.Add(-2*time.Hour).Format(time.RFC3339Nano), "checkpoint", now.Add(-time.Hour).Format(time.RFC3339Nano)
	if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{item}}); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now); err != nil {
		t.Fatal(err)
	}
	if row := storedSubmissionEvidenceSchedule(t, store); row["status"] != "failed" || len(queue.jobs) != 0 || !slices.Contains(scheduleStringSlice(row["book_ids"]), doc.BookID) {
		t.Fatalf("legacy batch lost its held book: row=%+v jobs=%+v", row, queue.jobs)
	}
}
func TestGenerateOnlyScheduleSettlesActiveBookWhileQueuePaused(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := &uniqueRecordingNovelFetchQueue{}
	doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"status": "processing"}}
	item := saveSubmissionEvidenceSchedule(t, store, doc, false)
	item["status"] = "running"
	if err := store.PutConfig(ctx, "tester", map[string]any{
		"novel_fetch_schedules": []any{item}, novelFetchAutomationConfigKey: map[string]any{"enabled": false},
	}); err != nil {
		t.Fatal(err)
	}
	doc.Meta["status"] = "completed"
	doc.Versions = map[string]any{"ai1": "第一版"}
	if err := store.PutDocument(ctx, "tester", doc); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", time.Now()); err != nil {
		t.Fatal(err)
	}
	if row := storedSubmissionEvidenceSchedule(t, store); row["status"] != "waiting_retry" || len(queue.jobs) != 0 || !slices.Contains(scheduleStringSlice(row["pending_book_ids"]), doc.BookID) {
		t.Fatalf("paused schedule must retain missing work without dispatch: row=%+v jobs=%+v", row, queue.jobs)
	}
	doc.Versions["ai2"] = "第二版"
	if err := store.PutDocument(ctx, "tester", doc); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", time.Now()); err != nil {
		t.Fatal(err)
	}
	if row := storedSubmissionEvidenceSchedule(t, store); row["status"] != "completed" || len(queue.jobs) != 0 || len(scheduleStringSlice(row["pending_book_ids"])) != 0 {
		t.Fatalf("paused queue masked completed work: row=%+v jobs=%+v", row, queue.jobs)
	}
}

const (
	cancelRaceA = "12345678901"
	cancelRaceB = "12345678902"
)

func cancelRaceQueue(t *testing.T) manualImportTestQueue {
	t.Helper()
	url := os.Getenv("V88_NOVEL_FETCH_TEST_REDIS_URL")
	if url == "" {
		return novelfetchworkshop.NewMemoryQueue()
	}
	prefix := fmt.Sprintf("dispatch-cancel-test:%d", time.Now().UnixNano())
	transport, err := mergeworker.NewRedisStore(url, prefix)
	if err != nil {
		t.Fatal(err)
	}
	q := novelfetchworkshop.NewRedisQueueWithDo(prefix, novelfetchworkshop.RedisDoFunc(transport.Do))
	t.Cleanup(func() { _ = q.ClearCancelledBooks(context.Background(), "tester", []string{cancelRaceA, cancelRaceB}) })
	return q
}

// Keep scheduler polling out of tests that drive dispatch themselves.
type cancelRaceStore struct {
	novelfetchworkshop.LifecycleStore
}

type cancelRaceImportStore struct {
	cancelRaceStore
	beforeCAS func(novelfetchworkshop.Document)
}

func (s *cancelRaceImportStore) CompareAndSwapDocument(ctx context.Context, owner string, doc novelfetchworkshop.Document) (string, error) {
	if s.beforeCAS != nil {
		s.beforeCAS(doc)
	}
	return s.cancelRaceStore.CompareAndSwapDocument(ctx, owner, doc)
}

type gatedStopWorkerQueue struct {
	*manualImportStopQueue
	claim chan struct{}
	acked chan struct{}
}

func (q *gatedStopWorkerQueue) Claim(ctx context.Context, wait time.Duration) (novelfetchworkshop.ProcessJob, error) {
	select {
	case <-ctx.Done():
		return novelfetchworkshop.ProcessJob{}, ctx.Err()
	case <-q.claim:
		return q.manualImportTestQueue.Claim(ctx, wait)
	}
}

func (q *gatedStopWorkerQueue) Ack(ctx context.Context, job novelfetchworkshop.ProcessJob) error {
	err := q.manualImportTestQueue.Ack(ctx, job)
	if err == nil {
		close(q.acked)
	}
	return err
}

func TestManualProcessStartPublishesCurrentJobAfterStop(t *testing.T) {
	for _, timing := range []string{"worker before response", "worker after response", "stop after response"} {
		t.Run(timing, func(t *testing.T) {
			ctx := context.Background()
			store := cancelRaceStore{novelfetchworkshop.NewMemoryStore()}
			queue := &gatedStopWorkerQueue{manualImportStopQueue: &manualImportStopQueue{manualImportTestQueue: cancelRaceQueue(t)}, claim: make(chan struct{}, 1), acked: make(chan struct{})}
			now := time.Now()
			const secret = "stop-registry-test"
			h := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue})
			stop := func() {
				r := doNovelFetchRequest(t, h, http.MethodPost, "/api/novel-fetch-workshop/process/stop", map[string]any{"ids": []string{cancelRaceA}}, now, secret)
				if r.Code != http.StatusOK {
					t.Fatalf("stop = %d", r.Code)
				}
			}
			finish := func() {
				queue.claim <- struct{}{}
				select {
				case <-queue.acked:
				case <-time.After(5 * time.Second):
					t.Fatal("worker did not acknowledge the stopped job")
				}
			}
			if timing != "stop after response" {
				queue.afterEnqueue = func() {
					stop()
					if timing == "worker before response" {
						finish()
					}
				}
			}
			r := doNovelFetchRequest(t, h, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"input_text": cancelRaceA, "selected_versions": []string{"ai1"}}, now, secret)
			var started map[string]any
			if err := json.Unmarshal(r.Body.Bytes(), &started); err != nil || r.Code != http.StatusAccepted {
				t.Fatalf("start = %d, %v", r.Code, err)
			}
			if timing == "stop after response" {
				stop()
			} else if started["status"] != "stopped" || started["queue_state"] != "stopped" {
				t.Errorf("start status = %v/%v after stop", started["status"], started["queue_state"])
			}
			if timing == "worker before response" && intValue(started["attempts"]) != 1 {
				t.Errorf("start lost the acknowledged worker attempt: %v", started["attempts"])
			}
			checkAPI := func(acknowledged bool) {
				for _, path := range []string{firstText(started["id"]), "latest"} {
					r := doNovelFetchRequest(t, h, http.MethodGet, "/api/novel-fetch-workshop/process/jobs/"+path, nil, now, secret)
					var view map[string]any
					if err := json.Unmarshal(r.Body.Bytes(), &view); err != nil || r.Code != http.StatusOK {
						t.Fatalf("job %s: %d, %v", path, r.Code, err)
					}
					if path == "latest" {
						view = mapOrEmpty(view["latest"])
					}
					if view["status"] != "stopped" || view["queue_state"] != "stopped" {
						t.Errorf("job %s = %v/%v after stop", path, view["status"], view["queue_state"])
					}
					if acknowledged && intValue(view["attempts"]) != 1 {
						t.Errorf("job %s lost the worker's registry entry: attempts=%v", path, view["attempts"])
					}
				}
			}
			checkAPI(timing == "worker before response")
			if timing != "worker before response" {
				finish()
				checkAPI(true)
			}
			doc, err := store.GetDocument(ctx, "tester", cancelRaceA)
			if err != nil || doc.Meta["status"] != "stopped" || doc.Meta["queueState"] != "stopped" {
				t.Fatalf("document status = %v/%v err=%v", doc.Meta["status"], doc.Meta["queueState"], err)
			}
			if _, err := queue.manualImportTestQueue.Claim(ctx, time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
				t.Fatalf("queue after Ack: %v", err)
			}
		})
	}
}

func TestScheduleDispatchKeepsSurvivorOfLateCancellation(t *testing.T) {
	for _, entry := range []string{"initial batch", "pending child", "pending missing version", "default submit options"} {
		t.Run(entry, func(t *testing.T) {
			ctx := context.Background()
			store := cancelRaceStore{novelfetchworkshop.NewMemoryStore()}
			queue := &manualImportStopQueue{manualImportTestQueue: cancelRaceQueue(t)}
			due := time.Now().UTC().Add(time.Hour)
			withSubmit := entry == "default submit options"
			item := createPreparedSchedule(t, store, map[string]any{"run_at": due.Format(time.RFC3339Nano), "input_text": cancelRaceA + " A\n" + cancelRaceB + " B", "selected_versions": []string{"ai1"}, "auto_submit": withSubmit})
			if entry == "pending child" || entry == "pending missing version" {
				item["pending_book_ids"], item["partial_dispatch"], item["job_id"] = []string{cancelRaceA, cancelRaceB}, true, "pending-parent"
				if err := saveNovelFetchScheduleItem(ctx, store, "tester", item); err != nil {
					t.Fatal(err)
				}
				if entry == "pending missing version" {
					doc, err := store.GetDocument(ctx, "tester", cancelRaceA)
					if err != nil {
						t.Fatal(err)
					}
					doc.Meta["status"], doc.Meta["queueState"] = "completed", "completed"
					if err := store.PutDocument(ctx, "tester", doc); err != nil {
						t.Fatal(err)
					}
				}
			}
			queue.beforeEnqueue = func() {
				queue.beforeEnqueue = nil
				if err := queue.CancelBooks(ctx, "tester", []string{cancelRaceA}); err != nil {
					t.Fatal(err)
				}
			}
			for i := 0; i < 2; i++ {
				if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(time.Duration(i+1)*time.Minute)); err != nil {
					t.Fatalf("dispatch %d: %v", i, err)
				}
			}
			row := storedSubmissionEvidenceSchedule(t, store)
			if row["retry_tracking"] != true || row["status"] != "queued" || len(scheduleStringSlice(row["pending_book_ids"])) != 0 || !slices.Equal(scheduleStringSlice(row["book_ids"]), []string{cancelRaceA, cancelRaceB}) {
				t.Errorf("parent=%v tracking=%v pending=%v books=%v", row["status"], row["retry_tracking"], row["pending_book_ids"], row["book_ids"])
			}
			var cancelled []string
			for _, raw := range arrayOrEmpty(row["skipped_books"]) {
				skipped := mapOrEmpty(raw)
				if skipped["reason"] == "cancelled" {
					cancelled = append(cancelled, firstText(skipped["book_id"]))
				}
			}
			if !slices.Equal(cancelled, []string{cancelRaceA}) {
				t.Errorf("cancelled book audit=%v", cancelled)
			}
			job, err := queue.manualImportTestQueue.Claim(ctx, time.Millisecond)
			if err != nil || !slices.Equal(job.BookIDs, []string{cancelRaceB}) || !slices.Equal(job.SelectedVersions, []string{"ai1"}) {
				t.Fatalf("admitted books=%v versions=%v err=%v", job.BookIDs, job.SelectedVersions, err)
			}
			stopped, err := store.GetDocument(ctx, "tester", cancelRaceA)
			if err != nil || !novelFetchStopped(firstText(stopped.Meta["queueState"])) {
				t.Errorf("cancelled book = %v/%v err=%v", stopped.Meta["status"], stopped.Meta["queueState"], err)
			}
			runs, err := store.ListRuns(ctx, "tester", "")
			if err != nil {
				t.Fatal(err)
			}
			var queued []string
			for _, run := range runs {
				if run.Status == "queued" {
					queued = append(queued, run.RunID)
				}
			}
			if !slices.Equal(queued, []string{job.ID}) {
				t.Errorf("queued audit jobs=%v, actual job=%v", queued, job.ID)
			}
			if err = queue.Ack(ctx, job); err != nil {
				t.Fatal(err)
			}
			done, err := store.GetDocument(ctx, "tester", cancelRaceB)
			if err != nil || done.Meta["currentProcessJobId"] != job.ID {
				t.Fatalf("survivor job=%v err=%v", done.Meta["currentProcessJobId"], err)
			}
			done.Meta["status"], done.Meta["queueState"] = "completed", "completed"
			done.Versions = map[string]any{"ai1": "B generated"}
			if withSubmit {
				done.Meta["siteSubmitStatus"], done.Meta["siteSubmitDoneVersions"] = "submitted", []string{"ai1"}
			}
			if err := store.PutDocument(ctx, "tester", done); err != nil {
				t.Fatal(err)
			}
			if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(3*time.Minute)); err != nil {
				t.Fatal(err)
			}
			row = storedSubmissionEvidenceSchedule(t, store)
			wantStatus := "completed"
			if withSubmit {
				wantStatus = "stopped"
			} // The current submission aggregation keeps a stopped book visible.
			if row["status"] != wantStatus || row["retry_tracking"] == true {
				t.Errorf("final parent=%v tracking=%v", row["status"], row["retry_tracking"])
			}
			if _, err := queue.manualImportTestQueue.Claim(ctx, time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
				t.Errorf("extra admission: %v", err)
			}
		})
	}
}

func TestScheduleKeepsSurvivorOfLateCancellationWhenWorkerFinishesBeforeDispatchReturns(t *testing.T) {
	ctx := context.Background()
	store := cancelRaceStore{novelfetchworkshop.NewMemoryStore()}
	queue := &manualImportStopQueue{manualImportTestQueue: cancelRaceQueue(t)}
	due := time.Now().UTC().Add(time.Hour)
	createPreparedSchedule(t, store, map[string]any{"run_at": due.Format(time.RFC3339Nano), "input_text": cancelRaceA + " A\n" + cancelRaceB + " B", "selected_versions": []string{"ai1"}, "auto_submit": false})
	cfg, err := store.GetConfig(ctx, "tester")
	if err != nil {
		t.Fatal(err)
	}
	cfg["workflow"] = map[string]any{"skip_short_original": true, "min_original_chars": 10}
	if err := store.PutConfig(ctx, "tester", cfg); err != nil {
		t.Fatal(err)
	}
	doc, err := store.GetDocument(ctx, "tester", cancelRaceB)
	if err != nil {
		t.Fatal(err)
	}
	doc.OriginalRaw = "短文"
	if err := store.PutDocument(ctx, "tester", doc); err != nil {
		t.Fatal(err)
	}
	queue.beforeEnqueue = func() {
		queue.beforeEnqueue = nil
		if err := queue.CancelBooks(ctx, "tester", []string{cancelRaceA}); err != nil {
			t.Fatal(err)
		}
	}
	var admitted int
	queue.onAccepted = func(enqueued novelfetchworkshop.ProcessJob) {
		admitted++
		job, err := queue.manualImportTestQueue.Claim(ctx, time.Millisecond)
		if err != nil || job.ID != enqueued.ID || !slices.Equal(job.BookIDs, []string{cancelRaceB}) {
			t.Fatalf("claim=%v err=%v", job, err)
		}
		worker := workshopProcessWorker{store: store, queue: queue, registry: newWorkshopProcessRegistry()}
		if err := worker.process(ctx, job); err != nil {
			t.Fatal(err)
		}
	}
	for i := 1; i <= 3; i++ {
		if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(time.Duration(i)*time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	row := storedSubmissionEvidenceSchedule(t, store)
	doc, err = store.GetDocument(ctx, "tester", cancelRaceB)
	if err != nil || admitted != 1 || row["status"] != "completed" || doc.Meta["queueState"] == "queued" || doc.Meta["aiStatus"] != "skipped_short_original" {
		t.Fatalf("admitted=%d parent=%v book=%v/%v ai=%v err=%v", admitted, row["status"], doc.Meta["status"], doc.Meta["queueState"], doc.Meta["aiStatus"], err)
	}
}

func TestManualProcessStartKeepsSurvivorOfLateCancellation(t *testing.T) {
	for _, timing := range []string{"before import", "after import read", "after snapshot read", "after preflight", "at enqueue", "new retry after atomic rejection"} {
		t.Run(timing, func(t *testing.T) {
			ctx := context.Background()
			store := &cancelRaceImportStore{cancelRaceStore: cancelRaceStore{novelfetchworkshop.NewMemoryStore()}}
			if err := store.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: cancelRaceA, Meta: map[string]any{"status": "failed"}}); err != nil {
				t.Fatal(err)
			}
			queue := &manualImportStopQueue{manualImportTestQueue: cancelRaceQueue(t)}
			now := time.Now()
			const secret = "manual-partial"
			h := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue})
			stop := func() {
				queue.beforeEnqueue, queue.afterPreflight = nil, nil
				store.beforeCAS = nil
				r := doNovelFetchRequest(t, h, http.MethodPost, "/api/novel-fetch-workshop/process/stop", map[string]any{"ids": []string{cancelRaceA}}, now, secret)
				if r.Code != http.StatusOK {
					t.Fatalf("stop=%d", r.Code)
				}
			}
			switch timing {
			case "before import":
				stop()
			case "after import read", "after snapshot read":
				store.beforeCAS = func(doc novelfetchworkshop.Document) {
					if doc.BookID == cancelRaceA && (timing == "after import read" || doc.Meta["processSnapshot"] != nil) {
						stop()
					}
				}
			case "after preflight":
				queue.afterPreflight = stop
			case "new retry after atomic rejection":
				queue.beforeEnqueue = stop
				queue.afterEnqueue = func() {
					queue.afterEnqueue = nil
					if err := queue.ClearCancelledBooks(ctx, "tester", []string{cancelRaceA}); err != nil {
						t.Fatal(err)
					}
					if err := updateWorkshopTaskRuntimeState(ctx, store, "tester", cancelRaceA, "waiting_retry", ""); err != nil {
						t.Fatal(err)
					}
				}
			default:
				queue.beforeEnqueue = stop
			}
			r := doNovelFetchRequest(t, h, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"input_text": cancelRaceA + " A\n" + cancelRaceB + " B", "selected_versions": []string{"ai1"}}, now, secret)
			var view map[string]any
			if err := json.Unmarshal(r.Body.Bytes(), &view); err != nil {
				t.Fatal(err)
			}
			if r.Code != http.StatusAccepted || !slices.Equal(scheduleStringSlice(view["book_ids"]), []string{cancelRaceB}) || mapOrEmpty(view["skipped_book_reasons"])[cancelRaceA] != "cancelled" {
				t.Errorf("start=%d accepted=%v skipped=%v", r.Code, view["book_ids"], view["skipped_book_reasons"])
			}
			job, err := queue.manualImportTestQueue.Claim(ctx, time.Millisecond)
			if err != nil || !slices.Equal(job.BookIDs, []string{cancelRaceB}) || job.ID != view["id"] {
				t.Fatalf("admitted=%v job=%v err=%v", job.BookIDs, job.ID, err)
			}
			stopped, err := store.GetDocument(ctx, "tester", cancelRaceA)
			want := "stopped"
			if timing == "new retry after atomic rejection" {
				want = "waiting_retry"
			}
			if err != nil || stopped.Meta["status"] != want || stopped.Meta["queueState"] != want {
				t.Fatalf("cancelled book=%v/%v err=%v", stopped.Meta["status"], stopped.Meta["queueState"], err)
			}
			if err := queue.Ack(ctx, job); err != nil {
				t.Fatal(err)
			}
			if _, err := queue.manualImportTestQueue.Claim(ctx, time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
				t.Errorf("extra admission: %v", err)
			}
		})
	}
}
