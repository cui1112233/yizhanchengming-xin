package httpapi

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestClassifyFailureTable(t *testing.T) {
	cases := []struct {
		name            string
		err             error
		stage           string
		wantCode        string
		wantClass       string
		wantRetryBudget bool
		wantNeedsManual bool
		wantMaxAttempts int
	}{
		{"transient default", errors.New("provider boom"), "comic_images", automationCodeTransient, automationFailureTransient, true, false, automationPageRetryMaxAttempts},
		{"rate limited", errors.New("HTTP 429 too many requests"), "comic_images", automationCodeRateLimited, automationFailureTransient, true, false, automationPageRetryMaxAttempts},
		{"network reset", errors.New("read tcp: wsarecv: connection reset by peer"), "comic_images", automationCodeNetwork, automationFailureTransient, true, false, automationPageRetryMaxAttempts},
		{"provider timeout", errors.New("context deadline exceeded"), "comic_images", automationCodeProviderTimeout, automationFailureTransient, true, false, automationRetryEscalationThreshold},
		{"infrastructure", errors.New("Error 1038: Out of sort memory"), "comic_images", automationCodeInfrastructure, automationFailureTransient, true, false, automationPageRetryMaxAttempts},
		{"content policy", errShuihuoImageContentPolicy, "comic_images", automationCodeContentPolicy, automationFailureManual, false, true, 1},
		{"insufficient balance", errors.New("余额不足，请充值"), "comic_images", automationCodeAuthQuota, automationFailureDeterministic, false, false, 1},
		{"page not found", batchfactoryv11.ErrNotFound, "comic_images", automationCodePageNotFound, automationFailureDeterministic, false, false, 1},
		{"page in flight", errors.New("该页正在生成，请勿重复提交"), "comic_images", automationCodePageInFlight, automationFailureTransient, false, false, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyAutomationFailure(tc.err, tc.stage)
			if got.Code != tc.wantCode || got.Class != tc.wantClass {
				t.Fatalf("classify=%+v, want code=%s class=%s", got, tc.wantCode, tc.wantClass)
			}
			if got.CountsTowardRetryBudget != tc.wantRetryBudget {
				t.Fatalf("countsTowardRetryBudget=%v, want %v", got.CountsTowardRetryBudget, tc.wantRetryBudget)
			}
			if got.needsManual() != tc.wantNeedsManual {
				t.Fatalf("needsManual=%v, want %v", got.needsManual(), tc.wantNeedsManual)
			}
			if got.MaxAttempts != tc.wantMaxAttempts {
				t.Fatalf("maxAttempts=%d, want %d", got.MaxAttempts, tc.wantMaxAttempts)
			}
		})
	}
}

// 瞬时失败第 1、2 次只跳过，不重试；第 3 次才进入重试队列。
func TestTransientFailureSkipsThenRetriesAtThird(t *testing.T) {
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	disposition := classifyAutomationFailure(errors.New("provider boom"), "comic_images")
	if disposition.MinAttemptsBeforeRetry != automationRetryEscalationThreshold {
		t.Fatalf("minAttemptsBeforeRetry=%d, want %d", disposition.MinAttemptsBeforeRetry, automationRetryEscalationThreshold)
	}
	ledger := map[string]any{}
	for attempt := 1; attempt < automationRetryEscalationThreshold; attempt++ {
		entry := applyAutomationPageFailure(ledger, "video-1", "漫画页 01", "comic_images",
			disposition.Code, "provider boom", disposition, automationPageImageUnitCost, now)
		if entry["state"] != "skipped" {
			t.Fatalf("attempt %d state=%v, want skipped", attempt, entry["state"])
		}
		if queued, _ := entry["retryQueued"].(bool); queued {
			t.Fatalf("attempt %d 不得进入重试队列", attempt)
		}
	}
	entry := applyAutomationPageFailure(ledger, "video-1", "漫画页 01", "comic_images",
		disposition.Code, "provider boom", disposition, automationPageImageUnitCost, now)
	if entry["state"] != "retrying" {
		t.Fatalf("third attempt state=%v, want retrying", entry["state"])
	}
	if queued, _ := entry["retryQueued"].(bool); !queued {
		t.Fatal("third attempt 必须进入重试队列")
	}
	if attempts, _ := entry["attempts"].(int); attempts != automationRetryEscalationThreshold {
		t.Fatalf("attempts=%d, want %d", attempts, automationRetryEscalationThreshold)
	}
	if cost, _ := entry["cost"].(float64); cost != automationPageImageUnitCost*3 {
		t.Fatalf("cost=%v, want %v", cost, automationPageImageUnitCost*3)
	}
}

// 确定性失败第 1 次即定论，不消耗 3 次预算、不重复调上游。
func TestDeterministicFailureSkipsBudget(t *testing.T) {
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	disposition := classifyAutomationFailure(errShuihuoImageContentPolicy, "comic_images")
	if disposition.CountsTowardRetryBudget {
		t.Fatal("确定性失败不得计入重试预算")
	}
	ledger := map[string]any{}
	entry := applyAutomationPageFailure(ledger, "video-1", "漫画页 01", "comic_images",
		disposition.Code, errShuihuoImageContentPolicy.Error(), disposition, automationPageImageUnitCost, now)
	if entry["state"] != "needs_manual" {
		t.Fatalf("state=%v, want needs_manual", entry["state"])
	}
	if manual, _ := entry["needsManual"].(bool); !manual {
		t.Fatal("确定性失败首次即需人工")
	}
	if queued, _ := entry["retryQueued"].(bool); queued {
		t.Fatal("确定性失败不得进入重试队列")
	}
	// 即使再次记录，也不应升级成 3 次预算后的重试。
	again := applyAutomationPageFailure(ledger, "video-1", "漫画页 01", "comic_images",
		disposition.Code, errShuihuoImageContentPolicy.Error(), disposition, automationPageImageUnitCost, now)
	if again["state"] != "needs_manual" {
		t.Fatalf("deterministic re-record state=%v, want needs_manual", again["state"])
	}
	if again["countsTowardRetryBudget"] != false {
		t.Fatalf("countsTowardRetryBudget=%v, want false", again["countsTowardRetryBudget"])
	}
}

// 「待补」只能在尝试次数用尽后产生，未达上限不允许标记缺失。
func TestPendingRequiresExhaustedAttempts(t *testing.T) {
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	disposition := classifyAutomationFailure(errors.New("provider boom"), "comic_images")
	ledger := map[string]any{}
	for attempt := 1; attempt <= automationPageRetryMaxAttempts; attempt++ {
		entry := applyAutomationPageFailure(ledger, "video-1", "漫画页 01", "comic_images",
			disposition.Code, "provider boom", disposition, automationPageImageUnitCost, now)
		switch {
		case attempt < automationRetryEscalationThreshold:
			if entry["state"] != "skipped" {
				t.Fatalf("attempt %d state=%v, want skipped", attempt, entry["state"])
			}
		default:
			if entry["state"] != "retrying" {
				t.Fatalf("attempt %d state=%v, want retrying", attempt, entry["state"])
			}
		}
		if manual, _ := entry["needsManual"].(bool); manual {
			t.Fatalf("attempt %d 未用尽上限不得转人工", attempt)
		}
	}
	entry := applyAutomationPageFailure(ledger, "video-1", "漫画页 01", "comic_images",
		disposition.Code, "provider boom", disposition, automationPageImageUnitCost, now)
	if entry["state"] != "needs_manual" {
		t.Fatalf("exhausted state=%v, want needs_manual", entry["state"])
	}
	if manual, _ := entry["needsManual"].(bool); !manual {
		t.Fatal("用尽上限后必须转人工")
	}
}

// 成功后当前账期必须清零：先失败 2 次、成功 1 次、再失败 1 次，
// 不得因历史计数直接跳级进入重试队列。
func TestSuccessResetsAttemptBudget(t *testing.T) {
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	disposition := classifyAutomationFailure(errors.New("provider boom"), "comic_images")
	entry := map[string]any{}
	ledger := automationFailureLedger(entry, "comic_images")
	for attempt := 0; attempt < 2; attempt++ {
		applyAutomationPageFailure(ledger, "video-1", "漫画页 01", "comic_images",
			disposition.Code, "provider boom", disposition, automationPageImageUnitCost, now)
	}
	if !resetAutomationPageLedgerNode(entry, "comic_images", now) {
		t.Fatal("成功后必须清零账本")
	}
	record := ledger["video-1"].(map[string]any)
	if record["attempts"] != 0 || record["state"] != "resolved" {
		t.Fatalf("reset 后 attempts=%v state=%v, want 0/resolved", record["attempts"], record["state"])
	}
	if _, ok := record["lastError"]; ok {
		t.Fatal("reset 后 lastError 必须清空")
	}
	if _, ok := record["firstFailedAt"]; ok {
		t.Fatal("reset 后 firstFailedAt 必须清空")
	}
	if lifetime := automationNumberValue(record["lifetimeAttempts"], 0); lifetime != 2 {
		t.Fatalf("lifetimeAttempts=%d, want 2（审计保留）", lifetime)
	}
	if _, ok := record["lastResolvedError"]; !ok {
		t.Fatal("reset 必须保留 lastResolvedError 供审计")
	}
	again := applyAutomationPageFailure(ledger, "video-1", "漫画页 01", "comic_images",
		disposition.Code, "provider boom", disposition, automationPageImageUnitCost, now)
	if attempts, _ := again["attempts"].(int); attempts != 1 {
		t.Fatalf("新一轮首次失败 attempts=%d, want 1", attempts)
	}
	if queued, _ := again["retryQueued"].(bool); queued {
		t.Fatal("新一轮第 1 次失败不得进入重试队列")
	}
}

// 重试 = 幂等续跑：已完成目标不进工作集，本轮已失败的目标不原地重试。
func TestRetryIsContinuationNotRestart(t *testing.T) {
	videos := []batchfactoryv11.Video{
		{ID: "video-1", Label: "漫画页 01"},
		{ID: "video-2", Label: "漫画页 02"},
		{ID: "video-3", Label: "漫画页 03"},
	}
	pages := []batchfactoryv11.ComicPage{
		{VideoID: "video-1", Images: []batchfactoryv11.ComicPageImage{{IsPrimary: true}}},
		{VideoID: "video-2", Images: []batchfactoryv11.ComicPageImage{{IsPrimary: true}}},
		{VideoID: "video-3", Images: []batchfactoryv11.ComicPageImage{}},
	}
	continuation := buildComicPageContinuation(videos, pages)
	if len(continuation.Done) != 2 {
		t.Fatalf("done=%d, want 2", len(continuation.Done))
	}
	if len(continuation.Pending) != 1 || continuation.Pending[0].ID != "video-3" {
		t.Fatalf("pending=%+v, want only video-3", continuation.Pending)
	}
	work := missingComicPageWorkItems(videos, pages)
	if len(work) != 1 || work[0].page.ID != "video-3" {
		t.Fatalf("work=%+v, want only video-3", work)
	}
	filtered := filterFailedComicPageWork(work, map[string]struct{}{"video-3": {}})
	if len(filtered) != 0 {
		t.Fatalf("同一轮内失败页不得原地重试: %+v", filtered)
	}
}

// 超过阈值无进展必须判死，并能在账本里带出位置、原因、次数与花费。
func TestNoSilentStall(t *testing.T) {
	if comicGenerationIdleExtend != automationStallTimeout {
		t.Fatalf("生成租约 idle=%v, want %v", comicGenerationIdleExtend, automationStallTimeout)
	}
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	if automationProgressStalled(now.Add(-automationStallTimeout+time.Second), now) {
		t.Fatal("未超过阈值不应判死")
	}
	if !automationProgressStalled(now.Add(-automationStallTimeout), now) {
		t.Fatal("超过阈值必须判死")
	}
	entry := map[string]any{}
	disposition := classifyAutomationFailure(errors.New("provider boom"), "comic_images")
	ledger := automationFailureLedger(entry, "comic_images")
	record := applyAutomationPageFailure(ledger, "video-7", "漫画页 07", "comic_images",
		disposition.Code, "provider boom", disposition, automationPageImageUnitCost, now)
	message := describeAutomationFailure("comic_images", "video-7", record["attempts"].(int),
		record["lastErrorCode"].(string), record["lastError"].(string), record["cost"].(float64))
	for _, want := range []string{"video-7", "1 次", automationCodeTransient, "¥0.03"} {
		if !strings.Contains(message, want) {
			t.Fatalf("terminal message %q missing %q", message, want)
		}
	}
	if record["label"] != "漫画页 07" || record["stage"] != "comic_images" {
		t.Fatalf("ledger must carry position: %#v", record)
	}
}

// reaper 必须自治：即使 dispatcher 门禁把调度挡住，僵尸 running 也要被回收。
func TestReapZombieJobsIgnoresDispatchNextCheck(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "automation.json")
	now := time.Date(2026, 11, 6, 10, 30, 0, 0, time.UTC)
	state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{
		"job-zombie": {
			"id": "job-zombie", "owner": "alice", "batchId": "batch-1",
			"state": "running", "updatedAt": now.Add(-20 * time.Minute),
			"counts": automationCounts(1, 1, 0, 0),
			"books": map[string]any{
				"book-1": map[string]any{"bookId": "book-1", "status": "failed", "stage": "comic_images"},
			},
		},
	}}
	writeLegacyAutomationState(t, path, state)
	coordinator := newDispatchTestCoordinator(path, now)
	coordinator.dispatchNextCheck = now.Add(time.Hour)
	coordinator.dispatchNextCheckSet = true

	if err := coordinator.DispatchDueJobs(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err := coordinator.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	if automationStringValue(got.Jobs["job-zombie"]["state"], "") != "running" {
		t.Fatal("dispatch 门禁应挡住本次调度")
	}
	if err := coordinator.ReapZombieJobs(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err = coordinator.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	if automationStringValue(got.Jobs["job-zombie"]["state"], "") != "failed" {
		t.Fatalf("reaper 未回收僵尸 job: %#v", got.Jobs["job-zombie"])
	}
}
