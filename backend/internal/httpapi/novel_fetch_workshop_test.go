package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/novelfetchworkshop"
)

type recordingNovelFetchQueue struct {
	jobs []novelfetchworkshop.ProcessJob
}

func TestParseWorkshopDocumentsAcceptsLegacyFiveDigitSourceID(t *testing.T) {
	documents, parsed, emptyIDs, duplicates := parseWorkshopDocuments("12345\t短ID小说", 4000)
	if parsed != 1 || emptyIDs != 0 || duplicates != 0 || len(documents) != 1 {
		t.Fatalf("parse result = parsed=%d empty=%d duplicates=%d documents=%#v", parsed, emptyIDs, duplicates, documents)
	}
	if documents[0].BookID != "12345" {
		t.Fatalf("book id = %q, want 12345", documents[0].BookID)
	}
}

func TestNovelFetchAutomationRetriesPersistedSensitiveAIStage(t *testing.T) {
	document := novelfetchworkshop.Document{
		BookID:      "12345",
		OriginalRaw: "完整原文",
		Meta:        map[string]any{"originalStatus": "waiting_sensitive_ai", "status": "failed"},
	}
	if !novelFetchAutomationEligible(document, time.Now().UTC()) {
		t.Fatal("persisted raw text waiting for sensitive AI should be automation eligible")
	}
}

func TestNovelFetchClassificationWaitResumesOnlyAfterConfirmation(t *testing.T) {
	pending := novelfetchworkshop.Document{Gender: "未知", Meta: map[string]any{"status": "waiting_classification", "style": "待判定"}}
	if novelFetchAutomationEligible(pending, time.Now().UTC()) {
		t.Fatal("unresolved classification should wait for manual confirmation")
	}
	pending.Gender = "女频"
	pending.Meta["style"] = "现代悬疑"
	if !novelFetchAutomationEligible(pending, time.Now().UTC()) {
		t.Fatal("confirmed classification should make the task retryable")
	}
}

func TestNovelFetchSubmitChunksHonorsConfiguredBatchCap(t *testing.T) {
	chunks := novelFetchSubmitChunks([]string{"1", "2", "3", "4", "5"}, 2)
	if len(chunks) != 3 || len(chunks[0]) != 2 || len(chunks[1]) != 2 || len(chunks[2]) != 1 || chunks[2][0] != "5" {
		t.Fatalf("chunks=%#v", chunks)
	}
	if len(novelFetchSubmitChunks([]string{"1", "2"}, 0)) != 1 {
		t.Fatal("invalid batch size should use the default instead of disabling submission")
	}
}

type duplicateNovelFetchQueue struct{ recordingNovelFetchQueue }

func TestAppendWorkshopTaskLogPersistsTaskOperationLog(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(ctx, "alice", novelfetchworkshop.Document{BookID: "book-log", Meta: map[string]any{"bookName": "日志小说"}}); err != nil {
		t.Fatal(err)
	}

	if err := appendWorkshopTaskLog(ctx, store, "alice", "book-log", "process_completed", "处理完成"); err != nil {
		t.Fatal(err)
	}
	document, err := store.GetDocument(ctx, "alice", "book-log")
	if err != nil {
		t.Fatal(err)
	}
	if len(document.Logs) != 1 {
		t.Fatalf("logs=%#v", document.Logs)
	}
	log, ok := document.Logs[0].(map[string]any)
	if !ok || log["event"] != "process_completed" || log["detail"] != "处理完成" || strings.TrimSpace(firstText(log["time"])) == "" {
		t.Fatalf("persisted log=%#v", document.Logs[0])
	}
}

func TestAppendWorkshopTaskLogBoundsHistoryAndRedactsCredentials(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	logs := make([]any, 205)
	for index := range logs {
		logs[index] = map[string]any{"time": "2026-10-01T00:00:00Z", "event": "old", "detail": strconv.Itoa(index)}
	}
	if err := store.PutDocument(ctx, "alice", novelfetchworkshop.Document{BookID: "book-log-safe", Logs: logs}); err != nil {
		t.Fatal(err)
	}

	if err := appendWorkshopTaskLog(ctx, store, "alice", "book-log-safe", "fetch_failed", "Authorization: Bearer tok123 api_key=sekret Cookie: sid=abc"); err != nil {
		t.Fatal(err)
	}
	document, err := store.GetDocument(ctx, "alice", "book-log-safe")
	if err != nil {
		t.Fatal(err)
	}
	if len(document.Logs) != 200 {
		t.Fatalf("log history length=%d, want 200", len(document.Logs))
	}
	first, _ := document.Logs[0].(map[string]any)
	if first["detail"] != "6" {
		t.Fatalf("oldest retained event=%#v, want original event 6", first)
	}
	last, _ := document.Logs[len(document.Logs)-1].(map[string]any)
	lastDetail, _ := last["detail"].(string)
	for _, secret := range []string{"tok123", "sekret", "sid=abc"} {
		if strings.Contains(lastDetail, secret) {
			t.Fatalf("secret %q leaked into log detail %q", secret, lastDetail)
		}
	}
	if last["event"] != "fetch_failed" || !strings.Contains(lastDetail, "[已隐藏]") {
		t.Fatalf("last event=%#v, want redacted fetch failure", last)
	}
}

func TestNovelFetchAutomationStatusSeparatesConfigurationBlocksFromQueue(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(ctx, "alice", novelfetchworkshop.Document{BookID: "blocked", Meta: map[string]any{
		"status": "waiting_configuration", "siteSubmitStatus": "waiting_configuration",
	}}); err != nil {
		t.Fatal(err)
	}
	status, err := novelFetchAutomationStatus(ctx, store, newWorkshopProcessRegistry(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	counts := status["counts"].(map[string]int)
	if counts["waiting_configuration"] != 1 || status["pending_count"] != 0 || status["processing_count"] != 0 {
		t.Fatalf("configuration block must stay outside active queue: %#v", status)
	}
}

func TestNovelFetchAutomationStatusScopesCountsToBatchAndDate(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	historyBatch := fmt.Sprintf("batch-%d", time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC).UnixNano())
	currentBatch := fmt.Sprintf("batch-%d", time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC).UnixNano())
	if err := store.PutConfig(ctx, "alice", map[string]any{
		"novel_fetch_current_batch_id": currentBatch,
		"novel_fetch_automation":       map[string]any{"enabled": true},
	}); err != nil {
		t.Fatal(err)
	}
	for _, doc := range []novelfetchworkshop.Document{
		{BookID: "today-queued", Meta: map[string]any{"status": "queued", "currentBatchId": currentBatch}},
		{BookID: "today-running", Meta: map[string]any{"status": "running", "currentBatchId": currentBatch}},
		{BookID: "old-queued", Meta: map[string]any{"status": "queued", "currentBatchId": historyBatch}},
		{BookID: "old-running", Meta: map[string]any{"status": "processing", "currentBatchId": historyBatch}},
	} {
		if err := store.PutDocument(ctx, "alice", doc); err != nil {
			t.Fatal(err)
		}
	}

	status, err := novelFetchAutomationStatus(ctx, store, newWorkshopProcessRegistry(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	counts := status["counts"].(map[string]int)
	if counts["queued"] != 1 || counts["running"] != 1 || status["pending_count"] != 1 {
		t.Fatalf("default counts must only include the current batch: %#v", status)
	}
	if scope := status["scope"].(map[string]any); scope["batch_id"] != currentBatch {
		t.Fatalf("default scope=%#v, want current batch %q", scope, currentBatch)
	}

	history, err := novelFetchAutomationStatusScoped(ctx, store, newWorkshopProcessRegistry(), "alice", novelFetchQueueScope{Date: "2026-09-24"})
	if err != nil {
		t.Fatal(err)
	}
	historyCounts := history["counts"].(map[string]int)
	if historyCounts["queued"] != 1 || historyCounts["running"] != 1 || historyCounts["failed"] != 0 {
		t.Fatalf("date scope must only include that day's batch: %#v", history)
	}

	explicit, err := novelFetchAutomationStatusScoped(ctx, store, newWorkshopProcessRegistry(), "alice", novelFetchQueueScope{BatchID: historyBatch})
	if err != nil {
		t.Fatal(err)
	}
	if explicit["counts"].(map[string]int)["queued"] != 1 || explicit["counts"].(map[string]int)["running"] != 1 {
		t.Fatalf("explicit batch scope lost its own documents: %#v", explicit)
	}
}

func TestNovelFetchDispatcherUsesReferenceCadence(t *testing.T) {
	if novelFetchScheduleDispatchInterval != 5*time.Second {
		t.Fatalf("schedule interval=%s, want 5s", novelFetchScheduleDispatchInterval)
	}
	if novelFetchAutomationScanInterval != 20*time.Second {
		t.Fatalf("automation scan interval=%s, want 20s", novelFetchAutomationScanInterval)
	}
}

func (q *duplicateNovelFetchQueue) EnqueueUnique(context.Context, string, novelfetchworkshop.ProcessJob, time.Duration) (bool, error) {
	return false, nil
}

func (q *duplicateNovelFetchQueue) EnqueueUniqueBooks(_ context.Context, job novelfetchworkshop.ProcessJob, _ time.Duration) (bool, string, error) {
	return false, job.ID, nil
}

func (q *recordingNovelFetchQueue) Enqueue(_ context.Context, job novelfetchworkshop.ProcessJob) error {
	q.jobs = append(q.jobs, job)
	return nil
}

type uniqueRecordingNovelFetchQueue struct {
	recordingNovelFetchQueue
	keys       map[string]bool
	bookDedupe map[string]string
}

func (q *uniqueRecordingNovelFetchQueue) EnqueueUnique(ctx context.Context, key string, job novelfetchworkshop.ProcessJob, _ time.Duration) (bool, error) {
	if q.keys == nil {
		q.keys = map[string]bool{}
	}
	if q.keys[key] {
		return false, nil
	}
	q.keys[key] = true
	job.DedupeKey = key
	return true, q.Enqueue(ctx, job)
}

func (q *uniqueRecordingNovelFetchQueue) EnqueueUniqueBooks(ctx context.Context, job novelfetchworkshop.ProcessJob, _ time.Duration) (bool, string, error) {
	if q.bookDedupe == nil {
		q.bookDedupe = map[string]string{}
	}
	for _, bookID := range job.BookIDs {
		key := job.Owner + ":" + job.PlatformID + ":" + bookID
		if existingID := q.bookDedupe[key]; existingID != "" {
			return false, existingID, nil
		}
	}
	for _, bookID := range job.BookIDs {
		q.bookDedupe[job.Owner+":"+job.PlatformID+":"+bookID] = job.ID
	}
	job.DedupeKey = "books:" + job.Owner + ":" + job.PlatformID
	return true, job.ID, q.Enqueue(ctx, job)
}

func (q *uniqueRecordingNovelFetchQueue) ConflictingBookIDs(_ context.Context, job novelfetchworkshop.ProcessJob) (map[string]string, error) {
	conflicts := map[string]string{}
	for _, bookID := range job.BookIDs {
		if existingID := q.bookDedupe[job.Owner+":"+job.PlatformID+":"+bookID]; existingID != "" {
			conflicts[bookID] = existingID
		}
	}
	return conflicts, nil
}

func TestRecoverStaleNovelFetchScheduleReusesPersistedDispatchIdentity(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := &uniqueRecordingNovelFetchQueue{}
	now := time.Date(2026, 10, 17, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	old := now.Add(-10 * time.Minute).Format(time.RFC3339Nano)
	if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{map[string]any{
		"id": "schedule-recover", "status": "dispatching", "updatedAt": old, "dispatch_id": "dispatch-schedule-recover",
		"job_id": "go-schedule-stable-1", "batch_id": "batch-stable-1", "book_ids": []any{"12345678901"},
		"run_at": now.Add(-time.Hour).Format(time.RFC3339), "input_text": "12345678901\t恢复任务", "platform_id": "2", "max_txt": 4000,
	}}}); err != nil {
		t.Fatal(err)
	}

	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now); err != nil {
		t.Fatal(err)
	}
	if len(queue.jobs) != 1 || queue.jobs[0].ID != "go-schedule-stable-1" || queue.jobs[0].BatchID != "batch-stable-1" {
		t.Fatalf("recovery did not reuse the persisted job identity: %#v", queue.jobs)
	}
	settings, err := store.GetConfig(ctx, "tester")
	if err != nil {
		t.Fatal(err)
	}
	schedule := settings["novel_fetch_schedules"].([]any)[0].(map[string]any)
	if schedule["status"] != "queued" {
		t.Fatalf("recovered schedule status=%#v, want queued", schedule["status"])
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if len(queue.jobs) != 1 {
		t.Fatalf("recovered schedule was enqueued twice: %#v", queue.jobs)
	}
}

func TestRecoverStaleNovelFetchScheduleAcceptsExistingQueueDedupe(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := &duplicateNovelFetchQueue{}
	now := time.Date(2026, 10, 17, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{map[string]any{
		"id": "schedule-existing", "status": "dispatching", "updatedAt": now.Add(-10 * time.Minute).Format(time.RFC3339Nano),
		"dispatch_id": "dispatch-existing", "job_id": "go-schedule-existing", "batch_id": "batch-existing", "book_ids": []any{"12345678901"},
		"run_at": now.Add(-time.Hour).Format(time.RFC3339), "input_text": "12345678901\t已有队列任务", "platform_id": "2", "max_txt": 4000,
	}}}); err != nil {
		t.Fatal(err)
	}

	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now); err != nil {
		t.Fatal(err)
	}
	settings, err := store.GetConfig(ctx, "tester")
	if err != nil {
		t.Fatal(err)
	}
	schedule := settings["novel_fetch_schedules"].([]any)[0].(map[string]any)
	if schedule["status"] != "queued" || len(queue.jobs) != 0 {
		t.Fatalf("existing queued work was not reconciled safely: status=%#v jobs=%#v", schedule["status"], queue.jobs)
	}
}

func (q *recordingNovelFetchQueue) Claim(ctx context.Context, wait time.Duration) (novelfetchworkshop.ProcessJob, error) {
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return novelfetchworkshop.ProcessJob{}, ctx.Err()
	case <-timer.C:
		return novelfetchworkshop.ProcessJob{}, novelfetchworkshop.ErrQueueEmpty
	}
}

func (q *recordingNovelFetchQueue) Ack(context.Context, novelfetchworkshop.ProcessJob) error {
	return nil
}

func (q *recordingNovelFetchQueue) Fail(context.Context, novelfetchworkshop.ProcessJob, bool) error {
	return nil
}

func signNovelFetchRequest(req *http.Request, username string, isOwner bool, now time.Time, secret string) {
	issuedAt := strconv.FormatInt(now.Unix(), 10)
	owner := strconv.FormatBool(isOwner)
	payload := username + "\n" + issuedAt + "\n" + owner + "\n" + req.Method + "\n" + req.URL.Path
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	req.Header.Set(HeaderUsername, username)
	req.Header.Set(HeaderIsOwner, owner)
	req.Header.Set(HeaderIssuedAt, issuedAt)
	req.Header.Set(HeaderSignature, hex.EncodeToString(mac.Sum(nil)))
}

func doNovelFetchRequest(t *testing.T, handler http.Handler, method, path string, body any, now time.Time, secret string) *httptest.ResponseRecorder {
	t.Helper()
	var raw []byte
	if body != nil {
		var err error
		raw, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	signNovelFetchRequest(req, "tester", true, now, secret)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestNovelFetchWorkshopRealtimeStatusIsAvailableOnCanonicalRoute(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "book-realtime-status",
		Meta:   map[string]any{"aiStatus": "running"},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/realtime/status", nil, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("canonical realtime status=%d body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Counts map[string]int `json:"counts"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Counts["running"] != 1 {
		t.Fatalf("counts=%#v, want one running task", payload.Counts)
	}
}

func TestNovelFetchRealtimeStatusRestoresLatestJobFromPersistentRuns(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutRun(context.Background(), "tester", novelfetchworkshop.RunRecord{
		RunID: "queue_1792315781954_1d8ed3fc", BookID: "batch-old", Stage: "novel_fetch_process", Status: "done", TextModelID: "model-1", ModelID: "novel-fetch-workshop", ModelDisplayName: "小说获取完整流程", StartedAt: now.Add(-time.Minute).Format(time.RFC3339Nano), FinishedAt: now.Format(time.RFC3339Nano),
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/realtime/status", nil, now, secret)
	if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"id":"queue_1792315781954_1d8ed3fc"`)) || !bytes.Contains(response.Body.Bytes(), []byte(`"status":"done"`)) {
		t.Fatalf("persistent latest job missing: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestNovelFetchAutomationControlPersistsAndDispatchesPendingWork(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	queue := &recordingNovelFetchQueue{}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "12345678901",
		Meta:   map[string]any{"status": "input_ready", "platformId": "2", "currentBatchId": "batch-auto"},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue})
	started := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/automation/start", map[string]any{}, now, secret)
	if started.Code != http.StatusOK {
		t.Fatalf("start automation status=%d body=%s", started.Code, started.Body.String())
	}
	var status struct {
		AutoQueueEnabled bool   `json:"auto_queue_enabled"`
		ActiveJobID      string `json:"active_job_id"`
		PendingCount     int    `json:"pending_count"`
	}
	if err := json.Unmarshal(started.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if !status.AutoQueueEnabled || status.ActiveJobID == "" || status.PendingCount != 1 {
		t.Fatalf("unexpected started status: %#v", status)
	}
	if len(queue.jobs) != 1 {
		t.Fatalf("automatic work was not queued: %#v", queue.jobs)
	}
	job := queue.jobs[0]
	if !job.AutoSubmit || len(job.BookIDs) != 1 || job.BookIDs[0] != "12345678901" {
		t.Fatalf("unexpected automatic job: %#v", job)
	}
	settings, err := store.GetConfig(context.Background(), "tester")
	if err != nil {
		t.Fatal(err)
	}
	automation := mapOrEmpty(settings["novel_fetch_automation"])
	if automation["enabled"] != true {
		t.Fatalf("automation setting was not persisted: %#v", automation)
	}
	paused := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/automation/pause", map[string]any{}, now, secret)
	if paused.Code != http.StatusOK || !bytes.Contains(paused.Body.Bytes(), []byte(`"auto_queue_enabled":false`)) {
		t.Fatalf("pause automation status=%d body=%s", paused.Code, paused.Body.String())
	}
}

func TestNovelFetchAutomationDoesNotRedispatchBeforeRetryTime(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	now := time.Date(2026, 10, 18, 20, 0, 0, 0, time.UTC)
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"novel_fetch_automation": map[string]any{"enabled": true},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "22345678901",
		Meta: map[string]any{
			"status": "completed", "siteSubmitStatus": "pending_upload",
			"automationNextRetryAt": now.Add(time.Minute).Format(time.RFC3339Nano),
		},
	}); err != nil {
		t.Fatal(err)
	}
	dispatched, err := dispatchNovelFetchAutomationOwner(context.Background(), store, queue, "tester", now)
	if err != nil {
		t.Fatal(err)
	}
	if dispatched != nil {
		t.Fatalf("task was dispatched before retry time: %#v", dispatched)
	}
	if _, err := queue.Claim(context.Background(), 20*time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
		t.Fatalf("queue should stay empty before retry time: %v", err)
	}
}

func TestNovelFetchAutomationDispatchesDueFetchRetryWhenGlobalAutomationIsPaused(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	queue := &recordingNovelFetchQueue{}
	now := time.Date(2026, 10, 27, 1, 0, 0, 0, time.UTC)
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"novel_fetch_automation": map[string]any{"enabled": false},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "due-fetch-retry",
		Meta: map[string]any{
			"status": "waiting_retry", "queueState": "waiting_retry", "originalStatus": "waiting_fetch_retry",
			"originalFetchNextRetryAt": now.Add(-time.Minute).Format(time.RFC3339Nano),
		},
	}); err != nil {
		t.Fatal(err)
	}
	job, err := dispatchNovelFetchAutomationOwner(context.Background(), store, queue, "tester", now)
	if err != nil {
		t.Fatal(err)
	}
	if job == nil || len(job.BookIDs) != 1 || job.BookIDs[0] != "due-fetch-retry" {
		t.Fatalf("due fetch retry must dispatch while automation is paused: %#v", job)
	}
}

func TestNovelFetchAutomationDispatchesCompletedTaskWithoutSubmissionState(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	queue := &recordingNovelFetchQueue{}
	now := time.Date(2026, 10, 19, 1, 20, 0, 0, time.UTC)
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"novel_fetch_automation": map[string]any{"enabled": true},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "completed-without-submission",
		Meta:   map[string]any{"status": "completed", "originalStatus": "done", "aiStatus": "ai_done"},
	}); err != nil {
		t.Fatal(err)
	}

	job, err := dispatchNovelFetchAutomationOwner(context.Background(), store, queue, "tester", now)
	if err != nil {
		t.Fatal(err)
	}
	if job == nil || !job.AutoSubmit || len(job.BookIDs) != 1 || job.BookIDs[0] != "completed-without-submission" {
		t.Fatalf("completed task without a submission state must re-enter automatic submission: %#v", job)
	}
}

func TestNovelFetchAutomationUsesPersistedSensitiveAISetting(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	queue := &recordingNovelFetchQueue{}
	now := time.Date(2026, 10, 19, 1, 20, 0, 0, time.UTC)
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"novel_fetch_automation": map[string]any{"enabled": true},
		"sensitive_ai":           map[string]any{"enabled": true},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "sensitive-ai-automatic",
		Meta:   map[string]any{"status": "completed", "originalStatus": "done", "aiStatus": "ai_done"},
	}); err != nil {
		t.Fatal(err)
	}
	job, err := dispatchNovelFetchAutomationOwner(context.Background(), store, queue, "tester", now)
	if err != nil {
		t.Fatal(err)
	}
	if job == nil || !job.SensitiveAIEnabled {
		t.Fatalf("automatic job must inherit sensitive AI setting: %#v", job)
	}
}

func TestRecoverStaleNovelFetchTaskMakesItRetryableAndClosesRun(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	now := time.Date(2026, 10, 19, 1, 20, 0, 0, time.UTC)
	old := now.Add(-10 * time.Minute).Format(time.RFC3339Nano)
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID:    "stale-book",
		UpdatedAt: old,
		Meta:      map[string]any{"status": "processing", "queueState": "processing", "currentBatchId": "stale-batch"},
	}); err != nil {
		t.Fatal(err)
	}
	staleDocument, err := store.GetDocument(context.Background(), "tester", "stale-book")
	if err != nil {
		t.Fatal(err)
	}
	staleDocument.UpdatedAt = old
	recoverStaleNovelFetchTasks(context.Background(), store, "tester", map[string]any{"workflow": map[string]any{"auto_queue_task_timeout_seconds": 180}}, []novelfetchworkshop.Document{staleDocument}, now)
	document, err := store.GetDocument(context.Background(), "tester", "stale-book")
	if err != nil {
		t.Fatal(err)
	}
	if document.Meta["status"] != "waiting_retry" {
		t.Fatalf("stale document status=%#v, want waiting_retry", document.Meta["status"])
	}
}

type staleRecoveryRecordingQueue struct {
	recordingNovelFetchQueue
	jobIDs []string
}

func (q *staleRecoveryRecordingQueue) RecoverStaleJob(_ context.Context, jobID string) (bool, error) {
	q.jobIDs = append(q.jobIDs, jobID)
	return true, nil
}

func TestRecoverStaleNovelFetchTaskReleasesItsOldQueueJob(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	now := time.Date(2026, 10, 19, 1, 20, 0, 0, time.UTC)
	old := now.Add(-10 * time.Minute).Format(time.RFC3339Nano)
	if err := store.PutDocument(ctx, "tester", novelfetchworkshop.Document{
		BookID: "stale-book-with-job-id", UpdatedAt: old,
		Meta: map[string]any{"status": "processing", "queueState": "processing", "currentBatchId": "stale-batch", "currentProcessJobId": "stale-job-id"},
	}); err != nil {
		t.Fatal(err)
	}
	document, err := store.GetDocument(ctx, "tester", "stale-book-with-job-id")
	if err != nil {
		t.Fatal(err)
	}
	document.UpdatedAt = old
	queue := &staleRecoveryRecordingQueue{}
	recoverStaleNovelFetchTasks(ctx, store, "tester", map[string]any{"workflow": map[string]any{"auto_queue_task_timeout_seconds": 180}}, []novelfetchworkshop.Document{document}, now, queue)
	if len(queue.jobIDs) != 1 || queue.jobIDs[0] != "stale-job-id" {
		t.Fatalf("stale queue job was not released: %#v", queue.jobIDs)
	}
	updated, err := store.GetDocument(ctx, "tester", document.BookID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Meta["status"] != "waiting_retry" || !novelFetchDocumentHasLogEvent(updated, "process_recovered") {
		t.Fatalf("stale task recovery state/log missing: %#v", updated)
	}
}

func TestNovelFetchAutomationDispatchesStoredAIWhenStatusMarkerIsStale(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	queue := &recordingNovelFetchQueue{}
	now := time.Date(2026, 10, 19, 1, 20, 0, 0, time.UTC)
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"novel_fetch_automation": map[string]any{"enabled": true},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID:   "stored-ai-with-stale-status",
		Versions: map[string]any{"ai1": "已经持久化的 AI 正文"},
		Meta:     map[string]any{"status": "completed", "originalStatus": "done", "aiStatus": "未开始"},
	}); err != nil {
		t.Fatal(err)
	}

	job, err := dispatchNovelFetchAutomationOwner(context.Background(), store, queue, "tester", now)
	if err != nil {
		t.Fatal(err)
	}
	if job == nil || !job.AutoSubmit || len(job.BookIDs) != 1 || job.BookIDs[0] != "stored-ai-with-stale-status" {
		t.Fatalf("stored AI output with stale status must re-enter automatic submission: %#v", job)
	}
}

func TestNovelFetchAutomationHydratesStoredBodyRefsWhenDocumentIsLightweight(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	queue := &recordingNovelFetchQueue{}
	now := time.Date(2026, 10, 19, 1, 20, 0, 0, time.UTC)
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"novel_fetch_automation": map[string]any{"enabled": true},
	}); err != nil {
		t.Fatal(err)
	}
	const bookID = "lightweight-stored-ai"
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: bookID,
		Meta:   map[string]any{"status": "completed", "originalStatus": "done", "aiStatus": "未开始"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{
		BookID:  bookID,
		BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1", State: "ready"},
		Content: "已经持久化到正文表的 AI 正文",
	}); err != nil {
		t.Fatal(err)
	}

	job, err := dispatchNovelFetchAutomationOwner(context.Background(), store, queue, "tester", now)
	if err != nil {
		t.Fatal(err)
	}
	if job == nil || !job.AutoSubmit || len(job.BookIDs) != 1 || job.BookIDs[0] != bookID {
		t.Fatalf("lightweight document with stored AI body must enter automatic submission: %#v", job)
	}
}

func TestNovelFetchAutomationDoesNotTreatEmptyAIAsReady(t *testing.T) {
	document := novelfetchworkshop.Document{
		BookID:   "empty-ai",
		Versions: map[string]any{"ai1": "  "},
		BodyRefs: map[string]novelfetchworkshop.BodyRef{"ai1": {State: "ready", CharCount: 0}},
		Meta:     map[string]any{"status": "completed", "originalStatus": "done", "aiStatus": "未开始"},
	}
	if novelFetchAutomationEligible(document, time.Now().UTC()) {
		t.Fatalf("empty AI output must not be treated as ready: %#v", document)
	}
}

func TestNovelFetchAutomationDedupedJobDoesNotLeaveFalseQueuedStatus(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	queue := &duplicateNovelFetchQueue{}
	now := time.Date(2026, 10, 18, 20, 0, 0, 0, time.UTC)
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"novel_fetch_automation": map[string]any{"enabled": true},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "dedupe-pending-upload", Meta: map[string]any{"status": "completed", "siteSubmitStatus": "pending_upload"},
	}); err != nil {
		t.Fatal(err)
	}
	job, err := dispatchNovelFetchAutomationOwner(context.Background(), store, queue, "tester", now)
	if err != nil || job != nil {
		t.Fatalf("deduped job should not create a second queue item: job=%#v err=%v", job, err)
	}
	document, err := store.GetDocument(context.Background(), "tester", "dedupe-pending-upload")
	if err != nil {
		t.Fatal(err)
	}
	if document.Meta["status"] != "completed" || document.Meta["queueState"] != nil {
		t.Fatalf("deduped job must keep its retryable state instead of a false queue state: %#v", document.Meta)
	}
}

func TestNovelFetchAutomationConfigurationBlockDoesNotRetry(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	const owner = "tester"
	const bookID = "blocked-style-book"
	if err := store.PutDocument(context.Background(), owner, novelfetchworkshop.Document{BookID: bookID, Meta: map[string]any{
		"status": "completed", "siteSubmitStatus": "pending_upload",
		"siteSubmitFailedErrors": map[string]any{"ai1": "无法映射网站字段：风格类型"},
	}}); err != nil {
		t.Fatal(err)
	}
	markNovelFetchAutomationSubmitResult(context.Background(), store, owner, []string{bookID}, errors.New("1 个版本待上传"))
	document, err := store.GetDocument(context.Background(), owner, bookID)
	if err != nil {
		t.Fatal(err)
	}
	if document.Meta["siteSubmitStatus"] != "waiting_configuration" || document.Meta["status"] != "waiting_configuration" {
		t.Fatalf("configuration blocker must wait for user configuration: %#v", document.Meta)
	}
	if _, exists := document.Meta["automationNextRetryAt"]; exists {
		t.Fatalf("configuration blocker must not receive a retry time: %#v", document.Meta)
	}
	if novelFetchAutomationEligible(document, time.Now().UTC().Add(time.Hour)) {
		t.Fatalf("configuration blocker must not re-enter the automatic queue: %#v", document.Meta)
	}
}

func TestNovelFetchAutomationTemporarySubmitFailureKeepsBackoffRetry(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	const owner = "tester"
	const bookID = "retry-upload-book"
	if err := store.PutDocument(context.Background(), owner, novelfetchworkshop.Document{BookID: bookID, Meta: map[string]any{"status": "completed"}}); err != nil {
		t.Fatal(err)
	}
	markNovelFetchAutomationSubmitResult(context.Background(), store, owner, []string{bookID}, errors.New("上传连接暂时超时"))
	document, err := store.GetDocument(context.Background(), owner, bookID)
	if err != nil {
		t.Fatal(err)
	}
	if document.Meta["siteSubmitStatus"] != "pending_upload" || firstText(document.Meta["automationNextRetryAt"]) == "" {
		t.Fatalf("temporary submit failure must retain delayed retry: %#v", document.Meta)
	}
}

func TestWorkshopWorkerClassifiesReadyOriginalWithoutRefetching(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	const owner = "tester"
	const bookID = "ready-needs-classification"
	if err := store.PutDocument(context.Background(), owner, novelfetchworkshop.Document{
		BookID: bookID, Gender: "待确认", Original: "已有原文", ProcessedText: "已有原文",
		BookInfo: novelfetchworkshop.BookInfo{Category: "男频", Genre: "古风通用"},
		Meta:     map[string]any{"originalStatus": "done", "style": "待判定"},
	}); err != nil {
		t.Fatal(err)
	}
	worker := &workshopProcessWorker{store: store, queue: novelfetchworkshop.NewMemoryQueue(), registry: newWorkshopProcessRegistry()}
	if err := worker.execute(context.Background(), novelfetchworkshop.ProcessJob{Owner: owner, BookIDs: []string{bookID}}); err != nil {
		t.Fatal(err)
	}
	document, err := store.GetDocument(context.Background(), owner, bookID)
	if err != nil {
		t.Fatal(err)
	}
	if document.Gender != "男频" || firstText(document.Meta["style"]) != "古风通用" {
		t.Fatalf("ready original was not classified without refetch: %+v", document)
	}
}

func TestNovelFetchOriginalFetchRetryIntervalUsesFetchConfiguration(t *testing.T) {
	delay := novelFetchOriginalFetchRetryInterval(map[string]any{
		"fetch": map[string]any{"original_fetch_retry_interval_minutes": 15},
	})
	if delay != 15*time.Minute {
		t.Fatalf("fetch retry interval = %s, want 15m", delay)
	}
}

func TestSetSensitiveDetailsCountsFailures(t *testing.T) {
	meta := map[string]any{}
	setSensitiveDetails(meta, []map[string]any{{"status": "done"}, {"status": "failed"}}, 2)
	fixed := meta["sensitiveFixed"].(map[string]any)
	if fixed["fixed_count"] != 1 || fixed["failed_count"] != 1 {
		t.Fatalf("sensitive counts=%#v", fixed)
	}
}

func TestNovelFetchWorkshopTaskListBackfillsPlatformAndGender(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID:   "2093734568914334892",
		Gender:   "待确认",
		BookInfo: novelfetchworkshop.BookInfo{Category: "男频", Genre: "都市"},
		Meta:     map[string]any{"platformId": "1", "originalStatus": "done"},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks", nil, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("tasks status=%d body=%s", response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"platform_name":"黑岩付费"`)) {
		t.Fatalf("platform name was not backfilled: %s", response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"gender":"男频"`)) || !bytes.Contains(response.Body.Bytes(), []byte(`"classify_status":"classified"`)) {
		t.Fatalf("gender was not backfilled: %s", response.Body.String())
	}
}

func TestNovelFetchWorkshopTaskListBackfillsDeterministicStyle(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID:   "book-style-backfill",
		Gender:   "男频",
		BookInfo: novelfetchworkshop.BookInfo{Category: "男频", Genre: "古风通用"},
		Meta:     map[string]any{"platformId": "1", "originalStatus": "done"},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks", nil, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("tasks status=%d body=%s", response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"style":"古风通用"`)) {
		t.Fatalf("deterministic style was not backfilled: %s", response.Body.String())
	}
}

func TestNovelFetchWorkshopSparsePutPreservesExistingGenderSource(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "book-existing-gender", Gender: "女频", GenderSource: novelfetchworkshop.GenderAI, GenderReason: "AI 判断",
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	put := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/tasks/book-existing-gender", map[string]any{"bookInfo": map[string]string{"category": "女频"}}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("sparse put status=%d body=%s", put.Code, put.Body.String())
	}
	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/book-existing-gender", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	var document novelfetchworkshop.Document
	if err := json.Unmarshal(get.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document.Gender != "女频" || document.GenderSource != novelfetchworkshop.GenderAI {
		t.Fatalf("sparse put changed existing gender source: %#v", document)
	}
}

func TestNovelFetchWorkshopPutResolvesPendingGenderFromBookInfo(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "book-pending-gender", Gender: "待确认",
		BookInfo: novelfetchworkshop.BookInfo{Category: "男频", Genre: "都市"},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	put := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/tasks/book-pending-gender", map[string]any{
		"bookInfo": map[string]string{"category": "男频", "genre": "都市"},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}
	stored, err := store.GetDocument(context.Background(), "tester", "book-pending-gender")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Gender != "男频" || stored.GenderSource != novelfetchworkshop.GenderBookInfo {
		t.Fatalf("pending gender was not resolved during put: %#v", stored)
	}
	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/book-pending-gender", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	var document novelfetchworkshop.Document
	if err := json.Unmarshal(get.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document.Gender != "男频" || document.GenderSource != novelfetchworkshop.GenderBookInfo {
		t.Fatalf("pending gender was not retained from book info: %#v", document)
	}
}

func TestNovelFetchWorkshopBatchRetryUsesGoStore(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID:        "already-done",
		Versions:      map[string]any{"ai1": "已有 AI1 正文"},
		Original:      "已有正文",
		ProcessedText: "已有正文",
		Meta:          map[string]any{"originalStatus": "done", "aiStatus": "done"},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/batch-retry", map[string]any{"ids": []string{"already-done"}}, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("batch retry status=%d body=%s", response.Code, response.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["skipped"] != float64(1) || payload["retried"] != float64(0) {
		t.Fatalf("unexpected batch retry result: %#v", payload)
	}
}

func TestNovelFetchWorkshopBatchRetryReturnsAffectedTaskIDs(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "already-done", Original: "已有正文", ProcessedText: "已有正文",
		Versions: map[string]any{"ai1": "已有 AI1 正文"},
		Meta:     map[string]any{"originalStatus": "done", "aiStatus": "done"},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/batch-retry", map[string]any{"ids": []string{"already-done", "already-done"}}, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("batch retry status=%d body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Requested int      `json:"requested"`
		Updated   int      `json:"updated"`
		Skipped   int      `json:"skipped"`
		Failed    int      `json:"failed"`
		TaskIDs   []string `json:"task_ids"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Requested != 1 || payload.Updated != 0 || payload.Skipped != 1 || payload.Failed != 0 || len(payload.TaskIDs) != 0 {
		t.Fatalf("batch retry affected result=%#v", payload)
	}
}

func TestNovelFetchWorkshopPutPreservesExistingPersistentFields(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})
	put := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/tasks/book-preserve", map[string]any{
		"originalRaw": "完整原文", "processedText": "处理正文", "bookInfo": map[string]string{"category": "男频", "genre": "都市"},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("initial put status=%d body=%s", put.Code, put.Body.String())
	}
	put = doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/tasks/book-preserve", map[string]any{"gender": "女频"}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("sparse put status=%d body=%s", put.Code, put.Body.String())
	}
	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/book-preserve", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	var document novelfetchworkshop.Document
	if err := json.Unmarshal(get.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document.OriginalRaw != "完整原文" || document.ProcessedText != "处理正文" || document.BookInfo.Category != "男频" || document.Gender != "女频" {
		t.Fatalf("sparse put lost existing fields: %#v", document)
	}
}

func TestNovelFetchWorkshopBridgeRoundTrip(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{
		BridgeSecret:    secret,
		Now:             func() time.Time { return now },
		NovelFetchStore: store,
	})

	put := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/tasks/book-1", map[string]any{
		"meta":        map[string]any{"bookName": "测试小说", "targetVersions": []any{"original", "ai1", "ai3"}},
		"original":    "正文",
		"originalRaw": "原始正文",
		"versions":    map[string]any{"ai1": "AI1", "ai3": "AI3"},
		"logs":        []any{map[string]any{"event": "saved"}},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}

	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/book-1", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	var document map[string]any
	if err := json.Unmarshal(get.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document["bookId"] != "book-1" || document["original"] != "正文" {
		t.Fatalf("unexpected document: %#v", document)
	}
	versions := document["versions"].(map[string]any)
	if versions["ai1"] != "AI1" || versions["ai3"] != "AI3" {
		t.Fatalf("sparse versions lost: %#v", versions)
	}

	list := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks", nil, now, secret)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	var listed struct {
		Tasks []novelfetchworkshop.Document `json:"tasks"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Tasks) != 1 || listed.Tasks[0].BookID != "book-1" {
		t.Fatalf("unexpected tasks: %#v", listed.Tasks)
	}

	configPut := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/config", map[string]any{
		"settings": map[string]any{"fetch": map[string]any{"timeout_seconds": 9}},
	}, now, secret)
	if configPut.Code != http.StatusOK {
		t.Fatalf("config put status=%d body=%s", configPut.Code, configPut.Body.String())
	}
	configGet := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/config", nil, now, secret)
	if configGet.Code != http.StatusOK {
		t.Fatalf("config get status=%d body=%s", configGet.Code, configGet.Body.String())
	}
	var config struct {
		Settings map[string]any `json:"settings"`
	}
	if err := json.Unmarshal(configGet.Body.Bytes(), &config); err != nil {
		t.Fatal(err)
	}
	if config.Settings["fetch"].(map[string]any)["timeout_seconds"] != float64(9) {
		t.Fatalf("config did not round trip: %#v", config.Settings)
	}

	deleted := doNovelFetchRequest(t, handler, http.MethodDelete, "/api/novel-fetch-workshop/tasks", map[string]any{"ids": []string{"book-1"}}, now, secret)
	if deleted.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", deleted.Code, deleted.Body.String())
	}
	missing := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/book-1", nil, now, secret)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing status=%d body=%s", missing.Code, missing.Body.String())
	}
}

func TestRunAuditPersistsSelectedAndResolvedModelWithoutCredential(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})

	put := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/runs", map[string]any{
		"runId": "run-1", "bookId": "book-1", "stage": "rewrite", "status": "succeeded", "attempts": 1,
		"textModelId": "gemini-3", "modelId": "gemini-3", "modelDisplayName": "Gemini 3", "api_key": "must-not-persist",
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}

	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/runs?bookId=book-1", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	if bytes.Contains(get.Body.Bytes(), []byte("must-not-persist")) || bytes.Contains(bytes.ToLower(get.Body.Bytes()), []byte("api_key")) {
		t.Fatalf("audit leaked secret: %s", get.Body.String())
	}
	var response struct {
		Runs []novelfetchworkshop.RunRecord `json:"runs"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Runs) != 1 || response.Runs[0].TextModelID != "gemini-3" || response.Runs[0].ModelID != "gemini-3" {
		t.Fatalf("unexpected audit records: %#v", response.Runs)
	}
}

func TestNovelFetchWorkshopRejectsUnsignedRequest(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: "test-secret", NovelFetchStore: store})
	req := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-workshop/tasks", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestNovelFetchWorkshopAcceptsBrowserSessionToken(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	store := novelfetchworkshop.NewMemoryStore()
	authStore := &authTestStore{
		account: auth.Account{Username: "alice", Active: true, IsOwner: true},
		session: auth.Session{Token: "browser-token", Username: "alice", ExpiresAt: now.Add(time.Hour)},
	}
	handler := NewRouter(RouterOptions{
		BridgeSecret:    "unused-bridge-secret",
		Now:             func() time.Time { return now },
		AuthStore:       authStore,
		NovelFetchStore: store,
	})
	req := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-workshop/tasks", nil)
	req.Header.Set("Authorization", "Bearer browser-token")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("browser session status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestNovelFetchWorkshopConfigMatchesFrontendContract(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})
	put := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/config", map[string]any{
		"appConfig": map[string]any{"workflow": map[string]any{"auto_fetch_original": true}},
		"platforms": []any{map[string]any{"id": "2", "name": "起点"}},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("config put status=%d body=%s", put.Code, put.Body.String())
	}
	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/config", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("config get status=%d body=%s", get.Code, get.Body.String())
	}
	var response struct {
		AppConfig map[string]any   `json:"appConfig"`
		Platforms []map[string]any `json:"platforms"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	workflow, ok := response.AppConfig["workflow"].(map[string]any)
	if !ok || workflow["auto_fetch_original"] != true || len(response.Platforms) != 1 || response.Platforms[0]["name"] != "起点" {
		t.Fatalf("frontend config contract mismatch: %#v", response)
	}
	platforms := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/platforms", nil, now, secret)
	if platforms.Code != http.StatusOK || !bytes.Contains(platforms.Body.Bytes(), []byte(`"name":"起点"`)) {
		t.Fatalf("platforms status=%d body=%s", platforms.Code, platforms.Body.String())
	}
}

func TestNovelFetchWorkshopProcessCreatesTasksFromPastedList(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	process := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process", map[string]any{
		"inputText": "书籍ID\t书名\t男女频\n12345678901\t第一本\t男频\n12345678901\t第一本\t男频\n22345678901\t第二本\t女频",
	}, now, secret)
	if process.Code != http.StatusOK {
		t.Fatalf("process status=%d body=%s", process.Code, process.Body.String())
	}
	var response struct {
		UniqueTasks    int                           `json:"uniqueTasks"`
		DuplicateCount int                           `json:"duplicateCount"`
		Tasks          []novelfetchworkshop.Document `json:"tasks"`
	}
	if err := json.Unmarshal(process.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.UniqueTasks != 2 || response.DuplicateCount != 1 || len(response.Tasks) != 2 {
		t.Fatalf("unexpected process response: %#v", response)
	}
}

func TestNovelFetchWorkshopCurrentBatchOnlyContainsLatestInput(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	first := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process", map[string]any{"inputText": "12345678901\t旧批次"}, now, secret)
	if first.Code != http.StatusOK {
		t.Fatalf("first process status=%d body=%s", first.Code, first.Body.String())
	}
	second := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process", map[string]any{"inputText": "22345678901\t当前批次"}, now, secret)
	if second.Code != http.StatusOK {
		t.Fatalf("second process status=%d body=%s", second.Code, second.Body.String())
	}
	current := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/batches/current", nil, now, secret)
	if current.Code != http.StatusOK {
		t.Fatalf("current status=%d body=%s", current.Code, current.Body.String())
	}
	var response struct {
		Batch struct {
			TaskIDs []string `json:"taskIds"`
		} `json:"batch"`
	}
	if err := json.Unmarshal(current.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Batch.TaskIDs) != 1 || response.Batch.TaskIDs[0] != "22345678901" {
		t.Fatalf("current batch leaked historical tasks: %#v", response.Batch.TaskIDs)
	}
}

func TestNovelFetchWorkshopBatchHistoryRerunAndStopStayOnGoStore(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	first := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process", map[string]any{"inputText": "12345678901\t第一批"}, now, secret)
	if first.Code != http.StatusOK {
		t.Fatalf("first process status=%d body=%s", first.Code, first.Body.String())
	}
	second := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process", map[string]any{"inputText": "22345678901\t第二批"}, now, secret)
	if second.Code != http.StatusOK {
		t.Fatalf("second process status=%d body=%s", second.Code, second.Body.String())
	}
	var firstResult struct {
		BatchID string `json:"batch_id"`
	}
	if err := json.Unmarshal(first.Body.Bytes(), &firstResult); err != nil {
		t.Fatal(err)
	}
	var secondResult struct {
		BatchID string `json:"batch_id"`
	}
	if err := json.Unmarshal(second.Body.Bytes(), &secondResult); err != nil {
		t.Fatal(err)
	}
	history := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/batches", nil, now, secret)
	if history.Code != http.StatusOK {
		t.Fatalf("history status=%d body=%s", history.Code, history.Body.String())
	}
	var historyResult struct {
		Batches []struct {
			ID      string   `json:"id"`
			TaskIDs []string `json:"taskIds"`
		} `json:"batches"`
	}
	if err := json.Unmarshal(history.Body.Bytes(), &historyResult); err != nil {
		t.Fatal(err)
	}
	if len(historyResult.Batches) < 2 || firstResult.BatchID == "" || secondResult.BatchID == "" {
		t.Fatalf("batch history missing: first=%q second=%q history=%#v", firstResult.BatchID, secondResult.BatchID, historyResult.Batches)
	}
	rerun := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/batches/"+firstResult.BatchID+"/rerun", map[string]any{"mode": "all"}, now, secret)
	if rerun.Code != http.StatusOK || !bytes.Contains(rerun.Body.Bytes(), []byte(`"12345678901"`)) {
		t.Fatalf("rerun status=%d body=%s", rerun.Code, rerun.Body.String())
	}
	stop := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/stop-selected", map[string]any{"ids": []string{"22345678901"}}, now, secret)
	if stop.Code != http.StatusOK || !bytes.Contains(stop.Body.Bytes(), []byte(`"updated":1`)) {
		t.Fatalf("stop status=%d body=%s", stop.Code, stop.Body.String())
	}
	document, err := store.GetDocument(context.Background(), "tester", "22345678901")
	if err != nil {
		t.Fatal(err)
	}
	if document.Meta["status"] != "stopped" {
		t.Fatalf("stop did not persist status: %#v", document.Meta)
	}
}

func TestWorkshopProcessWorkerDoesNotRestartStoppedTask(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	const owner = "tester"
	const bookID = "22345678901"
	if err := store.PutDocument(context.Background(), owner, novelfetchworkshop.Document{
		BookID: bookID,
		Meta:   map[string]any{"status": "stopped", "queueState": "stopped"},
	}); err != nil {
		t.Fatal(err)
	}
	worker := &workshopProcessWorker{
		store:    store,
		queue:    novelfetchworkshop.NewMemoryQueue(),
		registry: newWorkshopProcessRegistry(),
	}
	worker.process(context.Background(), novelfetchworkshop.ProcessJob{
		ID: "job-stopped", Owner: owner, BatchID: "batch-stopped", BookIDs: []string{bookID},
	})
	document, err := store.GetDocument(context.Background(), owner, bookID)
	if err != nil {
		t.Fatal(err)
	}
	if document.Meta["status"] != "stopped" || document.Meta["queueState"] != "stopped" {
		t.Fatalf("stopped task was restarted by worker: %#v", document.Meta)
	}
}

func TestWorkshopProcessWorkerIsolatesFailedBookAndDefersOnlyThatBook(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	for _, bookID := range []string{"11111111111", "22222222222"} {
		if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: bookID, Meta: map[string]any{"platformId": "2"}}); err != nil {
			t.Fatal(err)
		}
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("bookid") == "11111111111" {
			http.Error(w, "temporary failure", http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": "可正常处理的正文", "bookinfo": map[string]any{"book_name": "成功书"}})
	}))
	defer upstream.Close()
	worker := &workshopProcessWorker{store: store, queue: queue, registry: newWorkshopProcessRegistry(), fetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}}
	worker.process(context.Background(), novelfetchworkshop.ProcessJob{ID: "job-isolation", Owner: "tester", BatchID: "batch-isolation", BookIDs: []string{"11111111111", "22222222222"}, MaxTxt: 4000})

	failed, _ := store.GetDocument(context.Background(), "tester", "11111111111")
	succeeded, _ := store.GetDocument(context.Background(), "tester", "22222222222")
	if failed.Meta["status"] != "waiting_retry" || failed.Meta["originalStatus"] != "waiting_fetch_retry" || succeeded.Meta["status"] != "completed" {
		t.Fatalf("per-book status was not isolated: failed=%#v succeeded=%#v", failed.Meta, succeeded.Meta)
	}
	if _, err := queue.Claim(context.Background(), 50*time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
		t.Fatalf("failed original fetch must wait before it is retried: %v", err)
	}
}

func TestWorkshopProcessWorkerHonorsConfiguredFetchConcurrency(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	owner := "tester"
	if err := store.PutConfig(ctx, owner, map[string]any{"fetch": map[string]any{"concurrency": 3}}); err != nil {
		t.Fatal(err)
	}
	bookIDs := []string{"12345678901", "12345678902", "12345678903"}
	for _, bookID := range bookIDs {
		if err := store.PutDocument(ctx, owner, novelfetchworkshop.Document{BookID: bookID, Meta: map[string]any{"platformId": "2"}}); err != nil {
			t.Fatal(err)
		}
	}
	var active atomic.Int32
	var maximum atomic.Int32
	twoRequests := make(chan struct{})
	var releaseRequests sync.Once
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		current := active.Add(1)
		for previous := maximum.Load(); current > previous && !maximum.CompareAndSwap(previous, current); previous = maximum.Load() {
		}
		if current >= 2 {
			releaseRequests.Do(func() { close(twoRequests) })
		}
		select {
		case <-twoRequests:
		case <-time.After(2 * time.Second):
		}
		active.Add(-1)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": "并发抓取正文", "bookinfo": map[string]any{"category": "男频", "genre": "都市"}})
	}))
	defer upstream.Close()
	worker := &workshopProcessWorker{store: store, queue: queue, registry: newWorkshopProcessRegistry(), fetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}}
	worker.process(ctx, novelfetchworkshop.ProcessJob{ID: "job-fetch-concurrency", Owner: owner, BatchID: "batch-fetch-concurrency", BookIDs: bookIDs, MaxTxt: 4000})
	if maximum.Load() < 2 {
		t.Fatalf("maximum concurrent upstream requests=%d, want at least 2", maximum.Load())
	}
}

func TestWorkshopProcessWorkerDefersOriginalFetchFailureBeforeAutoRetry(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	const owner = "tester"
	const bookID = "22345678901"
	if err := store.PutConfig(context.Background(), owner, map[string]any{
		"novel_fetch_automation": map[string]any{"enabled": true},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), owner, novelfetchworkshop.Document{BookID: bookID, Meta: map[string]any{"platformId": "2"}}); err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "no content", http.StatusBadGateway)
	}))
	defer upstream.Close()
	worker := &workshopProcessWorker{store: store, queue: queue, registry: newWorkshopProcessRegistry(), fetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}}

	worker.process(context.Background(), novelfetchworkshop.ProcessJob{ID: "job-fetch-wait", Owner: owner, BatchID: "batch-fetch-wait", BookIDs: []string{bookID}, MaxTxt: 4000})

	document, err := store.GetDocument(context.Background(), owner, bookID)
	if err != nil {
		t.Fatal(err)
	}
	if document.Meta["status"] != "waiting_retry" || document.Meta["originalStatus"] != "waiting_fetch_retry" {
		t.Fatalf("fetch failure should wait for a later retry: %#v", document.Meta)
	}
	nextRetry, err := time.Parse(time.RFC3339Nano, firstText(document.Meta["originalFetchNextRetryAt"]))
	if err != nil {
		t.Fatalf("missing retry time: %#v", document.Meta)
	}
	remaining := time.Until(nextRetry)
	if remaining < 59*time.Minute || remaining > 61*time.Minute {
		t.Fatalf("retry delay = %s, want about one hour", remaining)
	}
	if _, err := queue.Claim(context.Background(), 20*time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
		t.Fatalf("original fetch failure was immediately queued again: %v", err)
	}
	if dispatched, err := dispatchNovelFetchAutomationOwner(context.Background(), store, queue, owner, time.Now().UTC()); err != nil || dispatched != nil {
		t.Fatalf("task was redispatched before its retry time: job=%#v err=%v", dispatched, err)
	}
}

func TestWorkshopProcessWorkerPublishesPerBookProgress(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	const owner = "tester"
	const bookID = "12345678901"
	if err := store.PutDocument(context.Background(), owner, novelfetchworkshop.Document{BookID: bookID, Meta: map[string]any{"bookName": "日志测试书"}}); err != nil {
		t.Fatal(err)
	}
	registry := newWorkshopProcessRegistry()
	worker := &workshopProcessWorker{store: store, queue: queue, registry: registry}
	worker.process(context.Background(), novelfetchworkshop.ProcessJob{ID: "job-progress", Owner: owner, BatchID: "batch-progress", BookIDs: []string{bookID}})

	job, ok := registry.get(owner, "job-progress")
	if !ok {
		t.Fatal("process job was not published")
	}
	progress, _ := job["progress"].([]any)
	if len(progress) < 3 {
		t.Fatalf("expected start, per-book and completion progress, got %#v", job["progress"])
	}
	joined, _ := json.Marshal(progress)
	for _, want := range []string{"任务开始", bookID, "日志测试书", "处理完成"} {
		if !bytes.Contains(joined, []byte(want)) {
			t.Fatalf("progress missing %q: %s", want, joined)
		}
	}
}

func TestWorkshopProcessWorkerFinalOriginalFetchFailureWaitsForAutoRetry(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	const bookID = "12345678901"
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: bookID}); err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "final failure", http.StatusBadGateway)
	}))
	defer upstream.Close()
	registry := newWorkshopProcessRegistry()
	worker := &workshopProcessWorker{store: store, queue: queue, registry: registry, fetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}}
	worker.process(context.Background(), novelfetchworkshop.ProcessJob{ID: "job-final-failure", Owner: "tester", BatchID: "batch-final-failure", BookIDs: []string{bookID}, Attempts: 2})
	job, _ := registry.get("tester", "job-final-failure")
	progress, _ := json.Marshal(job["progress"])
	if !bytes.Contains(progress, []byte("等待下次自动抓取")) || bytes.Contains(progress, []byte("任务失败")) {
		t.Fatalf("original fetch retry progress is misleading: %s", progress)
	}
}

func TestWorkshopProcessWorkerRetrySkipsCompletedFetchAndAIVersions(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	document := novelfetchworkshop.Document{
		BookID: "33333333333", Gender: "男频", Original: "已经抓取的正文", ProcessedText: "已经抓取的正文",
		Versions: map[string]any{"ai1": "已经生成的第一版"},
		Meta:     map[string]any{"platformId": "2", "originalStatus": "done", "aiStatus": "failed"},
	}
	if err := store.PutDocument(context.Background(), "tester", document); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{BookID: document.BookID, BodyRef: novelfetchworkshop.BodyRef{VersionID: "processed_text"}, Content: document.ProcessedText}); err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "fetch must not run", http.StatusInternalServerError)
	}))
	defer upstream.Close()
	rewriteCalls := 0
	classificationCalls := 0
	worker := &workshopProcessWorker{
		store: store, queue: novelfetchworkshop.NewMemoryQueue(), registry: newWorkshopProcessRegistry(),
		fetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL},
		resolveRewriteProvider: func(context.Context, string) (novelfetchworkshop.RewriteProvider, error) {
			return rewriteProviderFunc(func(_ context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
				if strings.Contains(request.SystemPrompt, "小说分类器") {
					classificationCalls++
					return `{"style":"现代悬疑","confidence":0.9,"reason":"正文围绕案件线索推进"}`, nil
				}
				rewriteCalls++
				return `{"id":"33333333333","changed_lines":[{"line_no":1,"text":"新生成的第二版"}]}`, nil
			}), nil
		},
	}
	err := worker.execute(context.Background(), novelfetchworkshop.ProcessJob{Owner: "tester", BookIDs: []string{document.BookID}, SelectedVersions: []string{"ai1", "ai2"}})
	if err != nil {
		t.Fatal(err)
	}
	updated, _ := store.GetDocument(context.Background(), "tester", document.BookID)
	if rewriteCalls != 1 || classificationCalls != 1 || updated.Versions["ai1"] != "已经生成的第一版" || updated.Versions["ai2"] == nil {
		t.Fatalf("retry regenerated completed stages: rewrite=%d classification=%d versions=%#v", rewriteCalls, classificationCalls, updated.Versions)
	}
}

func TestDispatchDueNovelFetchSchedulesEnqueuesOnce(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	now := time.Date(2026, 10, 17, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"novel_fetch_schedules": []any{map[string]any{
			"id": "schedule-1", "status": "scheduled", "run_at": now.Add(-time.Minute).Format(time.RFC3339),
			"input_text": "12345678901\t定时任务", "platform_id": "2", "max_txt": 4000, "auto_submit": true,
		}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(context.Background(), store, queue, "tester", now); err != nil {
		t.Fatal(err)
	}
	job, err := queue.Claim(context.Background(), 50*time.Millisecond)
	if err != nil {
		t.Fatalf("due schedule was not enqueued: %v", err)
	}
	if job.Owner != "tester" || len(job.BookIDs) != 1 || job.BookIDs[0] != "12345678901" {
		t.Fatalf("unexpected scheduled job: %#v", job)
	}
	if !job.AutoSubmit {
		t.Fatalf("scheduled job lost automatic upload intent: %#v", job)
	}
	if err := queue.Ack(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(context.Background(), store, queue, "tester", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err := queue.Claim(context.Background(), 20*time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
		t.Fatalf("schedule was enqueued more than once: %v", err)
	}
}

func TestDispatchDueNovelFetchSchedulesResumesExistingUnfinishedBooks(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	now := time.Date(2026, 10, 17, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	if err := store.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: "12345678901", OriginalRaw: "已保存原文", Meta: map[string]any{"platformId": "2", "originalStatus": "done"}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{map[string]any{
		"id": "schedule-duplicate", "status": "scheduled", "run_at": now.Add(-time.Minute).Format(time.RFC3339), "input_text": "12345678901\t重复定时书目", "platform_id": "2",
	}}}); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now); err != nil {
		t.Fatal(err)
	}
	job, err := queue.Claim(ctx, 20*time.Millisecond)
	if err != nil || len(job.BookIDs) != 1 || job.BookIDs[0] != "12345678901" {
		t.Fatalf("unfinished book not resumed: %#v %v", job, err)
	}
	settings, err := store.GetConfig(ctx, "tester")
	if err != nil {
		t.Fatal(err)
	}
	schedule := settings["novel_fetch_schedules"].([]any)[0].(map[string]any)
	if schedule["status"] != "queued" {
		t.Fatalf("unfinished schedule falsely completed: %#v", schedule)
	}
	doc, _ := store.GetDocument(ctx, "tester", "12345678901")
	if doc.OriginalRaw != "已保存原文" {
		t.Fatal("existing original overwritten")
	}
}

func TestConcurrentNovelFetchScheduleDispatchQueuesOnlyOneJob(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	now := time.Date(2026, 10, 17, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{map[string]any{
		"id": "schedule-race", "status": "scheduled", "run_at": now.Add(-time.Minute).Format(time.RFC3339),
		"input_text": "12345678901\t并发定时任务", "platform_id": "2", "max_txt": 4000,
	}}}); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			errs <- dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now)
		}()
	}
	close(start)
	for range 2 {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	if _, err := queue.Claim(ctx, 50*time.Millisecond); err != nil {
		t.Fatalf("scheduled task was not queued: %v", err)
	}
	if _, err := queue.Claim(ctx, 20*time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
		t.Fatalf("concurrent dispatch created a second job: %v", err)
	}
}

func TestRecoverIncompleteNovelFetchScheduleDoesNotBlindlyRequeue(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	now := time.Date(2026, 10, 17, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{map[string]any{
		"id": "schedule-incomplete", "status": "dispatching", "updatedAt": now.Add(-10 * time.Minute).Format(time.RFC3339Nano),
		"run_at": now.Add(-time.Hour).Format(time.RFC3339), "input_text": "12345678901\t信息缺失任务", "platform_id": "2",
	}}}); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now); err != nil {
		t.Fatal(err)
	}
	if _, err := queue.Claim(ctx, 20*time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
		t.Fatalf("incomplete dispatch was blindly requeued: %v", err)
	}
	settings, err := store.GetConfig(ctx, "tester")
	if err != nil {
		t.Fatal(err)
	}
	schedule := settings["novel_fetch_schedules"].([]any)[0].(map[string]any)
	if schedule["status"] != "failed" || !strings.Contains(firstText(schedule["error"]), "记录不完整") {
		t.Fatalf("incomplete schedule did not retain a clear failure: %#v", schedule)
	}
}

func TestRecoverNovelFetchScheduleWithUncertainSubmissionWaitsForVerification(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	now := time.Date(2026, 10, 17, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	jobID, batchID := novelFetchScheduleDispatchIDs("tester", "schedule-verify", now.Add(-time.Hour))
	if err := store.PutRun(ctx, "tester", novelfetchworkshop.RunRecord{RunID: jobID, BookID: batchID, Stage: "novel_fetch_process", Status: "upload_pending", TextModelID: "未指定", ModelID: "novel-fetch-workshop", ErrorMessage: "提交回执待确认"}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{map[string]any{
		"id": "schedule-verify", "status": "dispatching", "updatedAt": now.Add(-10 * time.Minute).Format(time.RFC3339Nano),
		"dispatch_id": "dispatch-verify", "job_id": jobID, "batch_id": batchID, "book_ids": []any{"12345678901"},
		"run_at": now.Add(-time.Hour).Format(time.RFC3339), "input_text": "12345678901\t待核验", "platform_id": "2",
	}}}); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now); err != nil {
		t.Fatal(err)
	}
	if _, err := queue.Claim(ctx, 20*time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
		t.Fatalf("uncertain external submission was replayed: %v", err)
	}
	settings, err := store.GetConfig(ctx, "tester")
	if err != nil {
		t.Fatal(err)
	}
	schedule := settings["novel_fetch_schedules"].([]any)[0].(map[string]any)
	if schedule["status"] != "pending_verification" || !strings.Contains(firstText(schedule["error"]), "回执") {
		t.Fatalf("uncertain submission was not held for verification: %#v", schedule)
	}
}

func TestNovelFetchScheduleDefaultsToAutomaticUpload(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	created := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/schedules", map[string]any{
		"run_at": "2027-01-01T12:00", "input_text": "12345678901 定时上传测试", "selected_versions": []string{"ai1"}, "text_model_id": "text-model-1",
		"input_format": "paid_id_title_gender_reason", "column_preset": "付费 ID + 书名 + 男女频 + 理由", "column_order": []string{"bookId", "bookName", "gender", "reason"},
	}, now, secret)
	if created.Code != http.StatusOK {
		t.Fatalf("create schedule status=%d body=%s", created.Code, created.Body.String())
	}
	var response struct {
		Schedule map[string]any `json:"schedule"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Schedule["auto_submit"] != true {
		t.Fatalf("new schedule must default to automatic upload: %#v", response.Schedule)
	}
	if response.Schedule["text_model_id"] != "text-model-1" || response.Schedule["input_format"] != "paid_id_title_gender_reason" {
		t.Fatalf("new schedule lost processing metadata: %#v", response.Schedule)
	}
}

func TestWorkshopProcessWorkerCompletesScheduledTaskStatus(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"novel_fetch_schedules": []any{map[string]any{"id": "schedule-done", "status": "queued"}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "book-done"}); err != nil {
		t.Fatal(err)
	}
	worker := &workshopProcessWorker{store: store, queue: novelfetchworkshop.NewMemoryQueue(), registry: newWorkshopProcessRegistry()}
	worker.process(context.Background(), novelfetchworkshop.ProcessJob{
		ID: "job-done", Owner: "tester", BatchID: "batch-done", BookIDs: []string{"book-done"}, ScheduleID: "schedule-done",
	})
	settings, err := store.GetConfig(context.Background(), "tester")
	if err != nil {
		t.Fatal(err)
	}
	schedules := settings["novel_fetch_schedules"].([]any)
	if schedules[0].(map[string]any)["status"] != "completed" {
		t.Fatalf("scheduled task status was not completed: %#v", schedules)
	}
}

func TestWorkshopProcessWorkerAutomaticallyUploadsCompletedScheduledTask(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"novel_fetch_schedules": []any{map[string]any{"id": "schedule-upload", "status": "queued", "auto_submit": true}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "book-upload"}); err != nil {
		t.Fatal(err)
	}
	called := 0
	worker := &workshopProcessWorker{
		store: store, queue: novelfetchworkshop.NewMemoryQueue(), registry: newWorkshopProcessRegistry(),
		autoSubmit: func(_ context.Context, job novelfetchworkshop.ProcessJob) error {
			called++
			if job.ScheduleID != "schedule-upload" || len(job.BookIDs) != 1 || job.BookIDs[0] != "book-upload" {
				t.Fatalf("unexpected automatic upload job: %#v", job)
			}
			return nil
		},
	}
	worker.process(context.Background(), novelfetchworkshop.ProcessJob{
		ID: "job-upload", Owner: "tester", BatchID: "batch-upload", BookIDs: []string{"book-upload"}, ScheduleID: "schedule-upload", AutoSubmit: true,
	})
	if called != 1 {
		t.Fatalf("automatic upload calls=%d, want 1", called)
	}
}

func TestNovelFetchWorkshopProcessPreviewDoesNotPersistTasks(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	preview := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/preview", map[string]any{
		"input_text": "书籍ID\t书名\t男女频\n12345678901\t预览一本\t男频\n12345678901\t预览一本\t男频\n22345678901\t预览二本\t女频",
		"max_txt":    4000,
	}, now, secret)
	if preview.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", preview.Code, preview.Body.String())
	}
	var response struct {
		Parsed         int              `json:"parsed"`
		UniqueTasks    int              `json:"unique_tasks"`
		DuplicateCount int              `json:"duplicate_count"`
		Tasks          []map[string]any `json:"tasks"`
	}
	if err := json.Unmarshal(preview.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Parsed != 3 || response.UniqueTasks != 2 || response.DuplicateCount != 1 || len(response.Tasks) != 2 {
		t.Fatalf("unexpected preview response: %#v", response)
	}
	documents, err := store.ListDocuments(context.Background(), "tester")
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) != 0 {
		t.Fatalf("preview must not persist tasks: %d", len(documents))
	}
}

func TestNovelFetchWorkshopProcessPreviewParsesIDAndBookNameSeparatedByWhitespace(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})
	preview := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/preview", map[string]any{
		"input_text": "2094511573566019532    鼓励式老公，把行动都给了闺蜜\n2094099450415395858\u00a0\u00a0寒灯验骨尘缘尽",
		"max_txt":    4000,
	}, now, secret)
	if preview.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", preview.Code, preview.Body.String())
	}
	var response struct {
		Tasks []map[string]any `json:"tasks"`
	}
	if err := json.Unmarshal(preview.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Tasks) != 2 {
		t.Fatalf("unexpected tasks: %#v", response.Tasks)
	}
	if got := response.Tasks[0]["bookName"]; got != "鼓励式老公，把行动都给了闺蜜" {
		t.Fatalf("ordinary-space book name=%q", got)
	}
	if got := response.Tasks[1]["bookName"]; got != "寒灯验骨尘缘尽" {
		t.Fatalf("non-breaking-space book name=%q", got)
	}
}

func TestWorkshopPreviewKeepsInputFormatMetadata(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})
	preview := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/preview", map[string]any{
		"input_text":   "2094511573566019532\t鼓励式老公\t女频\t闺蜜背刺",
		"input_format": "paid_id_title_gender_reason", "column_preset": "付费 ID + 书名 + 男女频 + 理由", "column_order": []string{"bookId", "bookName", "gender", "reason"},
	}, now, secret)
	if preview.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", preview.Code, preview.Body.String())
	}
	var response map[string]any
	if err := json.Unmarshal(preview.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response["input_format"] != "paid_id_title_gender_reason" || response["column_preset"] != "付费 ID + 书名 + 男女频 + 理由" {
		t.Fatalf("preview lost input metadata: %#v", response)
	}
	columns, ok := response["column_order"].([]any)
	if !ok || len(columns) != 4 || columns[0] != "bookId" {
		t.Fatalf("preview column order=%#v", response["column_order"])
	}
}

func TestWorkshopStartPersistsInputFormatMetadataOnTasks(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: novelfetchworkshop.NewMemoryQueue()})
	started := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{
		"inputText":    "2094511573566019532\t鼓励式老公\t女频\t闺蜜背刺",
		"input_format": "paid_id_title_gender_reason", "column_preset": "付费 ID + 书名 + 男女频 + 理由", "column_order": []string{"bookId", "bookName", "gender", "reason"},
	}, now, secret)
	if started.Code != http.StatusAccepted {
		t.Fatalf("start status=%d body=%s", started.Code, started.Body.String())
	}
	document, err := store.GetDocument(context.Background(), "tester", "2094511573566019532")
	if err != nil {
		t.Fatal(err)
	}
	if document.Meta["inputFormat"] != "paid_id_title_gender_reason" || document.Meta["columnPreset"] != "付费 ID + 书名 + 男女频 + 理由" {
		t.Fatalf("task lost input metadata: %#v", document.Meta)
	}
}

func TestWorkshopStartAlwaysAutoSubmitsToVideoSystem(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: novelfetchworkshop.NewMemoryQueue()})
	started := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{
		"inputText": "12345678901\t直接提交测试",
	}, now, secret)
	if started.Code != http.StatusAccepted {
		t.Fatalf("start status=%d body=%s", started.Code, started.Body.String())
	}
	document, err := store.GetDocument(context.Background(), "tester", "12345678901")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, ok := document.Meta["processSnapshot"].(map[string]any)
	if !ok || snapshot["auto_submit"] != true {
		t.Fatalf("manual start should auto submit, snapshot=%#v", document.Meta["processSnapshot"])
	}
}

func TestDeriveNovelFetchBatchStatusUsesTaskStates(t *testing.T) {
	completed := novelfetchworkshop.Document{Meta: map[string]any{"status": "completed"}}
	queued := novelfetchworkshop.Document{Meta: map[string]any{"status": "queued"}}
	if got := deriveNovelFetchBatchStatus([]novelfetchworkshop.Document{completed}); got != "done" {
		t.Fatalf("completed batch status=%q", got)
	}
	if got := deriveNovelFetchBatchStatus([]novelfetchworkshop.Document{completed, queued}); got != "queued" {
		t.Fatalf("mixed batch status=%q", got)
	}
}

func TestNovelFetchWorkshopProcessStartCanBePolled(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue})
	start := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"inputText": "12345678901\t可轮询任务"}, now, secret)
	if start.Code != http.StatusAccepted {
		t.Fatalf("start status=%d body=%s", start.Code, start.Body.String())
	}
	var job struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(start.Body.Bytes(), &job); err != nil {
		t.Fatal(err)
	}
	if job.ID == "" || job.Status != "queued" {
		t.Fatalf("unexpected job=%+v", job)
	}
	polled := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/process/jobs/"+job.ID, nil, now, secret)
	if polled.Code != http.StatusOK || !bytes.Contains(polled.Body.Bytes(), []byte(job.ID)) {
		t.Fatalf("poll status=%d body=%s", polled.Code, polled.Body.String())
	}
	latest := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/process/jobs/latest?limit=5", nil, now, secret)
	if latest.Code != http.StatusOK || !bytes.Contains(latest.Body.Bytes(), []byte(job.ID)) {
		t.Fatalf("latest status=%d body=%s", latest.Code, latest.Body.String())
	}
	runs, err := store.ListRuns(context.Background(), "tester", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || runs[0].BookID == "" {
		t.Fatalf("process run audit was not persisted with a batch identifier: %#v", runs)
	}
}

func TestNovelFetchWorkshopProcessStartReturnsBeforeBackgroundFetchCompletes(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	started := make(chan struct{})
	release := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		_, _ = w.Write([]byte(`{"code":200,"data":"完整小说正文","bookinfo":{"category":"男频"}}`))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{
		BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store,
		NovelFetchQueue:   novelfetchworkshop.NewMemoryQueue(),
		NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL},
	})
	response := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		response <- doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"inputText": "12345678901\t异步任务"}, now, secret)
	}()
	var start *httptest.ResponseRecorder
	select {
	case start = <-response:
		if start.Code != http.StatusAccepted {
			t.Fatalf("start status=%d body=%s", start.Code, start.Body.String())
		}
	case <-time.After(200 * time.Millisecond):
		close(release)
		t.Fatal("start request waited for the upstream fetch")
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		close(release)
		t.Fatal("background worker did not start fetch")
	}
	close(release)
	var job struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(start.Body.Bytes(), &job); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		polled := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/process/jobs/"+job.ID, nil, now, secret)
		if bytes.Contains(polled.Body.Bytes(), []byte(`"status":"done"`)) {
			document, err := store.GetDocument(context.Background(), "tester", "12345678901")
			if err != nil {
				t.Fatal(err)
			}
			for _, event := range []string{"fetch_started", "original_raw_saved", "sensitive_processing_completed", "classification_completed"} {
				if !novelFetchDocumentHasLogEvent(document, event) {
					t.Fatalf("persisted task logs missing %q: %#v", event, document.Logs)
				}
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("background fetch did not finish")
}

func TestNovelFetchPreviewHonorsLegacyParseModeAndColumnOrder(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/preview", map[string]any{
		"input_text":   "付费书名\t12345678901\t忽略列\t女频\t推荐理由",
		"parse_mode":   "custom",
		"column_order": "书名,书籍ID,忽略,男女频,推荐理由",
		"max_txt":      4000,
	}, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Tasks []map[string]any `json:"tasks"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || len(body.Tasks) != 1 {
		t.Fatalf("preview decode=%v body=%s", err, response.Body.String())
	}
	if firstText(body.Tasks[0]["bookId"], body.Tasks[0]["book_id"]) != "12345678901" || firstText(body.Tasks[0]["bookName"]) != "付费书名" || firstText(body.Tasks[0]["gender"]) != "女频" {
		t.Fatalf("legacy column order ignored: %#v", body.Tasks[0])
	}
}

func TestNovelFetchPreviewHonorsFixedPresetAndHeader(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})
	cases := []map[string]any{
		{"input_text": "忽略\t免费ID\t12345678901\t固定书名\t9.5\tx\ty\tz\tw\t男频\t固定理由", "parse_mode": "fixed_full_11"},
		{"input_text": "书籍ID\t书名\t男女频\t推荐理由\n12345678901\t表头书名\t女频\t表头理由", "parse_mode": "header"},
	}
	expected := []struct{ name, gender string }{{"固定书名", "男频"}, {"表头书名", "女频"}}
	for index, payload := range cases {
		payload["max_txt"] = 4000
		response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/preview", payload, now, secret)
		var body struct {
			Tasks []map[string]any `json:"tasks"`
		}
		if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &body) != nil || len(body.Tasks) != 1 {
			t.Fatalf("case %d preview failed: status=%d body=%s", index, response.Code, response.Body.String())
		}
		if firstText(body.Tasks[0]["bookName"]) != expected[index].name || firstText(body.Tasks[0]["gender"]) != expected[index].gender {
			t.Fatalf("case %d wrong parse: %#v", index, body.Tasks[0])
		}
	}
}

func TestNovelFetchTaskDetailPayloadUsesDerivedStatus(t *testing.T) {
	document := novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{
		"status": "completed", "queueState": "running", "aiStatus": "failed",
	}}
	payload := novelFetchTaskDetailPayload(document)
	if payload["status"] != novelfetchworkshop.TaskStatusProcessing || payload["effective_status"] != novelfetchworkshop.TaskStatusProcessing {
		t.Fatalf("detail payload status=%v effective=%v", payload["status"], payload["effective_status"])
	}
	meta := mapOrEmpty(payload["meta"])
	if meta["status"] != novelfetchworkshop.TaskStatusProcessing || meta["effectiveStatus"] != novelfetchworkshop.TaskStatusProcessing {
		t.Fatalf("detail meta status=%v effective=%v", meta["status"], meta["effectiveStatus"])
	}
}

func TestNovelFetchWorkshopProcessStartReturnsExistingTaskForActiveDuplicateBook(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	queue := &uniqueRecordingNovelFetchQueue{}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue})
	first := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"inputText": "12345678901\t重复书目"}, now, secret)
	if first.Code != http.StatusAccepted {
		t.Fatalf("first start status=%d body=%s", first.Code, first.Body.String())
	}
	var firstJob map[string]any
	if err := json.Unmarshal(first.Body.Bytes(), &firstJob); err != nil {
		t.Fatal(err)
	}
	second := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"inputText": "12345678901\t重复书目"}, now, secret)
	if second.Code != http.StatusAccepted {
		t.Fatalf("duplicate start status=%d body=%s", second.Code, second.Body.String())
	}
	var secondJob map[string]any
	if err := json.Unmarshal(second.Body.Bytes(), &secondJob); err != nil {
		t.Fatal(err)
	}
	if firstJob["id"] != secondJob["id"] || secondJob["existing"] != true || len(queue.jobs) != 1 {
		t.Fatalf("duplicate request did not return the active job: first=%#v second=%#v jobs=%#v", firstJob, secondJob, queue.jobs)
	}
}

func novelFetchDocumentHasLogEvent(document novelfetchworkshop.Document, want string) bool {
	for _, raw := range document.Logs {
		log, ok := raw.(map[string]any)
		if ok && firstText(log["event"]) == want {
			return true
		}
	}
	return false
}

func TestNovelFetchWorkshopFetchOriginalPersistsRawAndProcessedText(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("bookid") != "12345678901" || req.URL.Query().Get("platform") != "15" {
			t.Fatalf("unexpected upstream query: %s", req.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","data":"第一行\n第二行\n第三行","bookinfo":{"category":"男频","genre":"都市","work_title":"测试小说"}}`))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "12345678901", MaxTxt: 4, Meta: map[string]any{"platformId": "15"}}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}})
	fetch := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/12345678901/fetch", map[string]any{"maxTxt": 4}, now, secret)
	if fetch.Code != http.StatusOK {
		t.Fatalf("fetch status=%d body=%s", fetch.Code, fetch.Body.String())
	}
	document, err := store.GetDocument(context.Background(), "tester", "12345678901")
	if err != nil {
		t.Fatal(err)
	}
	if document.OriginalRaw != "第一行\n第二行\n第三行" || document.ProcessedText != "第一行" || document.ProcessedChars != 3 || document.OriginalRawChars <= document.ProcessedChars || document.Gender != "男频" || document.Meta["sensitiveStatus"] != "done" {
		t.Fatalf("fetch persistence mismatch: %#v", document)
	}
	if document.BookInfo.WorkTitle != "测试小说" || document.Meta["bookName"] != "测试小说" {
		t.Fatalf("fetched title was not synchronized: %#v", document)
	}
}

func TestNovelFetchWorkshopRestoreOriginalReprocessesStoredRawText(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "12345678901", OriginalRaw: "完整原文内容", ProcessedText: "旧版本", Original: "旧版本", MaxTxt: 4}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	restore := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/12345678901/restore-original", nil, now, secret)
	if restore.Code != http.StatusOK {
		t.Fatalf("restore status=%d body=%s", restore.Code, restore.Body.String())
	}
	document, err := store.GetDocument(context.Background(), "tester", "12345678901")
	if err != nil {
		t.Fatal(err)
	}
	if document.OriginalRaw != "完整原文内容" || document.ProcessedText != "完整原文" || document.Meta["originalStatus"] != "restored" {
		t.Fatalf("restore mismatch: %#v", document)
	}
}

func TestNovelFetchBatchRetryReusesStoredRawInsteadOfFetchingAgain(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	requests := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"code":200,"data":"不应重新抓取"}`))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	document := novelfetchworkshop.Document{BookID: "stored-raw", OriginalRaw: "已保存的完整原文", Gender: "女频", MaxTxt: 4000, Meta: map[string]any{
		"style": "现代悬疑", "platformId": "15", "originalStatus": "waiting_sensitive_ai", "sensitiveStatus": "waiting_retry",
		"status": "waiting_retry", "queueState": "waiting_retry", "error": "旧的敏感词处理失败",
	}}
	if err := store.PutDocument(context.Background(), "tester", document); err != nil {
		t.Fatal(err)
	}
	settings, _ := store.GetConfig(context.Background(), "tester")
	mapOrEmpty(settings["workflow"])["skip_short_original"] = false
	_ = store.PutConfig(context.Background(), "tester", settings)
	queue := &uniqueRecordingNovelFetchQueue{}
	fetcher := &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue, NovelFetchFetcher: fetcher})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/batch-retry", map[string]any{"ids": []string{"stored-raw"}}, now, secret)
	if response.Code != http.StatusAccepted {
		t.Fatalf("batch retry status=%d body=%s", response.Code, response.Body.String())
	}
	if len(queue.jobs) != 1 {
		t.Fatal("retry was not queued")
	}
	job := queue.jobs[0]
	// Isolate the stored-raw recovery stage; missing AI versions are covered by the retry queue tests.
	job.SelectedVersions = nil
	worker := &workshopProcessWorker{store: store, queue: queue, registry: newWorkshopProcessRegistry(), fetcher: fetcher}
	worker.process(context.Background(), job)
	updated, err := store.GetDocument(context.Background(), "tester", "stored-raw")
	if err != nil {
		t.Fatal(err)
	}
	if requests != 0 || updated.OriginalRaw != "已保存的完整原文" || updated.Meta["originalStatus"] != "done" {
		t.Fatalf("retry did not resume from stored raw: requests=%d doc=%#v", requests, updated)
	}
	if updated.Meta["status"] != "completed" || updated.Meta["queueState"] != "completed" || updated.Meta["error"] != nil {
		t.Fatalf("successful retry kept a stale failure state: %#v", updated.Meta)
	}
	if len(updated.Logs) == 0 {
		t.Fatal("successful retry did not append a recovery log")
	}
	lastLog, _ := updated.Logs[len(updated.Logs)-1].(map[string]any)
	if lastLog["event"] != "process_completed" {
		t.Fatalf("successful retry left the task detail on a stale failure log: %#v", updated.Logs)
	}
}

func TestNovelFetchReimportDoesNotOverwriteExistingDocumentData(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	original := novelfetchworkshop.Document{BookID: "12345678901", OriginalRaw: "完整 raw", ProcessedText: "处理后正文", Original: "处理后正文", Versions: map[string]any{"ai1": "改写版本"}, Meta: map[string]any{"platformId": "15", "originalStatus": "done", "customFlag": "keep"}}
	if err := store.PutDocument(context.Background(), "tester", original); err != nil {
		t.Fatal(err)
	}
	if _, err := workshopProcessResult(context.Background(), store, "tester", "12345678901\t新导入标题", "15", "batch-new", 4000); err != nil {
		t.Fatal(err)
	}
	updated, err := store.GetDocument(context.Background(), "tester", "12345678901")
	if err != nil {
		t.Fatal(err)
	}
	if updated.OriginalRaw != original.OriginalRaw || updated.ProcessedText != original.ProcessedText || updated.Versions["ai1"] != original.Versions["ai1"] || updated.Meta["customFlag"] != "keep" {
		t.Fatalf("reimport overwrote existing document data: %#v", updated)
	}
}

func TestNovelFetchStartReprocessesCompletedBookWithMissingVersions(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	queue := &uniqueRecordingNovelFetchQueue{}
	original := novelfetchworkshop.Document{BookID: "12345678901", OriginalRaw: "已有完整原文", ProcessedText: "已有处理正文", Meta: map[string]any{"platformId": "15", "originalStatus": "done", "status": "completed"}}
	if err := store.PutDocument(context.Background(), "tester", original); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"inputText": "12345678901\t重复导入", "platform_id": "15", "selected_versions": []string{"original", "ai1"}}, now, secret)
	if response.Code != http.StatusAccepted || len(queue.jobs) != 1 {
		t.Fatalf("completed book with missing AI version was not reprocessed: status=%d queued=%d body=%s", response.Code, len(queue.jobs), response.Body.String())
	}
	updated, err := store.GetDocument(context.Background(), "tester", original.BookID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.OriginalRaw != original.OriginalRaw || updated.ProcessedText != original.ProcessedText {
		t.Fatalf("duplicate start changed existing text: %#v", updated)
	}
}

func TestNovelFetchStartSkipsCompletedBookWithSelectedVersionsDone(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	queue := &uniqueRecordingNovelFetchQueue{}
	original := novelfetchworkshop.Document{BookID: "12345678901", OriginalRaw: "已有完整原文", ProcessedText: "已有处理正文", Versions: map[string]any{"ai1": "已有AI正文"}, Meta: map[string]any{"platformId": "15", "originalStatus": "done", "aiStatus": "done", "status": "completed"}}
	if err := store.PutDocument(context.Background(), "tester", original); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"inputText": "12345678901\t重复导入", "platform_id": "15", "selected_versions": []string{"original", "ai1"}}, now, secret)
	if response.Code != http.StatusOK || len(queue.jobs) != 0 {
		t.Fatalf("completed book with all versions was not skipped: status=%d queued=%d body=%s", response.Code, len(queue.jobs), response.Body.String())
	}
}

func TestNovelFetchStartQueuesRemainingBooksWhenOneBookIsBusy(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	queue := &uniqueRecordingNovelFetchQueue{bookDedupe: map[string]string{"tester:15:12345678901": "existing-job"}}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{
		"inputText":         "12345678901\t书一\n12345678902\t书二",
		"platform_id":       "15",
		"selected_versions": []string{"ai1"},
	}, now, secret)
	if response.Code != http.StatusAccepted {
		t.Fatalf("unexpected status=%d body=%s", response.Code, response.Body.String())
	}
	if len(queue.jobs) != 1 {
		t.Fatalf("queued jobs=%d, want 1", len(queue.jobs))
	}
	if len(queue.jobs[0].BookIDs) != 1 || queue.jobs[0].BookIDs[0] != "12345678902" {
		t.Fatalf("queued bookIDs=%v, want only 12345678902", queue.jobs[0].BookIDs)
	}
	busy, err := store.GetDocument(context.Background(), "tester", "12345678901")
	if err != nil {
		t.Fatal(err)
	}
	if got := firstText(busy.Meta["status"], busy.Meta["queueState"]); got != "waiting_retry" {
		t.Fatalf("busy book status=%q, want waiting_retry", got)
	}
	queued, err := store.GetDocument(context.Background(), "tester", "12345678902")
	if err != nil {
		t.Fatal(err)
	}
	if got := firstText(queued.Meta["status"], queued.Meta["queueState"]); got != "queued" {
		t.Fatalf("free book status=%q, want queued", got)
	}
}

func TestNovelFetchStartCollapsesZeroPaddedAIVersions(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	queue := &uniqueRecordingNovelFetchQueue{}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: queue})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{
		"inputText":         "12345678901\t书一",
		"platform_id":       "15",
		"selected_versions": []string{"ai01", "ai1", "ai001"},
	}, now, secret)
	if response.Code != http.StatusAccepted || len(queue.jobs) != 1 {
		t.Fatalf("static_versions status=%d jobs=%d body=%s", response.Code, len(queue.jobs), response.Body.String())
	}
	if got := queue.jobs[0].SelectedVersions; len(got) != 1 || got[0] != "ai1" {
		t.Fatalf("selected versions=%v, want [ai1]", got)
	}

	invalid := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{
		"inputText":         "12345678903\t书三",
		"platform_id":       "15",
		"selected_versions": []string{"ai6"},
	}, now, secret)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("ai6 status=%d, want 400 body=%s", invalid.Code, invalid.Body.String())
	}
}

func TestNovelFetchProcessRejectsOversizedMaxTxt(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process", map[string]any{
		"inputText": "12345678901\t书一",
		"maxTxt":    novelFetchMaxTxtLimit + 1,
	}, now, secret)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("oversized maxTxt status=%d, want 400 body=%s", response.Code, response.Body.String())
	}
}

func TestNovelFetchDeleteSkipsRunningBookWithStaleCompletedStatus(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	ctx := context.Background()
	if err := store.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{
		"status": "completed", "queueState": "processing", "currentProcessJobId": "job-running",
	}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: "12345678902", Meta: map[string]any{
		"status": "completed",
	}}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodDelete, "/api/novel-fetch-workshop/tasks", map[string]any{
		"ids": []string{"12345678901", "12345678902"},
	}, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Deleted int              `json:"deleted"`
		Skipped []map[string]any `json:"skipped"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Deleted != 1 || len(payload.Skipped) != 1 {
		t.Fatalf("deleted=%d skipped=%d, want 1/1 body=%s", payload.Deleted, len(payload.Skipped), response.Body.String())
	}
	if _, err := store.GetDocument(ctx, "tester", "12345678901"); err != nil {
		t.Fatalf("running book was deleted: %v", err)
	}
	if _, err := store.GetDocument(ctx, "tester", "12345678902"); err == nil {
		t.Fatal("idle book was not deleted")
	}
}

func TestNovelFetchMemoryStoreIsOwnerScoped(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "alice", novelfetchworkshop.Document{BookID: "same", Original: "alice"}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "bob", novelfetchworkshop.Document{BookID: "same", Original: "bob"}); err != nil {
		t.Fatal(err)
	}
	a, _ := store.GetDocument(context.Background(), "alice", "same")
	b, _ := store.GetDocument(context.Background(), "bob", "same")
	if a.Original != "alice" || b.Original != "bob" {
		t.Fatalf("owner isolation broken: alice=%q bob=%q", a.Original, b.Original)
	}
}

func TestLegacyNovelFetchTaskUsesFetchedBookTitleWhenInputNameIsMissing(t *testing.T) {
	task := legacyNovelFetchTask(novelfetchworkshop.Document{
		BookID: "2094511573566019532",
		BookInfo: novelfetchworkshop.BookInfo{
			WorkTitle: "鼓励式老公，却把行动都给了闺蜜",
		},
		Meta: map[string]any{},
	})

	if got := task["book_name"]; got != "鼓励式老公，却把行动都给了闺蜜" {
		t.Fatalf("book_name = %q, want fetched work title", got)
	}
}

func TestNovelFetchRealtimeStatusPrefersRuntimeQueueState(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "2094511573566019532",
		Meta: map[string]any{
			"status":   "queued",
			"aiStatus": "ai_done",
		},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/realtime/status", nil, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Counts    map[string]int `json:"counts"`
		UpdatedAt string         `json:"updated_at"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Counts["queued"] != 1 || payload.UpdatedAt == "" {
		t.Fatalf("realtime status = %#v", payload)
	}
}

func TestNovelFetchManualFetchUsesConfiguredEndpointAndImmediateRetry(t *testing.T) {
	requests := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Query().Get("bookid") != "12345" || r.URL.Query().Get("platform") != "15" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		if requests == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte(`{"code":200,"data":"重试成功的原文"}`))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"fetch": map[string]any{
		"endpoint": upstream.URL, "timeout_seconds": 5, "retries": 1,
	}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "12345", MaxTxt: 4000, Meta: map[string]any{"platformId": "15"}}); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Endpoint: "http://invalid.local"}})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/12345/fetch", nil, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if requests != 2 {
		t.Fatalf("upstream requests=%d, want configured immediate retry", requests)
	}
	document, err := store.GetDocument(context.Background(), "tester", "12345")
	if err != nil {
		t.Fatal(err)
	}
	if document.OriginalRaw != "重试成功的原文" {
		t.Fatalf("original raw=%q", document.OriginalRaw)
	}
}
