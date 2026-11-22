package httpapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

// Redis 不可用时不能每分钟都改写同一批书；退避到期后才重新参与调度。
func TestEnqueueFailureBacksOffBeforeRetrying(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(ctx, "alice", novelfetchworkshop.Document{
		BookID: "book-1",
		Meta:   map[string]any{"status": "input_ready"},
	}); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 14, 3, 0, 0, 0, time.UTC)
	enqueueErr := errors.New("redis unavailable")

	markNovelFetchAutomationEnqueueRetry(ctx, store, "alice", []string{"book-1"}, now, enqueueErr)

	document, err := store.GetDocument(ctx, "alice", "book-1")
	if err != nil {
		t.Fatal(err)
	}
	if got := firstText(document.Meta["status"]); got != "waiting_retry" {
		t.Fatalf("status=%q, want waiting_retry", got)
	}
	if got := intValue(document.Meta["automationEnqueueAttempts"]); got != 1 {
		t.Fatalf("attempts=%d, want 1", got)
	}
	nextRaw := firstText(document.Meta["automationNextRetryAt"])
	next, parseErr := time.Parse(time.RFC3339Nano, nextRaw)
	if parseErr != nil {
		t.Fatalf("next retry at %q: %v", nextRaw, parseErr)
	}
	if !next.After(now) {
		t.Fatalf("next retry at %s must be in the future (now=%s)", next, now)
	}
	if novelFetchAutomationEligible(document, now.Add(time.Second)) {
		t.Fatal("document must not be eligible before the backoff deadline")
	}
	if !novelFetchAutomationEligible(document, next.Add(time.Second)) {
		t.Fatal("document must return to the queue after the backoff deadline")
	}

	// 第二次失败继续放大间隔，避免持续失败时形成固定频率的写风暴。
	markNovelFetchAutomationEnqueueRetry(ctx, store, "alice", []string{"book-1"}, next, enqueueErr)
	document, err = store.GetDocument(ctx, "alice", "book-1")
	if err != nil {
		t.Fatal(err)
	}
	if got := intValue(document.Meta["automationEnqueueAttempts"]); got != 2 {
		t.Fatalf("attempts=%d, want 2", got)
	}
}

// 自动重试预算耗尽后必须转人工，不能无限停在“等待重试”。
func TestEnqueueFailureStopsRetryingAtBudget(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(ctx, "alice", novelfetchworkshop.Document{
		BookID: "book-1",
		Meta: map[string]any{
			"status":                    "waiting_retry",
			"automationEnqueueAttempts": novelFetchEnqueueMaxAttempts - 1,
			"automationNextRetryAt":     "2026-11-14T03:00:00Z",
		},
	}); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 14, 3, 0, 0, 0, time.UTC)
	markNovelFetchAutomationEnqueueRetry(ctx, store, "alice", []string{"book-1"}, now, errors.New("redis unavailable"))

	document, err := store.GetDocument(ctx, "alice", "book-1")
	if err != nil {
		t.Fatal(err)
	}
	if got := firstText(document.Meta["status"]); got != "failed" {
		t.Fatalf("status=%q, want failed after retry budget is exhausted", got)
	}
	if _, ok := document.Meta["automationNextRetryAt"]; ok {
		t.Fatal("a failed task must not keep a next retry time")
	}
	if !strings.Contains(firstText(document.Meta["error"]), "已停止自动重试") {
		t.Fatalf("error=%q, want a manual-actionable reason", firstText(document.Meta["error"]))
	}
	if novelFetchAutomationEligible(document, now.Add(24*time.Hour)) {
		t.Fatal("a task that exhausted its retry budget must not be re-queued automatically")
	}
}

// 对照参考实现的 manual_ai_retry_required：模型截断属于确定性失败，必须转人工。
func TestNovelFetchAIManualRetryClassification(t *testing.T) {
	if !novelFetchAIManualRetryRequired(errors.New("2104672163164852230：文本模型输出被长度截断，请重试或增大输出上限")) {
		t.Fatal("a truncated model output must require manual retry")
	}
	if !novelFetchAIManualRetryRequired(errors.New("HTTP 403 · content policy violation")) {
		t.Fatal("a content policy rejection must require manual retry")
	}
	if novelFetchAIManualRetryRequired(errors.New("dial tcp: connection reset")) {
		t.Fatal("a transient network error must stay retryable")
	}
}

func TestMarkNovelFetchAIManualRetryBlocksAutomation(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(ctx, "owner", novelfetchworkshop.Document{BookID: "book-1", Meta: map[string]any{"status": "processing"}}); err != nil {
		t.Fatalf("seed document: %v", err)
	}
	if err := markNovelFetchAIManualRetry(ctx, store, "owner", "book-1", errors.New("文本模型输出被长度截断"), time.Now().UTC()); err != nil {
		t.Fatalf("mark manual retry: %v", err)
	}
	document, err := store.GetDocument(ctx, "owner", "book-1")
	if err != nil {
		t.Fatalf("load document: %v", err)
	}
	if document.Meta["manualOnly"] != true || document.Meta["manualAiRetryRequired"] != true {
		t.Fatalf("manual flags missing: %#v", document.Meta)
	}
	if document.Meta["status"] != "failed" {
		t.Fatalf("status must be failed, got %v", document.Meta["status"])
	}
	if novelFetchAutomationEligible(document, time.Now().UTC()) {
		t.Fatal("a book awaiting manual AI retry must not be picked up by automation")
	}
}

func TestPermanentOriginalFetchFailureStopsImmediately(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	now := time.Date(2026, 11, 23, 3, 0, 0, 0, time.UTC)
	meta := map[string]any{
		"status":                       "processing",
		"originalFetchNextRetryAt":     now.Format(time.RFC3339Nano),
		"original_fetch_next_retry_at": now.Format(time.RFC3339Nano),
		"automationNextRetryAt":        now.Format(time.RFC3339Nano),
		"automation_next_retry_at":     now.Format(time.RFC3339Nano),
	}
	scheduleNovelFetchTaskRetry(meta, novelFetchRetryStageOriginalFetch, errors.New("previous fetch"), time.Minute, now)
	scheduleNovelFetchTaskRetry(meta, novelFetchRetryStagePipeline, errors.New("previous pipeline"), time.Hour, now)
	if err := store.PutDocument(ctx, "owner", novelfetchworkshop.Document{BookID: "book-1", Meta: meta}); err != nil {
		t.Fatal(err)
	}
	fetchErr := fmt.Errorf("%w: %w", errNovelFetchOriginalFetch,
		&novelfetchworkshop.UpstreamError{Class: novelfetchworkshop.UpstreamFailurePermanent, BusinessCode: 400, Message: "获取书籍信息失败"})
	if err := markNovelFetchOriginalFetchRetry(ctx, store, "owner", "book-1", fetchErr, now); err != nil {
		t.Fatal(err)
	}
	document, err := store.GetDocument(ctx, "owner", "book-1")
	if err != nil {
		t.Fatal(err)
	}
	meta = document.Meta
	if meta["status"] != "failed" || meta["queueState"] != "failed" || meta["originalStatus"] != "waiting_manual_fetch" ||
		meta["manualOnly"] != true || meta["manualOriginalFetchRequired"] != true || intValue(meta["originalFetchRetryAttempts"]) != 1 {
		t.Fatalf("meta=%#v, want first failure held for manual fetch", meta)
	}
	if !strings.Contains(firstText(meta["error"]), "原文来源无效") || meta["originalFetchLastError"] != fetchErr.Error() {
		t.Fatalf("source failure reason was lost: %#v", meta)
	}
	for _, key := range []string{"originalFetchNextRetryAt", "original_fetch_next_retry_at", "automationNextRetryAt", "automation_next_retry_at", "next_retry_at", "retry_schedule"} {
		if _, ok := meta[key]; ok {
			t.Fatalf("permanent failure kept %s: %#v", key, meta)
		}
	}
	if novelFetchAutomationEligible(document, now.Add(24*time.Hour)) {
		t.Fatal("permanent failure must not return to automation")
	}
}

func TestConfigurationOriginalFetchFailureWaitsForManualRetry(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	now := time.Date(2026, 11, 23, 3, 0, 0, 0, time.UTC)
	meta := map[string]any{
		"status":                   "processing",
		"siteSubmitStatus":         "pending_upload",
		"originalFetchNextRetryAt": now.Format(time.RFC3339Nano),
		"automationNextRetryAt":    now.Format(time.RFC3339Nano),
	}
	scheduleNovelFetchTaskRetry(meta, novelFetchRetryStageOriginalFetch, errors.New("previous fetch"), time.Minute, now)
	if err := store.PutDocument(ctx, "owner", novelfetchworkshop.Document{BookID: "book-1", Meta: meta}); err != nil {
		t.Fatal(err)
	}
	fetchErr := &novelfetchworkshop.UpstreamError{Class: novelfetchworkshop.UpstreamFailureConfiguration, HTTPStatus: 403, Message: "request rejected"}
	if err := markNovelFetchOriginalFetchRetry(ctx, store, "owner", "book-1", fetchErr, now); err != nil {
		t.Fatal(err)
	}
	document, err := store.GetDocument(ctx, "owner", "book-1")
	if err != nil {
		t.Fatal(err)
	}
	meta = document.Meta
	if meta["status"] != "waiting_configuration" || meta["queueState"] != "waiting_configuration" ||
		meta["manualOnly"] != true || meta["manualOriginalFetchRequired"] != true || meta["originalFetchLastError"] != fetchErr.Error() ||
		!strings.Contains(firstText(meta["error"]), fetchErr.Error()) {
		t.Fatalf("meta=%#v, want configuration hold with cause", meta)
	}
	for _, key := range []string{"originalFetchNextRetryAt", "automationNextRetryAt", "next_retry_at", "retry_schedule"} {
		if _, ok := meta[key]; ok {
			t.Fatalf("configuration hold kept %s: %#v", key, meta)
		}
	}
	if novelFetchAutomationEligible(document, now.Add(24*time.Hour)) {
		t.Fatal("configuration failure must stay paused even with an upload pending")
	}
	queue := novelfetchworkshop.NewMemoryQueue()
	result, err := enqueueWorkshopRetries(ctx, store, queue, "owner", []string{"book-1"}, false)
	if err != nil || result.Queued != 1 {
		t.Fatalf("explicit retry did not queue held document: result=%#v err=%v", result, err)
	}
	document, err = store.GetDocument(ctx, "owner", "book-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"manualOnly", "manualOriginalFetchRequired", "originalFetchRetryAttempts", "originalFetchLastError"} {
		if _, ok := document.Meta[key]; ok {
			t.Fatalf("explicit retry kept %s: %#v", key, document.Meta)
		}
	}
}

func TestTransientOriginalFetchUsesBoundedBackoff(t *testing.T) {
	want := []time.Duration{time.Minute, 5 * time.Minute, 15 * time.Minute, 30 * time.Minute, time.Hour}
	for i, delay := range want {
		if got := novelFetchOriginalRetryDelay(nil, i+1); got != delay {
			t.Fatalf("attempt %d: delay=%s, want %s", i+1, got, delay)
		}
	}
	if got := novelFetchOriginalRetryDelay(nil, 6); got != time.Hour {
		t.Fatalf("attempt above schedule: delay=%s, want 1h", got)
	}
	if got := novelFetchOriginalRetryDelay(nil, 0); got != time.Minute {
		t.Fatalf("attempt below schedule: delay=%s, want 1m", got)
	}
}
func TestTransientOriginalFetchPersistsBackoff(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fetchErr error
		override int
		delays   []time.Duration
	}{
		{name: "typed", fetchErr: &novelfetchworkshop.UpstreamError{Class: novelfetchworkshop.UpstreamFailureTransient, HTTPStatus: 502}, delays: []time.Duration{time.Minute, 5 * time.Minute, 15 * time.Minute, 30 * time.Minute, time.Hour}},
		{name: "untyped", fetchErr: errors.New("connection reset"), delays: []time.Duration{time.Minute, 5 * time.Minute, 15 * time.Minute, 30 * time.Minute, time.Hour}},
		{name: "configured override", fetchErr: errors.New("connection reset"), override: 2, delays: []time.Duration{2 * time.Minute, 2 * time.Minute, 2 * time.Minute, 2 * time.Minute, 2 * time.Minute}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store := novelfetchworkshop.NewMemoryStore()
			if tc.override > 0 {
				if err := store.PutConfig(ctx, "owner", map[string]any{"fetch": map[string]any{"original_fetch_retry_interval_minutes": tc.override}}); err != nil {
					t.Fatal(err)
				}
			}
			if err := store.PutDocument(ctx, "owner", novelfetchworkshop.Document{BookID: "book-1"}); err != nil {
				t.Fatal(err)
			}
			now := time.Date(2026, 11, 23, 3, 0, 0, 0, time.UTC)
			for i, delay := range tc.delays {
				if err := markNovelFetchOriginalFetchRetry(ctx, store, "owner", "book-1", tc.fetchErr, now); err != nil {
					t.Fatal(err)
				}
				document, err := store.GetDocument(ctx, "owner", "book-1")
				if err != nil {
					t.Fatal(err)
				}
				next := now.Add(delay)
				nextRaw := next.Format(time.RFC3339Nano)
				if document.Meta["status"] != "waiting_retry" || document.Meta["queueState"] != "waiting_retry" ||
					intValue(document.Meta["originalFetchRetryAttempts"]) != i+1 || document.Meta["originalFetchNextRetryAt"] != nextRaw ||
					document.Meta["next_retry_at"] != nextRaw || document.Meta["originalFetchLastError"] != tc.fetchErr.Error() {
					t.Fatalf("attempt %d: meta=%#v, want waiting_retry until %s", i+1, document.Meta, nextRaw)
				}
				if novelFetchAutomationEligible(document, next.Add(-time.Second)) || !novelFetchAutomationEligible(document, next) {
					t.Fatalf("attempt %d: automation eligibility did not follow deadline %s", i+1, nextRaw)
				}
				now = next
			}
		})
	}
}

func TestTransientOriginalFetchStopsOnSixthFailure(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	now := time.Date(2026, 11, 23, 3, 0, 0, 0, time.UTC)
	meta := map[string]any{
		"status":                     "processing",
		"originalFetchRetryAttempts": 5,
		"originalFetchNextRetryAt":   now.Format(time.RFC3339Nano),
		"automationNextRetryAt":      now.Format(time.RFC3339Nano),
	}
	scheduleNovelFetchTaskRetry(meta, novelFetchRetryStageOriginalFetch, errors.New("previous fetch"), time.Minute, now)
	if err := store.PutDocument(ctx, "owner", novelfetchworkshop.Document{BookID: "book-1", Meta: meta}); err != nil {
		t.Fatal(err)
	}
	fetchErr := &novelfetchworkshop.UpstreamError{Class: novelfetchworkshop.UpstreamFailureTransient, HTTPStatus: 502}
	if err := markNovelFetchOriginalFetchRetry(ctx, store, "owner", "book-1", fetchErr, now); err != nil {
		t.Fatal(err)
	}
	document, err := store.GetDocument(ctx, "owner", "book-1")
	if err != nil {
		t.Fatal(err)
	}
	meta = document.Meta
	if meta["status"] != "failed" || meta["queueState"] != "failed" || meta["manualOnly"] != true ||
		meta["manualOriginalFetchRequired"] != true || intValue(meta["originalFetchRetryAttempts"]) != 6 {
		t.Fatalf("meta=%#v, want manual hold after sixth failed attempt", meta)
	}
	for _, key := range []string{"originalFetchNextRetryAt", "automationNextRetryAt", "next_retry_at", "retry_schedule"} {
		if _, ok := meta[key]; ok {
			t.Fatalf("exhausted retry budget kept %s: %#v", key, meta)
		}
	}
	if novelFetchAutomationEligible(document, now.Add(24*time.Hour)) {
		t.Fatal("exhausted retry budget must not return to automation")
	}
}

// 用户主动重试后必须能重新回到自动调度，否则 manualOnly 会永久挡住它。
func TestManualRetryClearsAutomationBlockers(t *testing.T) {
	document := novelfetchworkshop.Document{
		BookID: "book-1",
		Meta: map[string]any{
			"status":                    "failed",
			"manualOnly":                true,
			"automationEnqueueAttempts": 20,
			"automationNextRetryAt":     "2026-11-14T03:00:00Z",
		},
	}
	if !resetNovelFetchAutomationRetryState(&document) {
		t.Fatal("expected the retry state to be reset")
	}
	for _, key := range []string{"manualOnly", "automationEnqueueAttempts", "automationNextRetryAt"} {
		if _, ok := document.Meta[key]; ok {
			t.Fatalf("%s must be cleared on a manual retry", key)
		}
	}
	document.Meta["status"] = "waiting_retry"
	if !novelFetchAutomationEligible(document, time.Date(2026, 11, 14, 3, 0, 0, 0, time.UTC)) {
		t.Fatal("a manually retried book must be eligible for automation again")
	}
	if resetNovelFetchAutomationRetryState(&document) {
		t.Fatal("a clean document must not report a reset")
	}
}

// 入队成功后必须清掉退避字段，否则下一次扫描会被旧的截止时间挡住。
func TestMarkNovelFetchTaskQueuedClearsEnqueueBackoff(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(ctx, "alice", novelfetchworkshop.Document{
		BookID: "book-1",
		Meta: map[string]any{
			"status":                    "waiting_retry",
			"automationEnqueueAttempts": 3,
			"automationNextRetryAt":     "2026-11-14T03:30:00Z",
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := markNovelFetchTaskQueued(ctx, store, "alice", "book-1", "job-1", "batch-1", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	document, err := store.GetDocument(ctx, "alice", "book-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := document.Meta["automationNextRetryAt"]; ok {
		t.Fatal("automationNextRetryAt must be cleared after a successful enqueue")
	}
	if _, ok := document.Meta["automationEnqueueAttempts"]; ok {
		t.Fatal("automationEnqueueAttempts must be cleared after a successful enqueue")
	}
	if got := firstText(document.Meta["status"]); got != "queued" {
		t.Fatalf("status=%q, want queued", got)
	}
}
