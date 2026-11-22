package httpapi

import (
	"context"
	"errors"
	"net/http"
	"slices"
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

func TestManualProcessStartImportsFailedBookWithMissingBodyAndAutomationOff(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	store := novelfetchworkshop.NewMemoryStore()
	queue := &uniqueRecordingNovelFetchQueue{}
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
