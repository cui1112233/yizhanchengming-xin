package httpapi

import (
	"context"
	"errors"
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

// 原文 400 是确定性错误，达到上限必须停止自动重试，避免每小时重复失败。
func TestOriginalFetchRetryStopsAtBudget(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	seed := novelfetchworkshop.Document{BookID: "book-1", Meta: map[string]any{
		"status":                     "processing",
		"originalFetchRetryAttempts": 4,
	}}
	if err := store.PutDocument(ctx, "owner", seed); err != nil {
		t.Fatalf("seed document: %v", err)
	}
	fetchErr := errors.New("原文获取失败: 121 upstream business error 400: 获取书籍信息失败")
	if err := markNovelFetchOriginalFetchRetry(ctx, store, "owner", "book-1", fetchErr, time.Now().UTC()); err != nil {
		t.Fatalf("mark original fetch retry: %v", err)
	}
	document, err := store.GetDocument(ctx, "owner", "book-1")
	if err != nil {
		t.Fatalf("load document: %v", err)
	}
	if document.Meta["manualOnly"] != true || document.Meta["manualOriginalFetchRequired"] != true {
		t.Fatalf("manual flags missing at budget: %#v", document.Meta)
	}
	if document.Meta["status"] != "failed" {
		t.Fatalf("status must be failed at budget, got %v", document.Meta["status"])
	}
	if _, ok := document.Meta["originalFetchNextRetryAt"]; ok {
		t.Fatal("next retry must be cleared once the budget is exhausted")
	}
	if novelFetchAutomationEligible(document, time.Now().UTC()) {
		t.Fatal("a book awaiting manual fetch retry must not be picked up by automation")
	}

	// 未达上限仍保持自动退避，瞬态错误可以自行恢复。
	below := novelfetchworkshop.Document{BookID: "book-2", Meta: map[string]any{
		"status":                     "processing",
		"originalFetchRetryAttempts": 1,
	}}
	if err := store.PutDocument(ctx, "owner", below); err != nil {
		t.Fatalf("seed below-budget document: %v", err)
	}
	if err := markNovelFetchOriginalFetchRetry(ctx, store, "owner", "book-2", fetchErr, time.Now().UTC()); err != nil {
		t.Fatalf("mark below-budget retry: %v", err)
	}
	belowDocument, err := store.GetDocument(ctx, "owner", "book-2")
	if err != nil {
		t.Fatalf("load below-budget document: %v", err)
	}
	if belowDocument.Meta["status"] != "waiting_retry" {
		t.Fatalf("below budget must stay waiting_retry, got %v", belowDocument.Meta["status"])
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
