package httpapi

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/videomanagementconnector"
)

// Regression replay: no real database, provider or upload service.
func TestAuditRound2ChangedAICountRetry(t *testing.T) {
	for _, status := range []string{"completed", "failed"} {
		t.Run(status, func(t *testing.T) {
			ctx := context.Background()
			store := novelfetchworkshop.NewMemoryStore()
			queue := &uniqueRecordingNovelFetchQueue{}
			doc := novelfetchworkshop.Document{BookID: "12345678901", Original: "已保存原文", Versions: map[string]any{"ai1": "已完成AI1"}, Meta: map[string]any{"status": status, "originalStatus": "done", "aiStatus": "done", "aiCount": 1, "siteSubmitStatus": "submitted", "processSnapshot": map[string]any{"platform_id": "15", "max_txt": 4000, "selected_versions": []any{"ai1"}, "auto_submit": false}}}
			store.PutDocument(ctx, "tester", doc)
			now := time.Now()
			handler := NewRouter(RouterOptions{BridgeSecret: "round2", Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue})
			response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/batch-ai-count", map[string]any{"ids": []string{doc.BookID}, "count": 3}, now, "round2")
			if response.Code != 200 {
				t.Fatalf("count save: %d %s", response.Code, response.Body.String())
			}
			saved, _ := store.GetDocument(ctx, "tester", doc.BookID)
			t.Logf("saved aiCount=%v snapshot=%v", saved.Meta["aiCount"], mapOrEmpty(saved.Meta["processSnapshot"])["selected_versions"])
			result, err := enqueueWorkshopRetries(ctx, store, queue, "tester", []string{doc.BookID}, false)
			if err != nil {
				t.Fatal(err)
			}
			if result.Queued != 1 {
				t.Fatalf("missing AI2/AI3 not queued: %+v", result)
			}
			job := queue.jobs[0]
			if len(job.SelectedVersions) != 3 {
				t.Fatalf("retry still uses old target: %v", job.SelectedVersions)
			}
		})
	}
}

func TestAuditRound2ScheduleTerminalAndVersionEvidence(t *testing.T) {
	for _, scenario := range []string{"cancelled", "partial_submitted"} {
		t.Run(scenario, func(t *testing.T) {
			ctx := context.Background()
			store := novelfetchworkshop.NewMemoryStore()
			item := map[string]any{"id": "audit-schedule", "status": "waiting_retry", "retry_tracking": true, "book_ids": []any{"12345678901"}, "selected_versions": []any{"ai1", "ai2"}, "auto_submit": true}
			store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{item}})
			meta := map[string]any{"status": "cancelled"}
			if scenario == "partial_submitted" {
				meta = map[string]any{"status": "completed", "siteSubmitStatus": "submitted", "siteSubmitDoneVersions": []any{"ai1"}}
			}
			store.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: "12345678901", Meta: meta})
			cfg, _ := store.GetConfig(ctx, "tester")
			_, err := reconcileNovelFetchScheduleRetry(ctx, store, "tester", item, cfg, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			cfg, _ = store.GetConfig(ctx, "tester")
			row := mapOrEmpty(arrayOrEmpty(cfg["novel_fetch_schedules"])[0])
			t.Logf("actual schedule status=%v", row["status"])
			if scenario == "cancelled" && row["status"] != "cancelled" && row["status"] != "stopped" {
				t.Fatalf("cancelled book is not terminal: %v", row["status"])
			}
			if scenario == "partial_submitted" && row["status"] == "completed" {
				t.Fatal("AI2 missing but schedule marked completed")
			}
		})
	}
}

func TestAuditRound2OneBusyBookDoesNotAbandonOtherScheduledBook(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	due := time.Now().Add(time.Hour)
	item := createPreparedSchedule(t, store, map[string]any{"input_text": "12345678901 第一本\n12345678902 第二本", "run_at": due.Format(time.RFC3339Nano), "selected_versions": []string{"ai1"}, "auto_submit": false})
	occupied := novelfetchworkshop.ProcessJob{ID: "existing-busy-job", BatchID: "existing-batch", Owner: "tester", PlatformID: "2", BookIDs: []string{"12345678901"}}
	if ok, _, err := queue.EnqueueUniqueBooks(ctx, occupied, time.Hour); err != nil || !ok {
		t.Fatalf("queue setup: %v", err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	cfg, _ := store.GetConfig(ctx, "tester")
	row := mapOrEmpty(arrayOrEmpty(cfg["novel_fetch_schedules"])[0])
	doc, _ := store.GetDocument(ctx, "tester", "12345678902")
	t.Logf("schedule=%v book2=%v held=%v existingJob=%v scheduleID=%v", row["status"], doc.Meta["status"], doc.Meta["scheduledHold"], row["existing_job_id"], item["id"])
	if row["status"] == "skipped_existing" && doc.Meta["scheduledHold"] == true {
		t.Fatal("one Redis book lock abandoned the other book")
	}
	busy, err := queue.Claim(ctx, time.Millisecond)
	if err != nil || busy.ID != occupied.ID {
		t.Fatalf("existing job changed: %+v %v", busy, err)
	}
	free, err := queue.Claim(ctx, time.Millisecond)
	if err != nil || len(free.BookIDs) != 1 || free.BookIDs[0] != "12345678902" || free.AutoSubmit {
		t.Fatalf("free book did not independently retain its execution choice: %+v %v", free, err)
	}
	if len(scheduleStringSlice(row["pending_book_ids"])) != 1 || doc.Meta["status"] != "queued" {
		t.Fatalf("pending ownership lost: %+v %+v", row, doc.Meta)
	}
	// Completing the free child cannot complete a parent with a busy book.
	doc.Meta["status"], doc.Meta["queueState"] = "completed", "completed"
	doc.Versions = map[string]any{"ai1": "第二本结果"}
	store.PutDocument(ctx, "tester", doc)
	queue.Ack(ctx, free)
	updateNovelFetchScheduleRuntimeStatus(ctx, store, "tester", firstText(item["id"]), "completed", "")
	dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(2*time.Second))
	cfg, _ = store.GetConfig(ctx, "tester")
	row = mapOrEmpty(arrayOrEmpty(cfg["novel_fetch_schedules"])[0])
	if row["status"] == "completed" {
		t.Fatal("one child prematurely completed its parent")
	}
	// Releasing the original lock automatically dispatches only the missing book.
	queue.Ack(ctx, busy)
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	remaining, err := queue.Claim(ctx, time.Millisecond)
	if err != nil || len(remaining.BookIDs) != 1 || remaining.BookIDs[0] != "12345678901" {
		t.Fatalf("deferred book not resumed: %+v %v", remaining, err)
	}
	first, _ := store.GetDocument(ctx, "tester", "12345678901")
	first.Meta["status"], first.Meta["queueState"] = "completed", "completed"
	first.Versions = map[string]any{"ai1": "第一本结果"}
	store.PutDocument(ctx, "tester", first)
	queue.Ack(ctx, remaining)
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(4*time.Second)); err != nil {
		t.Fatal(err)
	}
	cfg, _ = store.GetConfig(ctx, "tester")
	row = mapOrEmpty(arrayOrEmpty(cfg["novel_fetch_schedules"])[0])
	if row["status"] != "completed" {
		t.Fatalf("all actual outputs did not settle parent: %+v", row)
	}
	if extra, err := queue.Claim(ctx, time.Millisecond); err == nil {
		t.Fatalf("completed output regenerated: %+v", extra)
	}
}

type auditRound2StopStore struct {
	novelfetchworkshop.LifecycleStore
	onPut func(context.Context, string, novelfetchworkshop.Document)
}

func (s *auditRound2StopStore) PutDocument(ctx context.Context, owner string, d novelfetchworkshop.Document) error {
	if err := s.LifecycleStore.PutDocument(ctx, owner, d); err != nil {
		return err
	}
	if s.onPut != nil {
		s.onPut(ctx, owner, d)
	}
	return nil
}
func TestAuditRound2StopDuringDispatchRemainsAuthoritative(t *testing.T) {
	ctx := context.Background()
	base := novelfetchworkshop.NewMemoryStore()
	store := &auditRound2StopStore{LifecycleStore: base}
	queue := novelfetchworkshop.NewMemoryQueue()
	due := time.Now().Add(time.Hour)
	createPreparedSchedule(t, store, map[string]any{"input_text": "12345678901 停止测试", "run_at": due.Format(time.RFC3339Nano), "auto_submit": false})
	fired := false
	store.onPut = func(ctx context.Context, owner string, d novelfetchworkshop.Document) {
		if fired || d.Meta["processSnapshot"] == nil {
			return
		}
		fired = true
		queue.CancelBooks(ctx, owner, []string{d.BookID})
		latest, _ := base.GetDocument(ctx, owner, d.BookID)
		latest.Meta["status"] = "stopped"
		latest.Meta["queueState"] = "stopped"
		base.PutDocument(ctx, owner, latest)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if !fired {
		t.Fatal("interleaving not exercised")
	}
	cancelled, _ := queue.IsBookCancelled(ctx, "tester", "12345678901")
	doc, _ := store.GetDocument(ctx, "tester", "12345678901")
	t.Logf("after manual stop: cancelled=%v documentStatus=%v", cancelled, doc.Meta["status"])
	if !cancelled || doc.Meta["status"] != "stopped" {
		t.Fatal("dispatch cleared a newer manual stop")
	}
}

func TestAuditRound2RetryKeepsExplicitGenerateOnlyChoice(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := &uniqueRecordingNovelFetchQueue{}
	store.PutConfig(ctx, "tester", map[string]any{novelFetchAutomationConfigKey: map[string]any{"enabled": true}})
	doc := novelfetchworkshop.Document{BookID: "12345678901", Original: "已有原文", Meta: map[string]any{"status": "failed", "originalStatus": "done", "aiStatus": "failed", "processSnapshot": map[string]any{"platform_id": "15", "max_txt": 4000, "selected_versions": []any{"ai1"}, "auto_submit": false}}}
	store.PutDocument(ctx, "tester", doc)
	result, err := enqueueWorkshopRetries(ctx, store, queue, "tester", []string{doc.BookID}, false)
	if err != nil || result.Queued != 1 {
		t.Fatalf("setup enqueue: %+v %v", result, err)
	}
	stored, _ := store.GetDocument(ctx, "tester", doc.BookID)
	t.Logf("job autoSubmit=%v saved snapshot auto_submit=%v", queue.jobs[0].AutoSubmit, mapOrEmpty(stored.Meta["processSnapshot"])["auto_submit"])
	if queue.jobs[0].AutoSubmit || mapOrEmpty(stored.Meta["processSnapshot"])["auto_submit"] != false {
		t.Fatal("retry changed explicit generate-only choice into submission")
	}
}

func TestWorkflowExplicitRetryRestartsStoppedBookWithoutChangingChoice(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	id := "12345678901"
	store.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: id, Meta: map[string]any{"status": "canceled", "manualOnly": true, "processSnapshot": map[string]any{"platform_id": "2", "auto_submit": false, "selected_versions": []string{"ai2"}}}})
	queue.CancelBooks(ctx, "tester", []string{id})
	result, err := enqueueWorkshopRetries(ctx, store, queue, "tester", []string{id}, false)
	if err != nil || result.Queued != 1 {
		t.Fatalf("explicit retry: %+v %v", result, err)
	}
	worker := &workshopProcessWorker{store: store, queue: queue}
	if worker.bookCancelled(ctx, "tester", id) {
		t.Fatal("explicit retry still blocked by old stopped state")
	}
	job, err := queue.Claim(ctx, time.Millisecond)
	if err != nil || job.AutoSubmit || len(job.SelectedVersions) != 1 || job.SelectedVersions[0] != "ai2" {
		t.Fatalf("retry changed original execution choices: %+v %v", job, err)
	}
}

func TestWorkflowAutomaticRecoveryKeepsGenerateOnlyChoice(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	store.PutConfig(ctx, "tester", map[string]any{novelFetchAutomationConfigKey: map[string]any{"enabled": true}})
	store.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"status": "waiting_retry", "processSnapshot": map[string]any{"platform_id": "2", "auto_submit": false, "selected_versions": []string{"ai1"}}}})
	job, err := dispatchNovelFetchAutomationOwner(ctx, store, queue, "tester", time.Now())
	if err != nil || job == nil || job.AutoSubmit {
		t.Fatalf("automatic recovery changed generate-only intent: %+v %v", job, err)
	}
}

func TestWorkflowUnknownSubmissionRemainsVerification(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	item := map[string]any{"id": "verification-schedule", "status": "waiting_retry", "retry_tracking": true, "book_ids": []string{"12345678901"}, "selected_versions": []string{"ai1", "ai2"}, "auto_submit": true}
	store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{item}})
	store.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"status": "completed", "siteSubmitStatus": "pending_verification", "siteSubmitDoneVersions": []string{"ai1"}}})
	if _, err := reconcileNovelFetchScheduleRetry(ctx, store, "tester", item, nil, time.Now()); err != nil {
		t.Fatal(err)
	}
	cfg, _ := store.GetConfig(ctx, "tester")
	row := mapOrEmpty(arrayOrEmpty(cfg["novel_fetch_schedules"])[0])
	if row["status"] != "pending_verification" {
		t.Fatalf("unknown outcome lost verification state: %+v", row)
	}
}

func TestScheduleUnknownSubmissionReconcilesWithoutUploadJob(t *testing.T) {
	for _, status := range []string{"pending_verification", "pending_confirmation"} {
		t.Run(status, func(t *testing.T) {
			store := novelfetchworkshop.NewMemoryStore()
			queue := &uniqueRecordingNovelFetchQueue{}
			doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{
				"status": "completed", "siteSubmitStatus": status, "siteSubmitDoneVersions": []string{"ai1"},
				"siteSubmitPendingConfirmation": map[string]any{"ai2": "pending-ai2"},
			}}
			item := saveSubmissionEvidenceSchedule(t, store, doc, true)
			ctx := context.Background()
			for i := 0; i < 2; i++ {
				if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", time.Now()); err != nil {
					t.Fatal(err)
				}
				if got := storedSubmissionEvidenceSchedule(t, store)["status"]; got != "pending_verification" {
					t.Errorf("schedule status = %v, want pending_verification", got)
				}
			}
			if len(queue.jobs) != 0 {
				t.Fatalf("unknown submission enqueued an upload job: %+v", queue.jobs)
			}
			// The existing readback path can later confirm the missing version.
			doc.Meta["siteSubmitStatus"] = "submitted"
			doc.Meta["siteSubmitDoneVersions"] = []string{"ai1", "ai2"}
			delete(doc.Meta, "siteSubmitPendingConfirmation")
			if err := store.PutDocument(ctx, "tester", doc); err != nil {
				t.Fatal(err)
			}
			if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", time.Now()); err != nil {
				t.Fatal(err)
			}
			if got := storedSubmissionEvidenceSchedule(t, store)["status"]; got != "completed" || len(queue.jobs) != 0 {
				t.Fatalf("schedule %s status = %v, jobs = %+v", item["id"], got, queue.jobs)
			}
		})
	}
}

func TestDueScheduleUnknownSubmissionReadsBackMatchingAttempt(t *testing.T) {
	ctx := context.Background()
	store, usersDir, doc := newSubmissionEvidenceTask(t)
	queue := &uniqueRecordingNovelFetchQueue{}
	doc.Meta["schedule_id"] = "evidence-schedule"
	doc.Meta["status"] = "upload_pending"
	doc.Meta["siteSubmitStatus"] = "pending_verification"
	doc.Meta["siteSubmitPendingConfirmation"] = map[string]any{"ai2": doc.BookID}
	fingerprint := fmt.Sprintf("%x", sha256.Sum256([]byte("第二版完整正文")))
	doc.Meta["siteSubmitAttempts"] = map[string]any{"ai2": map[string]any{"started_at": "2026-11-13T04:44:11Z", "content_sha256": fingerprint}}
	now := time.Now()
	item := saveSubmissionEvidenceSchedule(t, store, doc, true)
	item["status"], item["pending_book_ids"] = "scheduled", nil
	item["tasks_prepared"], item["run_at"] = true, now.Add(-time.Minute).Format(time.RFC3339Nano)
	cfg, err := store.GetConfig(ctx, "tester")
	if err != nil {
		t.Fatal(err)
	}
	cfg["novel_fetch_schedules"] = []any{item}
	if err := store.PutConfig(ctx, "tester", cfg); err != nil {
		t.Fatal(err)
	}
	connector := &scheduleSubmissionReadbackSpy{receipt: videomanagementconnector.Submission{
		ID: "first-ai1-receipt", Status: "submitted", ReadbackRaw: map[string]any{"bookid": doc.BookID, "addtime": "2026-11-10 23:54:40"},
	}}
	readback := newNovelFetchScheduleReadback(store, queue, usersDir, connector)
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now, readback); err != nil {
		t.Fatal(err)
	}
	if status := storedSubmissionEvidenceSchedule(t, store)["status"]; status != "pending_verification" || connector.readbackCalls != 1 || connector.submitCalls != 0 || len(queue.jobs) != 0 {
		t.Fatalf("old receipt settled a new version or started upload: status=%v readbacks=%d uploads=%d jobs=%v", status, connector.readbackCalls, connector.submitCalls, queue.jobs)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now.Add(time.Second), readback); err != nil || connector.readbackCalls != 1 {
		t.Fatalf("pending readback ignored retry interval: calls=%d error=%v", connector.readbackCalls, err)
	}
	connector.receipt = videomanagementconnector.Submission{ID: "new-ai2-receipt", Status: "submitted", ReadbackRaw: map[string]any{"bookid": doc.BookID, "addtime": "2026-11-13 12:45:35"}}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now.Add(time.Minute), readback); err != nil {
		t.Fatal(err)
	}
	if status := storedSubmissionEvidenceSchedule(t, store)["status"]; status != "pending_verification" || connector.readbackCalls != 2 {
		t.Fatalf("a book listing without a version receipt settled ai2: status=%v readbacks=%d", status, connector.readbackCalls)
	}
	mapOrEmpty(connector.receipt.ReadbackRaw)["content_sha256"] = fingerprint
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now.Add(2*time.Minute), readback); err != nil {
		t.Fatal(err)
	}
	if status := storedSubmissionEvidenceSchedule(t, store)["status"]; status != "completed" || connector.readbackCalls != 3 || connector.submitCalls != 0 || len(queue.jobs) != 0 {
		t.Fatalf("matching readback did not settle schedule: status=%v readbacks=%d uploads=%d jobs=%v", status, connector.readbackCalls, connector.submitCalls, queue.jobs)
	}
	saved, err := store.GetDocument(ctx, "tester", doc.BookID)
	if err != nil || !novelFetchVersionsSubmitted(saved, []string{"ai1", "ai2"}) || priorWorkshopSubmitReceipt(saved, "ai2") != "new-ai2-receipt" {
		t.Fatalf("missing confirmed receipt for ai2: %+v %v", saved, err)
	}
}

type scheduleSubmissionReadbackSpy struct {
	recordingSubmitConnector
	receipt       videomanagementconnector.Submission
	readbackCalls int
}

func TestScheduleReadbackThenQueuesRemainingUnsubmittedVersion(t *testing.T) {
	for _, scenario := range []struct {
		name                string
		legacy, alreadyRead bool
	}{{name: "keeps_pending_id"}, {name: "restores_legacy_id", legacy: true}, {name: "manual_readback", legacy: true, alreadyRead: true}} {
		t.Run(scenario.name, func(t *testing.T) {
			ctx := context.Background()
			store, usersDir, doc := newSubmissionEvidenceTask(t)
			queue := &uniqueRecordingNovelFetchQueue{}
			doc.Meta["status"] = "upload_pending"
			wantReadbacks := 0
			if !scenario.alreadyRead {
				wantReadbacks = 1
				doc.Logs = nil
				delete(doc.Meta, "siteSubmitDoneVersions")
				doc.Meta["siteSubmitStatus"] = "pending_confirmation"
				doc.Meta["siteSubmitPendingConfirmation"] = map[string]any{"ai1": "pending-ai1-receipt"}
			}
			item := saveSubmissionEvidenceSchedule(t, store, doc, true)
			if scenario.legacy {
				item["status"], item["pending_book_ids"] = "pending_verification", nil
				if err := saveNovelFetchScheduleItem(ctx, store, "tester", item); err != nil {
					t.Fatal(err)
				}
			}
			connector := &scheduleSubmissionReadbackSpy{receipt: videomanagementconnector.Submission{ID: "pending-ai1-receipt", Status: "readback"}}
			readback := newNovelFetchScheduleReadback(store, queue, usersDir, connector)
			now := time.Now()
			for i := 0; i < 4; i++ {
				if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now.Add(time.Duration(i)*time.Minute), readback); err != nil {
					t.Fatal(err)
				}
			}
			if connector.readbackCalls != wantReadbacks || connector.submitCalls != 0 || len(queue.jobs) != 1 || !slices.Equal(queue.jobs[0].SelectedVersions, []string{"ai2"}) {
				t.Fatalf("schedule stopped after ai1 readback: row=%+v jobs=%+v readbacks=%d uploads=%d", storedSubmissionEvidenceSchedule(t, store), queue.jobs, connector.readbackCalls, connector.submitCalls)
			}
		})
	}
}

func (c *scheduleSubmissionReadbackSpy) GetSubmission(context.Context, videomanagementconnector.Session, string) (videomanagementconnector.Submission, error) {
	c.readbackCalls++
	return c.receipt, nil
}

func TestSchedulePartialSubmissionQueuesOnlyMissingVersion(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := &uniqueRecordingNovelFetchQueue{}
	doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{
		"status": "failed", "siteSubmitStatus": "pending_upload",
		"siteSubmitDoneVersions": []string{"ai1"}, "siteSubmitFailedVersions": []string{"ai2"},
	}}
	saveSubmissionEvidenceSchedule(t, store, doc, true)
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", time.Now()); err != nil {
		t.Fatal(err)
	}
	if len(queue.jobs) != 1 || !queue.jobs[0].AutoSubmit || !slices.Equal(queue.jobs[0].SelectedVersions, []string{"ai2"}) {
		t.Fatalf("expected one upload job for ai2 only, got %+v", queue.jobs)
	}
	if got := scheduleStringSlice(storedSubmissionEvidenceSchedule(t, store)["selected_versions"]); !slices.Equal(got, []string{"ai1", "ai2"}) {
		t.Fatalf("schedule lost original version selection: %v", got)
	}
}

func TestWorkshopUploadPartialSubmissionReusesTheSavedFingerprint(t *testing.T) {
	ctx := context.Background()
	store, _, doc := newSubmissionEvidenceTask(t)
	doc.Meta["autoSubmitReadyVersions"] = []string{"ai1", "ai2"}
	if err := store.PutDocument(ctx, "tester", doc); err != nil {
		t.Fatal(err)
	}
	worker := &workshopProcessWorker{store: store, resolveRewriteProvider: func(context.Context, string) (novelfetchworkshop.RewriteProvider, error) {
		return nil, errors.New("模型不应在上传重试时启用")
	}}
	job := novelfetchworkshop.ProcessJob{Owner: "tester", BookIDs: []string{doc.BookID}, SelectedVersions: []string{"ai2"}, AutoSubmit: true, SensitiveAIEnabled: true}
	if err := worker.execute(ctx, job); err != nil {
		t.Fatalf("narrower upload retry discarded a matching readiness fingerprint: %v", err)
	}
}

func TestScheduleGenerateOnlyCompletionDoesNotRequireReceipt(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := &uniqueRecordingNovelFetchQueue{}
	doc := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{
		"status": "completed", "processSnapshot": map[string]any{"auto_submit": false, "selected_versions": []string{"ai1", "ai2"}},
	}, Versions: map[string]any{"ai1": "第一版正文", "ai2": "第二版正文"}}
	saveSubmissionEvidenceSchedule(t, store, doc, false)
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", time.Now()); err != nil {
		t.Fatal(err)
	}
	if got := storedSubmissionEvidenceSchedule(t, store)["status"]; got != "completed" || len(queue.jobs) != 0 {
		t.Fatalf("generate-only status = %v, jobs = %+v", got, queue.jobs)
	}
}

func TestScheduleCancelPreservesExistingAndCrossOwnerBooks(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	const existingID, createdID = "12345678901", "12345678902"
	for _, row := range []struct{ owner, id string }{{"tester", existingID}, {"other", createdID}} {
		if err := store.PutDocument(ctx, row.owner, novelfetchworkshop.Document{BookID: row.id, Original: "保留原文"}); err != nil {
			t.Fatal(err)
		}
	}
	item := createPreparedSchedule(t, store, map[string]any{
		"input_text": existingID + " 既有书\n" + createdID + " 新建书", "auto_submit": false,
		"selected_versions": []string{"ai1", "ai2"}, "run_at": time.Now().Add(time.Hour).Format(time.RFC3339Nano),
	})
	if !slices.Equal(scheduleStringSlice(item["created_ids"]), []string{createdID}) {
		t.Fatalf("expected just one schedule-created book, got %v", item["created_ids"])
	}
	response := scheduleControlRequest(store, queue, "tester", http.MethodPost, "/api/novel-fetch-workshop/schedules/"+firstText(item["id"])+"/cancel")
	if response.Code != http.StatusOK {
		t.Fatalf("cancel: %d %s", response.Code, response.Body.String())
	}
	if _, err := store.GetDocument(ctx, "tester", createdID); !errors.Is(err, novelfetchworkshop.ErrNotFound) {
		t.Fatalf("schedule-created book remains: %v", err)
	}
	for _, row := range []struct{ owner, id string }{{"tester", existingID}, {"other", createdID}} {
		if doc, err := store.GetDocument(ctx, row.owner, row.id); err != nil || doc.Original != "保留原文" {
			t.Fatalf("book %s/%s changed after cancellation: %+v %v", row.owner, row.id, doc, err)
		}
	}
}

func saveSubmissionEvidenceSchedule(t *testing.T, store novelfetchworkshop.Store, doc novelfetchworkshop.Document, autoSubmit bool) map[string]any {
	t.Helper()
	ctx := context.Background()
	if err := store.PutDocument(ctx, "tester", doc); err != nil {
		t.Fatal(err)
	}
	item := map[string]any{
		"id": "evidence-schedule", "job_id": "evidence-job", "batch_id": "evidence-batch",
		"status": "waiting_retry", "retry_tracking": true, "book_ids": []string{doc.BookID},
		"pending_book_ids": []string{doc.BookID}, "selected_versions": []string{"ai1", "ai2"}, "auto_submit": autoSubmit,
	}
	cfg, err := store.GetConfig(ctx, "tester")
	if err != nil {
		t.Fatal(err)
	}
	cfg["novel_fetch_schedules"] = []any{item}
	if err := store.PutConfig(ctx, "tester", cfg); err != nil {
		t.Fatal(err)
	}
	return item
}

func storedSubmissionEvidenceSchedule(t *testing.T, store novelfetchworkshop.Store) map[string]any {
	t.Helper()
	cfg, err := store.GetConfig(context.Background(), "tester")
	if err != nil {
		t.Fatal(err)
	}
	return mapOrEmpty(arrayOrEmpty(cfg["novel_fetch_schedules"])[0])
}
