package httpapi

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

// 失败分类。分类结果决定三件事：本次是否计入重试预算、是否需要人工介入、
// 以及是否允许中止整本书。它是 A 批「快速失败 + 跳过继续」的唯一入口。
const (
	automationFailureTransient     = "transient"
	automationFailureDeterministic = "deterministic"
	automationFailureManual        = "manual"
)

const (
	// automationRetryEscalationThreshold 同一目标累计失败达到该次数后，才进入
	// 自动重试队列。前两次只记账并跳过，不重复调用上游。
	automationRetryEscalationThreshold = 3
	// automationPageRetryMaxAttempts 是单个页级目标的总尝试上限（含升级前的
	// 两次）。达到上限后转 needs_manual，不再自动调用上游。
	automationPageRetryMaxAttempts = 5
	// automationStallTimeout 超过该时长没有任何进展心跳，按停滞处理。
	automationStallTimeout = 10 * time.Minute
	// automationPageImageUnitCost 单页出图成本估算（¥/次），用于费用记账。
	automationPageImageUnitCost = 0.03
)

// 稳定错误码。前端和账本只依赖这些码，不依赖错误文案。
const (
	automationCodeUnknown            = "UNKNOWN"
	automationCodeTransient          = "TRANSIENT_UPSTREAM"
	automationCodeRateLimited        = "RATE_LIMITED"
	automationCodeNetwork            = "NETWORK_ERROR"
	automationCodeProviderTimeout    = "PROVIDER_TIMEOUT"
	automationCodePageInFlight       = "PAGE_IN_FLIGHT"
	automationCodeInfrastructure     = "INFRASTRUCTURE"
	automationCodeAuthQuota          = "AUTH_OR_QUOTA"
	automationCodeContentPolicy      = "CONTENT_POLICY_REJECTED"
	automationCodeSensitiveReview    = "SENSITIVE_REVIEW_PENDING"
	automationCodePublishUnconfirmed = "PUBLISH_UNCONFIRMED"
	automationCodePageNotFound       = "PAGE_NOT_FOUND"
	automationCodePromptMissing      = "PROMPT_MISSING"
	automationCodePromptFailed       = "PROMPT_GENERATION_FAILED"
	automationCodePageRejected       = "PAGE_REJECTED_UPSTREAM"
	automationCodePageShortage       = "PAGE_SHORTAGE"
	automationCodeDeterministic      = "DETERMINISTIC_FAILURE"
)

type automationFailureDisposition struct {
	Code                    string
	Class                   string
	Repairable              bool
	AbortBook               bool
	AbortBatch              bool
	CountsTowardRetryBudget bool
	// MinAttemptsBeforeRetry 为 0 表示不需要累计，直接按 Class 处理；否则
	// 累计失败达到该值之前只跳过、不重试。
	MinAttemptsBeforeRetry int
	MaxAttempts            int
	BaseBackoff            time.Duration
	MaxBackoff             time.Duration
	BudgetKey              string
}

func (d automationFailureDisposition) needsManual() bool {
	return d.Class == automationFailureManual
}

func automationStageBudgetKey(stage string) string {
	switch {
	case strings.Contains(stage, "prompt"):
		return "page_prompt"
	case strings.Contains(stage, "image"), strings.Contains(stage, "comic"):
		return "page_image"
	case strings.Contains(stage, "publish"), strings.Contains(stage, "upload"), strings.Contains(stage, "121"):
		return "publish"
	case strings.Contains(stage, "asset"):
		return "assets"
	default:
		return "stage"
	}
}

// classifyAutomationFailure 把真实错误映射为稳定分类。判定顺序：
// 人工类 → 致命确定性 → 在途协调 → 瞬时 → 确定性兜底。
func classifyAutomationFailure(err error, stage string) automationFailureDisposition {
	budget := automationStageBudgetKey(stage)
	base := automationFailureDisposition{
		Code:                    automationCodeTransient,
		Class:                   automationFailureTransient,
		CountsTowardRetryBudget: true,
		MinAttemptsBeforeRetry:  automationRetryEscalationThreshold,
		MaxAttempts:             automationPageRetryMaxAttempts,
		BaseBackoff:             10 * time.Second,
		MaxBackoff:              5 * time.Minute,
		BudgetKey:               budget,
	}
	if err == nil {
		base.Code = automationCodeUnknown
		base.CountsTowardRetryBudget = false
		return base
	}
	msg := strings.ToLower(err.Error())

	switch {
	case errors.Is(err, errShuihuoImageContentPolicy) || strings.Contains(msg, "content policy") ||
		strings.Contains(msg, "内容审核"):
		return automationFailureDisposition{
			Code: automationCodeContentPolicy, Class: automationFailureManual,
			MaxAttempts: 1, BudgetKey: budget,
		}
	case errors.Is(err, errSensitiveReviewPending), strings.Contains(msg, "敏感词未处理"):
		return automationFailureDisposition{
			Code: automationCodeSensitiveReview, Class: automationFailureManual,
			MaxAttempts: 1, AbortBook: true, BudgetKey: budget,
		}
	case errors.Is(err, errPublishUnconfirmed), strings.Contains(msg, "回读确认失败"),
		strings.Contains(msg, "图片已占用"), strings.Contains(msg, "upload_incomplete"):
		return automationFailureDisposition{
			Code: automationCodePublishUnconfirmed, Class: automationFailureManual,
			MaxAttempts: 1, AbortBook: true, BudgetKey: budget,
		}
	case strings.Contains(msg, "余额不足"), strings.Contains(msg, "insufficient balance"),
		strings.Contains(msg, "payment required"), strings.Contains(msg, "unauthorized"),
		strings.Contains(msg, "forbidden"), strings.Contains(msg, "http 401"),
		strings.Contains(msg, "http 402"), strings.Contains(msg, "http 403"),
		strings.Contains(msg, "image_model_not_configured"):
		return automationFailureDisposition{
			Code: automationCodeAuthQuota, Class: automationFailureDeterministic,
			MaxAttempts: 1, AbortBook: true, AbortBatch: true, BudgetKey: budget,
		}
	case comicPageGenerationInFlightError(err), strings.Contains(msg, "正在生成"),
		strings.Contains(msg, "已有生成任务"):
		return automationFailureDisposition{
			Code: automationCodePageInFlight, Class: automationFailureTransient,
			MaxAttempts: 0, MinAttemptsBeforeRetry: 0, BudgetKey: budget,
		}
	case strings.Contains(msg, "error 1038"), strings.Contains(msg, "out of sort memory"),
		strings.Contains(msg, "error 1213"), strings.Contains(msg, "deadlock"),
		strings.Contains(msg, "bad connection"):
		d := base
		d.Code = automationCodeInfrastructure
		d.BaseBackoff = 5 * time.Second
		d.MaxBackoff = 2 * time.Minute
		return d
	case strings.Contains(msg, "429"), strings.Contains(msg, "rate limit"),
		strings.Contains(msg, "too many requests"):
		d := base
		d.Code = automationCodeRateLimited
		d.BaseBackoff = time.Second
		d.MaxBackoff = time.Minute
		return d
	case strings.Contains(msg, "wsarecv"), strings.Contains(msg, "connection reset"),
		strings.Contains(msg, "malformed http response"), strings.Contains(msg, "unexpected eof"),
		strings.Contains(msg, "i/o timeout"), strings.Contains(msg, "tls handshake timeout"):
		d := base
		d.Code = automationCodeNetwork
		d.BaseBackoff = 3 * time.Second
		d.MaxBackoff = 2 * time.Minute
		return d
	case strings.Contains(msg, "deadline exceeded"), strings.Contains(msg, "超时"):
		d := base
		d.Code = automationCodeProviderTimeout
		d.MaxAttempts = automationRetryEscalationThreshold
		d.BaseBackoff = 30 * time.Second
		d.MaxBackoff = 2 * time.Minute
		return d
	case strings.Contains(msg, "没有任何漫画提示词"), strings.Contains(msg, "还没有漫画提示词"),
		strings.Contains(msg, "缺少独立的图片提示词"):
		return automationFailureDisposition{
			Code: automationCodePromptMissing, Class: automationFailureDeterministic,
			Repairable: true, MaxAttempts: 2, BudgetKey: budget,
		}
	case errors.Is(err, batchfactoryv11.ErrNotFound), strings.Contains(msg, "漫画页不存在"):
		return automationFailureDisposition{
			Code: automationCodePageNotFound, Class: automationFailureDeterministic,
			Repairable: true, MaxAttempts: 1, BudgetKey: budget,
		}
	case isComicPageShortageError(err):
		d := base
		d.Code = automationCodePageShortage
		return d
	}
	if !retryableComicAutomationError(err) {
		return automationFailureDisposition{
			Code: automationCodeDeterministic, Class: automationFailureDeterministic,
			MaxAttempts: 1, BudgetKey: budget,
		}
	}
	return base
}

// applyAutomationPageFailure 是页级/条目级账本的唯一写入口。它保证：
//   - 确定性失败第 1 次即定论，不消耗 3 次预算；
//   - 瞬时失败累计到 3 次才进入重试队列；
//   - 达到总上限后转 needs_manual，绝不无限重试。
func applyAutomationPageFailure(
	ledger map[string]any,
	targetID, label, stage, code, message string,
	disposition automationFailureDisposition,
	cost float64,
	now time.Time,
) map[string]any {
	entry, _ := ledger[targetID].(map[string]any)
	if entry == nil {
		entry = map[string]any{}
	}
	// job_json 重新加载后数字会变成 float64，必须同时兼容 int 与 float64，
	// 否则每次重试都会把 attempts 重置成 1，永远到不了升级阈值。
	attempts := automationNumberValue(entry["attempts"], 0)
	attempts++
	entry["attempts"] = attempts
	entry["lastError"] = message
	entry["lastErrorCode"] = code
	entry["lastErrorAt"] = now.UTC().Format(time.RFC3339)
	entry["stage"] = stage
	entry["label"] = label
	if _, ok := entry["firstFailedAt"]; !ok {
		entry["firstFailedAt"] = now.UTC().Format(time.RFC3339)
	}
	previousCost := 0.0
	switch typed := entry["cost"].(type) {
	case float64:
		previousCost = typed
	case int:
		previousCost = float64(typed)
	}
	entry["cost"] = previousCost + cost

	state := "skipped"
	retryQueued := false
	needsManual := false
	switch {
	case !disposition.CountsTowardRetryBudget:
		// 确定性/人工类不吃重试预算。可本地修复的保持 skipped，等修复动作；
		// 其余直接定论为 needs_manual。
		if disposition.Repairable {
			state = "skipped"
		} else {
			state = "needs_manual"
			needsManual = true
		}
	case attempts < disposition.MinAttemptsBeforeRetry:
		state = "skipped"
	case attempts <= disposition.MaxAttempts:
		state = "retrying"
		retryQueued = true
	default:
		state = "needs_manual"
		needsManual = true
	}
	entry["state"] = state
	entry["retryQueued"] = retryQueued
	entry["needsManual"] = needsManual
	entry["countsTowardRetryBudget"] = disposition.CountsTowardRetryBudget
	ledger[targetID] = entry
	return entry
}

// resetAutomationPageLedgerRecord 在目标成功后清零「当前账期」，但保留
// 累计尝试次数/花费与最后一次已解决错误，审计仍可回答“为什么曾经跳过”。
func resetAutomationPageLedgerRecord(record map[string]any, now time.Time) bool {
	attempts := automationNumberValue(record["attempts"], 0)
	cost := 0.0
	switch typed := record["cost"].(type) {
	case float64:
		cost = typed
	case int:
		cost = float64(typed)
	}
	if attempts == 0 && automationStringValue(record["state"], "") == "resolved" {
		return false
	}
	if attempts > 0 {
		record["lifetimeAttempts"] = automationNumberValue(record["lifetimeAttempts"], 0) + attempts
	}
	if cost > 0 {
		lifetime := 0.0
		switch typed := record["lifetimeCost"].(type) {
		case float64:
			lifetime = typed
		case int:
			lifetime = float64(typed)
		}
		record["lifetimeCost"] = lifetime + cost
	}
	if message := automationStringValue(record["lastError"], ""); message != "" {
		record["lastResolvedError"] = message
		record["lastResolvedErrorCode"] = automationStringValue(record["lastErrorCode"], "")
	}
	record["attempts"] = 0
	record["cost"] = 0
	record["state"] = "resolved"
	record["resolvedAt"] = now.UTC().Format(time.RFC3339)
	record["lastResetAt"] = now.UTC().Format(time.RFC3339)
	record["retryQueued"] = false
	record["needsManual"] = false
	record["countsTowardRetryBudget"] = false
	delete(record, "lastError")
	delete(record, "lastErrorCode")
	delete(record, "lastErrorAt")
	delete(record, "firstFailedAt")
	return true
}

// resetAutomationPageLedgerNode 在某个阶段整体成功后清零该节点的页级账本。
// 只有确认成功才调用；后续新故障会按新的第 1 次重新累计。
func resetAutomationPageLedgerNode(entry map[string]any, node string, now time.Time) bool {
	if entry == nil {
		return false
	}
	nodes, _ := entry["nodes"].(map[string]any)
	if nodes == nil {
		return false
	}
	nodeEntry, _ := nodes[node].(map[string]any)
	if nodeEntry == nil {
		return false
	}
	changed := false
	if ledger, ok := nodeEntry["pageLedger"].(map[string]any); ok {
		for _, raw := range ledger {
			record, _ := raw.(map[string]any)
			if record != nil && resetAutomationPageLedgerRecord(record, now) {
				changed = true
			}
		}
	}
	if _, ok := nodeEntry["attempts"]; ok {
		if resetAutomationPageLedgerRecord(nodeEntry, now) {
			changed = true
		}
	}
	return changed
}

// automationFailureLedgerPath 把页级账本挂到 job_json.nodes.<node>.pageLedger，
// 零迁移。
func automationFailureLedger(entry map[string]any, node string) map[string]any {
	nodes, _ := entry["nodes"].(map[string]any)
	if nodes == nil {
		nodes = map[string]any{}
		entry["nodes"] = nodes
	}
	nodeEntry, _ := nodes[node].(map[string]any)
	if nodeEntry == nil {
		nodeEntry = map[string]any{}
		nodes[node] = nodeEntry
	}
	ledger, _ := nodeEntry["pageLedger"].(map[string]any)
	if ledger == nil {
		ledger = map[string]any{}
		nodeEntry["pageLedger"] = ledger
	}
	return ledger
}

// automationPageLedgerOutcome 汇总一轮页级失败的处理结果。
type automationPageLedgerOutcome struct {
	HasFailures bool
	MaxAttempts int
	RetryQueued bool
	NeedsManual bool
}

// recordAutomationPageLedger 从页级聚合错误中提取失败明细并记账。
// 返回值用于决定本书是"只跳过、稍后续跑"还是"进入重试队列/转人工"。
func recordAutomationPageLedger(entry map[string]any, err error, stage string, now time.Time) automationPageLedgerOutcome {
	outcome := automationPageLedgerOutcome{}
	if entry == nil || err == nil {
		return outcome
	}
	disposition := classifyAutomationFailure(err, stage)
	ledger := automationFailureLedger(entry, firstNonEmpty(stage, "page_image"))
	recordOne := func(nodeLedger map[string]any, videoID, label, itemStage string, itemDisposition automationFailureDisposition, code, message string, cost float64) {
		record := applyAutomationPageFailure(nodeLedger, videoID, label, itemStage,
			code, message, itemDisposition, cost, now)
		outcome.HasFailures = true
		if attempts, ok := record["attempts"].(int); ok && attempts > outcome.MaxAttempts {
			outcome.MaxAttempts = attempts
		}
		if queued, _ := record["retryQueued"].(bool); queued {
			outcome.RetryQueued = true
		}
		if manual, _ := record["needsManual"].(bool); manual {
			outcome.NeedsManual = true
		}
	}
	var pageErr *comicPageWorkError
	if errors.As(err, &pageErr) {
		for _, failure := range pageErr.Failures {
			itemDisposition := classifyAutomationFailure(failure.Err, stage)
			recordOne(ledger, failure.VideoID, failure.Label, stage, itemDisposition,
				itemDisposition.Code, failure.Message(), automationPageImageUnitCost)
		}
	}
	var promptErr *batchfactoryv11.ComicPromptPartialError
	if errors.As(err, &promptErr) {
		promptLedger := automationFailureLedger(entry, firstNonEmpty(stage, "page_prompt"))
		for _, failure := range promptErr.Result.Failed {
			itemDisposition := classifyAutomationFailure(errors.New(failure.Message), stage)
			itemDisposition.Code = failure.Code
			recordOne(promptLedger, failure.VideoID, failure.Label, stage, itemDisposition,
				failure.Code, failure.Message, 0)
		}
	}
	if !outcome.HasFailures {
		// 没有可细分的页级信息时，至少把节点级分类写入，避免静默。
		nodeEntry, _ := entry["nodes"].(map[string]any)
		if nodeEntry == nil {
			nodeEntry = map[string]any{}
			entry["nodes"] = nodeEntry
		}
		nodeEntry[firstNonEmpty(stage, "unknown")] = map[string]any{
			"state":        "failed",
			"errorCode":    disposition.Code,
			"failureClass": disposition.Class,
			"message":      err.Error(),
			"at":           now.UTC().Format(time.RFC3339),
		}
	}
	return outcome
}

func isComicPageShortageError(err error) bool {
	var shortage *comicPageShortageError
	return errors.As(err, &shortage)
}

// automationProgressStalled 判断节点是否静默停滞。超过阈值必须落终态，
// 不允许无限期挂 running。
func automationProgressStalled(lastProgressAt, now time.Time) bool {
	if lastProgressAt.IsZero() {
		return false
	}
	return now.Sub(lastProgressAt) >= automationStallTimeout
}

// automationContinuationBackoff 为未达升级阈值的续跑安排退避。前两次失败
// 只跳过、稍后续跑，不在原地重复调用上游。
func automationContinuationBackoff(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	backoff := 30 * time.Second << (attempts - 1)
	if backoff > 5*time.Minute {
		backoff = 5 * time.Minute
	}
	return backoff
}

// describeAutomationFailure 生成用户可见的一句话，必须包含位置、原因、
// 已试次数与花费，禁止"既没前进也没报错"。
func describeAutomationFailure(stage, targetID string, attempts int, code, message string, cost float64) string {
	return fmt.Sprintf("%s%s 失败 %d 次（%s）：%s，累计花费 ¥%.2f",
		stage, pageTargetSuffix(targetID), attempts, code, message, cost)
}

func pageTargetSuffix(targetID string) string {
	if strings.TrimSpace(targetID) == "" {
		return ""
	}
	return "[" + targetID + "]"
}
