package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

func TestShuihuoImageQuotaShortfallMessage(t *testing.T) {
	raw := `{"error":{"message":"Not enough quota to pre-consume for this request: ¥0.020000 remaining, ¥0.030000 required (request id: 20261112073614)"}}`
	message, ok := shuihuoImageQuotaShortfallMessage(raw)
	if !ok {
		t.Fatal("quota shortfall must be recognised")
	}
	if !strings.Contains(message, "余额不足") || !strings.Contains(message, "0.02") || !strings.Contains(message, "0.03") {
		t.Fatalf("quota message must be Chinese and carry both amounts, got %q", message)
	}
	if _, ok := shuihuoImageQuotaShortfallMessage(`{"error":"invalid api key"}`); ok {
		t.Fatal("plain auth error must not be reported as quota shortfall")
	}
}

func TestAutomationManualReasonTextTranslatesCodes(t *testing.T) {
	message := automationManualReasonText(automationCodeAuthQuota + "/" + automationCodeBilledNotRetrieved)
	if strings.Contains(message, "AUTH_OR_QUOTA") || !strings.Contains(message, "余额不足") {
		t.Fatalf("manual reason must be human readable Chinese, got %q", message)
	}
}

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
		// 超时回收器写进 last_error 的自动重试文案必须保持可重试，不能转人工。
		{"stale timeout auto retry 1", errors.New("漫画页生成超时（自动重试 1/2）"), "comic_images", automationCodeProviderTimeout, automationFailureTransient, true, false, automationRetryEscalationThreshold},
		{"stale timeout auto retry 2", errors.New("漫画页生成超时（自动重试 2/2）"), "comic_images", automationCodeProviderTimeout, automationFailureTransient, true, false, automationRetryEscalationThreshold},
		{"generation interrupted", errors.New(comicPageInterruptedMessage), "comic_images", automationCodeProviderTimeout, automationFailureTransient, true, false, automationRetryEscalationThreshold},
		{"infrastructure", errors.New("Error 1038: Out of sort memory"), "comic_images", automationCodeInfrastructure, automationFailureTransient, true, false, automationPageRetryMaxAttempts},
		// 内容审核拒绝必须走自动改写（最多 comicPromptRepairMaxRounds 轮），
		// 耗尽后由 errComicPromptRepairExhausted 转 needs_manual。
		{"content policy", errShuihuoImageContentPolicy, "comic_images", automationCodeContentPolicy, automationFailureTransient, true, false, comicPromptRepairMaxRounds},
		{"insufficient balance", errors.New("余额不足，请充值"), "comic_images", automationCodeAuthQuota, automationFailureDeterministic, false, false, 1},
		// 上游文案中文化后不能再靠英文关键词；结构化 401/402/403 必须直接判定为确定性失败。
		{"typed provider 402", shuihuoImageUpstreamError{StatusCode: http.StatusPaymentRequired, Message: "上游拒绝了本次请求"}, "comic_images", automationCodeAuthQuota, automationFailureDeterministic, false, false, 1},
		{"typed provider 401", shuihuoImageUpstreamError{StatusCode: http.StatusUnauthorized, Message: "上游拒绝了本次请求"}, "comic_images", automationCodeAuthQuota, automationFailureDeterministic, false, false, 1},
		{"page not found", batchfactoryv11.ErrNotFound, "comic_images", automationCodePageNotFound, automationFailureDeterministic, false, false, 1},
		{"page in flight", errors.New("该页正在生成，请勿重复提交"), "comic_images", automationCodePageInFlight, automationFailureTransient, false, false, 0},
		// 输出被长度上限截断是模型/通道抖动，不能按 ErrUnavailable 判成永久失败：
		// 首次失败即进入续跑队列，允许有限次自动回补。
		{"model output truncated", fmt.Errorf("%w: 文本模型请求失败：文本模型输出被长度截断，请重试或增大输出上限", batchfactoryv11.ErrUnavailable), "comic_assets", automationCodeTransient, automationFailureTransient, true, false, automationRetryEscalationThreshold},
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

func TestTruncatedModelOutputStaysRetryable(t *testing.T) {
	for _, message := range []string{
		"文本模型请求失败：文本模型输出被长度截断，请重试或增大输出上限",
		"文本模型请求失败：文本模型输出被长度截断，已自动压缩重试仍失败。请换用输出上限更大的文本模型后重试",
	} {
		err := fmt.Errorf("%w: %s", batchfactoryv11.ErrUnavailable, message)
		if !retryableComicAutomationError(err) {
			t.Fatalf("%q 必须保持可重试，否则整本会被误判为确定性失败", message)
		}
		if deterministicAutomationFailure(err) {
			t.Fatalf("%q 不得判为确定性失败", message)
		}
	}
}

func TestRecordAutomationPageLedgerKeepsInFlightRetryable(t *testing.T) {
	now := time.Date(2026, 11, 9, 11, 0, 0, 0, time.UTC)
	entry := map[string]any{}
	err := &comicPagePartialError{
		comicPageWorkError: &comicPageWorkError{Failures: []comicPageFailure{{
			VideoID: "video-1",
			Label:   "漫画页 01",
			Err:     fmt.Errorf("%w: 当前漫画页正在生成，请等待本次结果完成后再重试", batchfactoryv11.ErrConflict),
		}}},
		Completed: 34,
		Total:     35,
	}
	outcome := recordAutomationPageLedger(entry, err, "comic_images", now)
	if outcome.NeedsManual {
		t.Fatal("PAGE_IN_FLIGHT must remain eligible for automatic continuation")
	}
	record, _ := automationFailureLedger(entry, "comic_images")["video-1"].(map[string]any)
	if record == nil || record["lastErrorCode"] != automationCodePageInFlight {
		t.Fatalf("record=%#v, want PAGE_IN_FLIGHT ledger entry", record)
	}
	if record["state"] != "skipped" || automationBoolValue(record["needsManual"]) {
		t.Fatalf("record=%#v, want repairable skipped state", record)
	}
}

// 超时回收的自动重试文案必须第一次就进入续跑队列，而不是先跳过再升级。
func TestStaleTimeoutAutoRetryQueuesImmediately(t *testing.T) {
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	for _, message := range []string{
		"漫画页生成超时（自动重试 1/2）",
		"漫画页生成超时（自动重试 2/2）",
		comicPageInterruptedMessage,
	} {
		entry := map[string]any{}
		outcome := recordAutomationPageLedger(entry, &comicPageWorkError{Failures: []comicPageFailure{{
			VideoID: "video-1", Label: "漫画页 01", Err: errors.New(message),
		}}}, "comic_images", now)
		if outcome.NeedsManual {
			t.Fatalf("%q 必须保持可自动续跑", message)
		}
		if !outcome.RetryQueued {
			t.Fatalf("%q 第一次失败就应进入续跑队列", message)
		}
		record, _ := automationFailureLedger(entry, "comic_images")["video-1"].(map[string]any)
		if record == nil || record["state"] != "retrying" || record["lastErrorCode"] != automationCodeProviderTimeout {
			t.Fatalf("%q record=%#v, want retrying PROVIDER_TIMEOUT", message, record)
		}
	}
}

// 停滞与慢速必须分开记账：BODY_STALLED 计入已计费未取回并可记重发次数，
// BODY_SLOW 只记慢，不进入补偿统计。
func TestBilledReadFailureKindRecordedSeparately(t *testing.T) {
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	stalled := &shuihuoImageBilledReadFailure{message: "响应体停滞", kind: shuihuoImageBodyFailureStalled, retried: true}
	slow := &shuihuoImageBilledReadFailure{message: "响应体过慢", kind: shuihuoImageBodyFailureSlow}

	entry := map[string]any{}
	recordAutomationPageLedger(entry, &comicPageWorkError{Failures: []comicPageFailure{
		{VideoID: "p-stalled", Label: "停滞页", Err: fmt.Errorf("页图生成失败: %w", stalled)},
		{VideoID: "p-slow", Label: "慢速页", Err: fmt.Errorf("页图生成失败: %w", slow)},
	}}, "comic_images", now)

	ledger := automationFailureLedger(entry, "comic_images")
	stalledRecord, _ := ledger["p-stalled"].(map[string]any)
	if stalledRecord == nil {
		t.Fatalf("页级账本缺少 p-stalled: %+v", entry["nodes"])
	}
	if stalledRecord["lastErrorCode"] != automationCodeImageBodyStalled {
		t.Fatalf("stalled code=%v, want %s", stalledRecord["lastErrorCode"], automationCodeImageBodyStalled)
	}
	if !automationBoolValue(stalledRecord["billedNotRetrieved"]) {
		t.Fatalf("BODY_STALLED 必须计入已计费未取回: %+v", stalledRecord)
	}
	if got := automationNumberValue(stalledRecord["billedRetryCount"], 0); got != 1 {
		t.Fatalf("billedRetryCount=%d, want 1", got)
	}

	slowRecord, _ := ledger["p-slow"].(map[string]any)
	if slowRecord == nil {
		t.Fatalf("页级账本缺少 p-slow: %+v", entry["nodes"])
	}
	if slowRecord["lastErrorCode"] != automationCodeImageBodySlow {
		t.Fatalf("slow code=%v, want %s", slowRecord["lastErrorCode"], automationCodeImageBodySlow)
	}
	if automationBoolValue(slowRecord["billedNotRetrieved"]) {
		t.Fatalf("BODY_SLOW 不得计入已计费未取回: %+v", slowRecord)
	}
	if !automationBoolValue(slowRecord["bodySlow"]) {
		t.Fatalf("BODY_SLOW 必须记录 bodySlow 标记: %+v", slowRecord)
	}
	if got := automationNumberValue(entry["billedNotRetrievedCount"], -1); got != 1 {
		t.Fatalf("billedNotRetrievedCount=%d, want 1", got)
	}
	// 200 已计费后不再原地重发：停滞与慢速都一样（停滞的唯一一次重发
	// 发生在出图请求内部，由页级/日预算约束）。
	if retryableComicAutomationError(fmt.Errorf("页图生成失败: %w", stalled)) {
		t.Fatal("BODY_STALLED 不得在自动化层原地重发")
	}
	if retryableComicAutomationError(fmt.Errorf("页图生成失败: %w", slow)) {
		t.Fatal("BODY_SLOW 不得在自动化层原地重发")
	}
}

// 已计费但图没取回：必须记进页级账本（稳定错误码 + 已计费标记），
// 计数可读，且不得在原地重发（避免同一张图重复扣费）。
func TestBilledReadFailureRecordedInPageLedger(t *testing.T) {
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	billed := &shuihuoImageBilledReadFailure{message: "200 后读取响应体失败"}
	pageErr := &comicPageWorkError{Failures: []comicPageFailure{
		{VideoID: "p3", Label: "漫画页 03", Code: automationCodePageShortage, Err: fmt.Errorf("页图生成失败: %w", billed)},
	}}
	entry := map[string]any{}
	recordAutomationPageLedger(entry, pageErr, "comic_images", now)

	record, _ := automationFailureLedger(entry, "comic_images")["p3"].(map[string]any)
	if record == nil {
		t.Fatalf("页级账本缺少 p3: %+v", entry["nodes"])
	}
	if record["lastErrorCode"] != automationCodeBilledNotRetrieved {
		t.Fatalf("code=%v, want %s", record["lastErrorCode"], automationCodeBilledNotRetrieved)
	}
	if !automationBoolValue(record["billedNotRetrieved"]) {
		t.Fatalf("缺少已计费未取回标记: %+v", record)
	}
	if got := automationNumberValue(entry["billedNotRetrievedCount"], -1); got != 1 {
		t.Fatalf("billedNotRetrievedCount=%d, want 1", got)
	}
	if retryableComicAutomationError(pageErr) {
		t.Fatal("已计费未取回不得在原地重发")
	}
	// 已计费未取回不允许自动续跑：必须直接落成人工处理，并给出明确文案。
	if record["state"] != "needs_manual" || !automationBoolValue(record["needsManual"]) {
		t.Fatalf("已计费未取回必须转人工处理: %+v", record)
	}
	if record["lastError"] != automationBilledNotRetrievedMessage {
		t.Fatalf("文案=%v, want %q", record["lastError"], automationBilledNotRetrievedMessage)
	}
	if automationBoolValue(record["retryQueued"]) {
		t.Fatalf("已计费未取回不得进入自动续跑队列: %+v", record)
	}

	// 图成功取回后，账本与计数都应清零。
	if !resetAutomationPageLedgerNode(entry, "comic_images", now.Add(time.Minute)) {
		t.Fatal("成功取回后应能重置页级账本")
	}
	if got, ok := entry["billedNotRetrievedCount"].(int); !ok || got != 0 {
		t.Fatalf("重置后 billedNotRetrievedCount=%v, want 0", entry["billedNotRetrievedCount"])
	}
}

// 补缺提示词部分成功：只按页记账并继续出图；一页都没补上时才整本失败。
func TestComicPromptRecoveryPartialOnly(t *testing.T) {
	partial := &batchfactoryv11.ComicPromptPartialError{Result: batchfactoryv11.ComicPromptExtractionResult{
		Updated: []batchfactoryv11.Video{{ID: "v1"}},
		Failed: []batchfactoryv11.ComicPromptExtractionFailure{
			{VideoID: "v2", Label: "漫画页 02", Code: automationCodePromptFailed, Message: "期望 1 页，实际 0 页"},
		},
	}}
	if !comicPromptRecoveryIsPartial(partial, len(partial.Result.Updated)) {
		t.Fatal("部分成功且有补上的页时不得整本失败")
	}
	if comicPromptRecoveryIsPartial(partial, 0) {
		t.Fatal("一页都没补上时必须保持整本失败")
	}
	if comicPromptRecoveryIsPartial(errors.New("boom"), 1) {
		t.Fatal("非部分成功错误不得被当作部分成功")
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
// 注意：内容审核拒绝已改为可自动改写，不再是确定性失败；这里用余额不足。
func TestDeterministicFailureSkipsBudget(t *testing.T) {
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	balanceErr := errors.New("余额不足，请充值")
	disposition := classifyAutomationFailure(balanceErr, "comic_images")
	if disposition.CountsTowardRetryBudget {
		t.Fatal("确定性失败不得计入重试预算")
	}
	ledger := map[string]any{}
	entry := applyAutomationPageFailure(ledger, "video-1", "漫画页 01", "comic_images",
		disposition.Code, balanceErr.Error(), disposition, automationPageImageUnitCost, now)
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
		disposition.Code, balanceErr.Error(), disposition, automationPageImageUnitCost, now)
	if again["state"] != "needs_manual" {
		t.Fatalf("deterministic re-record state=%v, want needs_manual", again["state"])
	}
	if again["countsTowardRetryBudget"] != false {
		t.Fatalf("countsTowardRetryBudget=%v, want false", again["countsTowardRetryBudget"])
	}
}

// 内容审核拒绝第一轮必须进入自动重试，而不是直接 needs_manual。
func TestContentPolicyEntersAutoRepairRetry(t *testing.T) {
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	disposition := classifyAutomationFailure(errShuihuoImageContentPolicy, "comic_images")
	entry := applyAutomationPageFailure(map[string]any{}, "video-1", "漫画页 01", "comic_images",
		disposition.Code, errShuihuoImageContentPolicy.Error(), disposition, automationPageImageUnitCost, now)
	if entry["state"] != "retrying" {
		t.Fatalf("state=%v, want retrying (auto repair)", entry["state"])
	}
	if queued, _ := entry["retryQueued"].(bool); !queued {
		t.Fatal("内容审核拒绝必须进入自动改写队列")
	}
	if manual, _ := entry["needsManual"].(bool); manual {
		t.Fatal("内容审核拒绝第一轮不得转人工")
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

// 书内停滞兜底：先跳过这本书安排续跑，达到升级阈值后转人工，
// 并留下阶段、最后进展时间、被跳过多久等可读状态。
func TestMarkAutomationBookStalledSkipsThenNeedsManual(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	last := now.Add(-automationStallTimeout)
	entry := map[string]any{"status": "running", "stage": "comic_images", "updatedAt": last}
	if !markAutomationBookStalled(entry, "comic_images", last, now) {
		t.Fatal("超过阈值必须标记停滞")
	}
	if entry["status"] != "failed" || !automationBoolValue(entry["retryRequested"]) {
		t.Fatalf("首次停滞必须跳过并安排续跑: %#v", entry)
	}
	if _, ok := entry["retryAt"]; !ok {
		t.Fatalf("首次停滞必须写 retryAt: %#v", entry)
	}
	if automationStringValue(entry["stalledStage"], "") != "comic_images" {
		t.Fatalf("stalledStage=%#v", entry["stalledStage"])
	}
	if automationStringValue(entry["lastProgressAt"], "") != last.UTC().Format(time.RFC3339) {
		t.Fatalf("lastProgressAt=%#v, want %s", entry["lastProgressAt"], last.UTC().Format(time.RFC3339))
	}
	if automationStringValue(entry["stalledFor"], "") != automationStallTimeout.String() {
		t.Fatalf("stalledFor=%#v, want %s", entry["stalledFor"], automationStallTimeout)
	}

	manual := map[string]any{
		"status": "running", "stage": "video",
		"stallSkipCount": automationRetryEscalationThreshold - 1,
	}
	markAutomationBookStalled(manual, "video", last, now)
	if !automationBoolValue(manual["needsManual"]) || manual["status"] != "failed" {
		t.Fatalf("达到升级阈值必须转人工: %#v", manual)
	}
	if _, ok := manual["retryRequested"]; ok {
		t.Fatalf("转人工后不得继续自动续跑: %#v", manual)
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
			"state": "running", "updatedAt": now.Add(-(automationZombieRunningIdleTimeout + time.Minute)),
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

func TestReapZombieJobsReclaimsRestartOrphanAfterShortIdle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "automation.json")
	now := time.Date(2026, 11, 13, 11, 0, 0, 0, time.UTC)
	job := func(id string, idle time.Duration, status string) map[string]any {
		return map[string]any{
			"id": id, "owner": "alice", "batchId": "batch-" + id,
			"state": "running", "updatedAt": now.Add(-idle),
			"counts": automationCounts(1, 1, 0, 0),
			"books": map[string]any{
				"book-1": map[string]any{"bookId": "book-1", "status": status, "stage": "comic_images", "updatedAt": now.Add(-idle)},
			},
		}
	}
	state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{
		// 进程重启后遗留：没有活的 run，且已经 3 分钟没有状态写入。
		"job-orphan": job("job-orphan", 3*time.Minute, "running"),
		// 本进程仍在执行：同样的空闲时长，但有活的 run，不能被回收。
		"job-live": job("job-live", 3*time.Minute, "running"),
		// 刚启动不久：没有 run 但还没到 2 分钟短窗口，先不回收。
		"job-fresh": job("job-fresh", 30*time.Second, "running"),
	}}
	writeLegacyAutomationState(t, path, state)
	coordinator := newDispatchTestCoordinator(path, now)
	coordinator.runCancels = map[string]context.CancelFunc{"job-live": func() {}}

	if err := coordinator.ReapZombieJobs(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err := coordinator.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	if automationStringValue(got.Jobs["job-orphan"]["state"], "") != "failed" {
		t.Fatalf("重启遗留的 running 壳未被回收: %#v", got.Jobs["job-orphan"])
	}
	book, _ := got.Jobs["job-orphan"]["books"].(map[string]any)["book-1"].(map[string]any)
	if automationStringValue(book["status"], "") != "failed" || !strings.Contains(automationStringValue(book["error"], ""), "可直接重试") {
		t.Fatalf("书必须变成可重试的 failed 且给出可读原因: %#v", book)
	}
	for _, id := range []string{"job-live", "job-fresh"} {
		if automationStringValue(got.Jobs[id]["state"], "") != "running" {
			t.Fatalf("%s 不应被回收: %#v", id, got.Jobs[id])
		}
	}
}

func TestRecordAutomationPageLedgerQueueBusyDoesNotConsumePaidRetryBudget(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	entry := map[string]any{}
	failure := comicPageFailure{
		VideoID: "video-4",
		Label:   "漫画页 04",
		Err:     fmt.Errorf("page generation failed: %w", errShuihuoImageQueueBusy),
	}
	err := &comicPagePartialError{
		comicPageWorkError: &comicPageWorkError{Failures: []comicPageFailure{failure}},
		Completed:          22,
		Total:              26,
	}

	outcome := recordAutomationPageLedger(entry, err, "comic_images", now)
	if !outcome.HasFailures || outcome.NeedsManual || outcome.RetryQueued {
		t.Fatalf("queue wait must remain resumable without manual escalation: %#v", outcome)
	}
	nodes := entry["nodes"].(map[string]any)
	ledger := nodes["comic_images"].(map[string]any)["pageLedger"].(map[string]any)
	record := ledger["video-4"].(map[string]any)
	if automationNumberValue(record["attempts"], 0) != 0 {
		t.Fatalf("queue wait must not consume provider attempts: %#v", record)
	}
	if cost, _ := record["cost"].(float64); cost != 0 {
		t.Fatalf("queue wait never reached the provider and must cost zero: %#v", record)
	}
	if automationStringValue(record["state"], "") != "waiting" || automationNumberValue(record["queueWaits"], 0) != 1 {
		t.Fatalf("queue wait must be recorded as waiting: %#v", record)
	}
	if automationStringValue(record["lastErrorCode"], "") != automationCodeImageQueueBusy {
		t.Fatalf("queue wait needs a stable error code: %#v", record)
	}
}

// 上游已出图计费但本地磁盘写满：不得自动重跑，必须转人工并记入已计费未取回。
func TestBilledSaveFailureNeverRetriesAndNeedsManual(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	saveErr := &shuihuoImageBilledSaveFailure{
		message: automationBilledNotRetrievedSaveMessage,
		cause:   fmt.Errorf("%w: 94.0%% used", localartifact.ErrDiskPressure),
	}
	pageErr := &comicPageWorkError{Failures: []comicPageFailure{
		{VideoID: "p-disk", Label: "漫画页 01", Err: fmt.Errorf("页图生成失败: %w", saveErr)},
	}}
	if retryableComicAutomationError(pageErr) {
		t.Fatal("磁盘落盘失败不得自动重跑，否则会重复扣费")
	}
	disp := classifyAutomationFailure(saveErr, "comic_images")
	if disp.Class != automationFailureManual || disp.MaxAttempts != 1 || disp.CountsTowardRetryBudget {
		t.Fatalf("disk pressure disposition=%#v, want manual/1/no-budget", disp)
	}

	entry := map[string]any{}
	outcome := recordAutomationPageLedger(entry, pageErr, "comic_images", now)
	// 整本失败文案必须能拿到真实的人工原因码，而不是聚合错误的分类。
	if !outcome.NeedsManual || outcome.ManualPages != 1 ||
		len(outcome.ManualCodes) != 1 || outcome.ManualCodes[0] != automationCodeBilledNotRetrieved {
		t.Fatalf("manual outcome=%#v", outcome)
	}
	record, _ := automationFailureLedger(entry, "comic_images")["p-disk"].(map[string]any)
	if record == nil {
		t.Fatalf("页级账本缺少 p-disk: %+v", entry["nodes"])
	}
	if billed, _ := record["billedNotRetrieved"].(bool); !billed {
		t.Fatalf("必须记入已计费未取回: %#v", record)
	}
	if needsManual, _ := record["needsManual"].(bool); !needsManual {
		t.Fatalf("磁盘写满必须转人工: %#v", record)
	}
	if got := automationStringValue(record["lastError"], ""); got != automationBilledNotRetrievedSaveMessage {
		t.Fatalf("lastError=%q, want disk save message", got)
	}
}

// 没有经过写保护的 ENOSPC 也必须转人工：否则会被当成瞬时错误无限重试，
// 页面只会一直显示“执行中”，既看不到磁盘原因也无法重试。
func TestRawNoSpaceErrorIsManualInfrastructure(t *testing.T) {
	err := fmt.Errorf("save comic page image: %w", syscall.ENOSPC)
	disp := classifyAutomationFailure(err, "comic_images")
	if disp.Code != automationCodeInfrastructure || disp.Class != automationFailureManual {
		t.Fatalf("disposition=%#v, want infrastructure/manual", disp)
	}
	if automationManualReasonText(disp.Code) == automationCodeInfrastructure {
		t.Fatal("manual reason must be translated to user-facing text")
	}
}
