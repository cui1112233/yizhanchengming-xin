package httpapi

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
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
	// 20 版的 10min 会掐掉「慢图重试」：单次上游请求上限 180s，外层最多 3 次，
	// 单页最长约 9min，再加队列/退避就会越过 10min。取 20min，给单页最坏重试
	// 预算留约 2 倍余量；只要任意页有状态推进、有图落库或上游返回响应，心跳
	// 就会被刷新（见 reportComicGenerationHeartbeat）。
	automationStallTimeout = 20 * time.Minute
	// automationPageImageUnitCost 单页出图成本估算（¥/次），用于费用记账。
	automationPageImageUnitCost = 0.03
)

// 稳定错误码。前端和账本只依赖这些码，不依赖错误文案。
const (
	automationCodeUnknown         = "UNKNOWN"
	automationCodeTransient       = "TRANSIENT_UPSTREAM"
	automationCodeRateLimited     = "RATE_LIMITED"
	automationCodeNetwork         = "NETWORK_ERROR"
	automationCodeProviderTimeout = "PROVIDER_TIMEOUT"
	automationCodeImageQueueBusy  = "IMAGE_QUEUE_BUSY"
	automationCodePageInFlight    = "PAGE_IN_FLIGHT"
	automationCodeInfrastructure  = "INFRASTRUCTURE"
	automationCodeAuthQuota       = "AUTH_OR_QUOTA"
	automationCodeContentPolicy   = "CONTENT_POLICY_REJECTED"
	// automationCodeContentSoftRefusal 是对话式软拒绝，必须与标准 4xx 分开，
	// 前端据它判断“再改写一次”还是“换通道”。
	automationCodeContentSoftRefusal = "CONTENT_SOFT_REFUSAL"
	automationCodeSensitiveReview    = "SENSITIVE_REVIEW_PENDING"
	automationCodePublishUnconfirmed = "PUBLISH_UNCONFIRMED"
	automationCodePageNotFound       = "PAGE_NOT_FOUND"
	automationCodePromptMissing      = "PROMPT_MISSING"
	automationCodePromptFailed       = "PROMPT_GENERATION_FAILED"
	automationCodePageRejected       = "PAGE_REJECTED_UPSTREAM"
	automationCodePageShortage       = "PAGE_SHORTAGE"
	automationCodeDeterministic      = "DETERMINISTIC_FAILURE"
	// automationCodeBilledNotRetrieved 表示上游已按 HTTP 200 计费，但响应体
	// 读取失败、图没有取回。这类页不再自动重发（避免同一页反复扣费），
	// 只落成 needs_manual 等人工处理，必须可观测。
	automationCodeBilledNotRetrieved = "IMAGE_BILLED_NOT_RETRIEVED"
	// BODY_STALLED：连续停滞没有新字节，真断流；已计费，不自动重发。
	// BODY_SLOW：一直有数据但到总上限仍未读完，只记录慢，不算已计费未取回。
	automationCodeImageBodyStalled = "BODY_STALLED"
	automationCodeImageBodySlow    = "BODY_SLOW"
)

// automationBilledNotRetrievedMessage 是「上游已计费但图没取回」的唯一对外文案。
// 这类页不再自动续跑，只允许用户手动重试，避免同一页反复扣费。
const automationBilledNotRetrievedMessage = "上游已出图但我们没收到，不会自动重跑，可手动重试"

// automationBilledNotRetrievedSaveMessage 是「上游已出图并计费，但本地磁盘
// 写满导致落盘失败」的唯一对外文案。它和读失败一样禁止自动续跑。
const automationBilledNotRetrievedSaveMessage = "上游已出图并计费，但服务器磁盘空间不足无法保存；为避免重复扣费不会自动重跑，请清理磁盘后手动重试"

// automationBilledNotRetrievedDisposition 把已计费未取回固定为人工处理：
// 不消耗自动重试预算（CountsTowardRetryBudget=false、Repairable=false），
// applyAutomationPageFailure 会直接落成 needs_manual。
func automationBilledNotRetrievedDisposition(budgetKey string) automationFailureDisposition {
	return automationFailureDisposition{
		Code:                    automationCodeBilledNotRetrieved,
		Class:                   automationFailureManual,
		CountsTowardRetryBudget: false,
		MaxAttempts:             1,
		BudgetKey:               budgetKey,
	}
}

// automationProviderAuthFailure 用结构化状态码识别凭据/余额类确定性失败。
// 上游错误文案已经中文化，不能再依赖 “unauthorized / insufficient balance”
// 这类英文关键词；拿不到状态码的旧错误仍由调用处保留文字兜底。
func automationProviderAuthFailure(err error) bool {
	status, ok := automationErrorHTTPStatus(err)
	if !ok {
		return false
	}
	switch status {
	case http.StatusUnauthorized, http.StatusPaymentRequired, http.StatusForbidden:
		return true
	}
	return false
}

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
	case errors.Is(err, errShuihuoImageQueueBusy):
		// 本机图片名额排队超时，没有向供应商发请求、没有计费，也不是页面本身
		// 生成失败。保持可续跑，不消耗页级失败次数，更不能升级成人工处理。
		return automationFailureDisposition{
			Code: automationCodeImageQueueBusy, Class: automationFailureTransient,
			Repairable: true, CountsTowardRetryBudget: false,
			MaxAttempts: 0, BaseBackoff: 30 * time.Second,
			MaxBackoff: 5 * time.Minute, BudgetKey: budget,
		}
	case errors.Is(err, localartifact.ErrDiskPressure):
		// 磁盘写满时上游可能已经出图计费，重试只会重复扣费；固定为人工处理，
		// 等磁盘清理后由用户手动重试。
		return automationFailureDisposition{
			Code: automationCodeInfrastructure, Class: automationFailureManual,
			MaxAttempts: 1, BudgetKey: budget,
		}
	case errors.Is(err, errComicPromptRepairExhausted):
		// 自动改写轮次用尽：只把这一页转人工，不再调用上游。
		return automationFailureDisposition{
			Code: automationCodeContentPolicy, Class: automationFailureManual,
			MaxAttempts: 1, BudgetKey: budget,
		}
	case errors.Is(err, errShuihuoImageContentSoftRefusal):
		// 对话式软拒绝：先走提示词改写，和硬审核共用轮次预算，但错误码独立。
		return automationFailureDisposition{
			Code: automationCodeContentSoftRefusal, Class: automationFailureTransient,
			CountsTowardRetryBudget: true, MinAttemptsBeforeRetry: 1,
			MaxAttempts: comicPromptRepairMaxRounds, BudgetKey: budget,
		}
	case errors.Is(err, errShuihuoImageContentPolicy) || strings.Contains(msg, "content policy") ||
		strings.Contains(msg, "内容审核"):
		// 硬审核拒绝：每轮改写一次提示词。达到 comicPromptRepairMaxRounds
		// 后由 generatePage 返回 exhausted，页级账本转 needs_manual。
		return automationFailureDisposition{
			Code: automationCodeContentPolicy, Class: automationFailureTransient,
			CountsTowardRetryBudget: true, MinAttemptsBeforeRetry: 1,
			MaxAttempts: comicPromptRepairMaxRounds, BudgetKey: budget,
		}
	case strings.Contains(msg, "敏感词未处理"):
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
	case automationProviderAuthFailure(err), strings.Contains(msg, "余额不足"), strings.Contains(msg, "insufficient balance"),
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
			Code: automationCodePageInFlight, Class: automationFailureTransient, Repairable: true,
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

// comicBilledReadFailure 判断错误链里是否存在「已计费但图没取回」的失败。
// 页级聚合错误只暴露首个失败作为 Unwrap，因此还要展开全部失败明细。
// 这里不区分 BODY_STALLED / BODY_SLOW：两者都是 HTTP 200 已计费，都不允许
// 在自动化层或出图请求内部原地重发。BODY_SLOW 的区别只体现在账本统计上。
func comicBilledReadFailure(err error) bool {
	var billed *shuihuoImageBilledReadFailure
	if errors.As(err, &billed) {
		return true
	}
	var pageErr *comicPageWorkError
	if errors.As(err, &pageErr) {
		for _, failure := range pageErr.Failures {
			if errors.As(failure.Err, &billed) {
				return true
			}
		}
	}
	return false
}

// comicBilledSaveFailure 判断错误链里是否存在「上游已计费但本地落盘失败」。
// 这类失败不能自动重跑：重跑会让上游再生成一张并再次计费。
func comicBilledSaveFailure(err error) bool {
	var billedSave *shuihuoImageBilledSaveFailure
	if errors.As(err, &billedSave) {
		return true
	}
	var pageErr *comicPageWorkError
	if errors.As(err, &pageErr) {
		for _, failure := range pageErr.Failures {
			if errors.As(failure.Err, &billedSave) {
				return true
			}
		}
	}
	return false
}

// refreshAutomationBilledOutstanding 重算当前仍未取回的已计费图数量，写进
// entry 供前端/日志直接读取。成功取回后账本记录会被重置，数量随之下降。
func refreshAutomationBilledOutstanding(entry map[string]any) int {
	if entry == nil {
		return 0
	}
	count := 0
	nodes, _ := entry["nodes"].(map[string]any)
	for _, rawNode := range nodes {
		nodeEntry, _ := rawNode.(map[string]any)
		if nodeEntry == nil {
			continue
		}
		ledger, _ := nodeEntry["pageLedger"].(map[string]any)
		for _, rawRecord := range ledger {
			record, _ := rawRecord.(map[string]any)
			if record != nil && automationBoolValue(record["billedNotRetrieved"]) {
				count++
			}
		}
	}
	entry["billedNotRetrievedCount"] = count
	return count
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
	// 图已成功取回，「已计费未取回」状态结束。
	delete(record, "billedNotRetrieved")
	delete(record, "billedAt")
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
	if changed {
		refreshAutomationBilledOutstanding(entry)
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
	// ManualPages / ManualCodes 记录真正需要人工处理的页数与原因码。
	// 整本失败的文案必须用它们，而不是聚合错误的分类——聚合错误常常是
	// 队列繁忙/超时，会让用户以为整本是超时导致，实际是“已计费未取回”等页。
	ManualPages int
	ManualCodes []string
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
			outcome.ManualPages++
			if code != "" && !slices.Contains(outcome.ManualCodes, code) {
				outcome.ManualCodes = append(outcome.ManualCodes, code)
			}
		}
	}
	var pageErr *comicPageWorkError
	if errors.As(err, &pageErr) {
		for _, failure := range pageErr.Failures {
			if errors.Is(failure.Err, errShuihuoImageQueueBusy) {
				record, _ := ledger[failure.VideoID].(map[string]any)
				if record == nil {
					record = map[string]any{}
					ledger[failure.VideoID] = record
				}
				record["videoId"] = failure.VideoID
				record["label"] = failure.Label
				record["stage"] = stage
				record["state"] = "waiting"
				record["queueWaits"] = automationNumberValue(record["queueWaits"], 0) + 1
				record["lastError"] = failure.Message()
				record["lastErrorCode"] = automationCodeImageQueueBusy
				record["lastErrorAt"] = now.UTC().Format(time.RFC3339)
				record["retryQueued"] = false
				record["needsManual"] = false
				record["countsTowardRetryBudget"] = false
				// 排队阶段没有调用上游，因此没有图片费用，也没有供应商尝试次数。
				record["cost"] = 0.0
				outcome.HasFailures = true
				if outcome.MaxAttempts < 1 {
					outcome.MaxAttempts = 1
				}
				continue
			}
			itemDisposition := classifyAutomationFailure(failure.Err, stage)
			code := itemDisposition.Code
			var billedErr *shuihuoImageBilledReadFailure
			billed := errors.As(failure.Err, &billedErr)
			var billedSaveErr *shuihuoImageBilledSaveFailure
			billedSave := errors.As(failure.Err, &billedSaveErr)
			billedStalled := billed && billedErr.kind == shuihuoImageBodyFailureStalled
			billedSlow := billed && billedErr.kind == shuihuoImageBodyFailureSlow
			switch {
			case billedSave:
				// 上游已出图计费，但本地磁盘写满导致落盘失败。
				code = automationCodeBilledNotRetrieved
			case billedStalled:
				// 真断流：已计费但图没取回，单独记账并计入补偿统计。
				code = automationCodeImageBodyStalled
			case billedSlow:
				// 一直在传但到总上限仍未读完：是「慢」不是「丢」，不标记 billed。
				code = automationCodeImageBodySlow
			case billed:
				code = automationCodeBilledNotRetrieved
			}
			message := failure.Message()
			if billedSave {
				// 磁盘写满时自动续跑只会让同一页反复扣费；直接转人工，
				// 等磁盘清理后由用户手动重试。
				itemDisposition = automationBilledNotRetrievedDisposition(itemDisposition.BudgetKey)
				message = automationBilledNotRetrievedSaveMessage
			} else if billed {
				// 上游已按 HTTP 200 计费，图却没落到本地：自动续跑只会让同一页
				// 反复付费。这里直接定论为人工处理，仅保留“可手动重试”的入口。
				itemDisposition = automationBilledNotRetrievedDisposition(itemDisposition.BudgetKey)
				message = automationBilledNotRetrievedMessage
			}
			recordOne(ledger, failure.VideoID, failure.Label, stage, itemDisposition,
				code, message, automationPageImageUnitCost)
			if record, _ := ledger[failure.VideoID].(map[string]any); record != nil {
				if billedSave {
					record["billedNotRetrieved"] = true
					record["billedAt"] = now.UTC().Format(time.RFC3339)
					record["billedSaveBlocked"] = "disk_pressure"
					log.Printf("automation: image billed but save blocked by disk pressure page=%s label=%s stage=%s", failure.VideoID, failure.Label, stage)
				}
				// BODY_SLOW 是「在传但慢」，不算已计费未取回；其余 200 后读失败
				// （BODY_STALLED 或截断/连接重置等）都必须记入补偿统计。
				if billed && !billedSlow {
					record["billedNotRetrieved"] = true
					record["billedAt"] = now.UTC().Format(time.RFC3339)
					if billedErr.retried {
						record["billedRetryCount"] = automationNumberValue(record["billedRetryCount"], 0) + 1
						record["billedRetriedAt"] = now.UTC().Format(time.RFC3339)
					}
					log.Printf("automation: image billed but body not retrieved page=%s label=%s stage=%s kind=%s retried=%v", failure.VideoID, failure.Label, stage, billedErr.kind, billedErr.retried)
				}
				if billedSlow {
					record["bodySlow"] = true
					record["bodySlowAt"] = now.UTC().Format(time.RFC3339)
					log.Printf("automation: image body slow (not counted as billed loss) page=%s label=%s stage=%s", failure.VideoID, failure.Label, stage)
				}
			}
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
	refreshAutomationBilledOutstanding(entry)
	return outcome
}

// recordComicAssetShortfall 把缺失的资产参考图写进 comic_asset_images 账本。
// 资产图缺失不阻断整本，但必须可读、可回补，并与页图共用同一账本口径。
func recordComicAssetShortfall(entry map[string]any, shortfall comicAssetShortfall, now time.Time) {
	if entry == nil || len(shortfall.MissingAssetIDs) == 0 {
		return
	}
	disposition := automationFailureDisposition{
		Code: automationCodePageShortage, Class: automationFailureTransient,
		CountsTowardRetryBudget: true, MinAttemptsBeforeRetry: automationRetryEscalationThreshold,
		MaxAttempts: automationPageRetryMaxAttempts, BaseBackoff: 10 * time.Second,
		MaxBackoff: 5 * time.Minute, BudgetKey: "assets",
	}
	code := disposition.Code
	message := "资产参考图未写回"
	if shortfall.Err != nil {
		message = shortfall.Err.Error()
	}
	if shortfall.Billed {
		// 资产图同样不允许「已计费未取回」自动回补，否则每轮都会重复付费。
		code = automationCodeBilledNotRetrieved
		disposition = automationBilledNotRetrievedDisposition(disposition.BudgetKey)
		message = automationBilledNotRetrievedMessage
	}
	ledger := automationFailureLedger(entry, "comic_asset_images")
	for _, assetID := range shortfall.MissingAssetIDs {
		record := applyAutomationPageFailure(ledger, assetID, assetID, "comic_asset_images",
			code, message, disposition, 0, now)
		if shortfall.Billed {
			record["billedNotRetrieved"] = true
			record["billedAt"] = now.UTC().Format(time.RFC3339)
		}
	}
	entry["assetShortfallCount"] = len(shortfall.MissingAssetIDs)
	entry["assetShortfallAt"] = now.UTC().Format(time.RFC3339)
	refreshAutomationBilledOutstanding(entry)
	log.Printf("automation: comic asset images missing count=%d critical=%d/%d billed=%v",
		len(shortfall.MissingAssetIDs), shortfall.CriticalMissing, shortfall.CriticalTotal, shortfall.Billed)
}

// recordComicPromptRepair 把“这一页被自动改写过”写进 comic_images 页级账本。
// 与失败账本共用 pageLedger：成功清零只重置当前账期，不会抹掉修复历史。
func recordComicPromptRepair(entry map[string]any, repair comicPromptRepair, now time.Time) {
	if entry == nil || strings.TrimSpace(repair.VideoID) == "" {
		return
	}
	ledger := automationFailureLedger(entry, "comic_images")
	record, _ := ledger[repair.VideoID].(map[string]any)
	if record == nil {
		record = map[string]any{"videoId": repair.VideoID, "label": repair.Label}
		ledger[repair.VideoID] = record
	}
	record["autoRepaired"] = true
	attempts := automationNumberValue(record["repairAttempts"], 0)
	if repair.Attempts > attempts {
		attempts = repair.Attempts
	} else {
		attempts++
	}
	record["repairAttempts"] = attempts
	record["repairCount"] = attempts
	record["repairedFields"] = append([]string{}, repair.Fields...)
	record["repairBefore"] = comicPromptSummary(repair.Before)
	record["repairAfter"] = comicPromptSummary(repair.After)
	if repair.LastRefusalRaw != "" {
		record["lastRefusalRaw"] = repair.LastRefusalRaw
	}
	record["repairedAt"] = now.UTC().Format(time.RFC3339)
	log.Printf("automation: comic prompt auto repaired page=%s attempts=%d fields=%v", repair.VideoID, attempts, repair.Fields)
}

// comicPromptSummary 限制账本里的提示词长度，保留开头便于人工对照。
func comicPromptSummary(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	const limit = 160
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + fmt.Sprintf("…（已截断，原长 %d 字）", len(runes))
}

func isComicPageShortageError(err error) bool {
	var shortage *comicPageShortageError
	return errors.As(err, &shortage)
}

// automationNeedsManualPages 从所有书页级账本汇总“需人工页清单”，
// 供批次状态直接展示：哪本书、哪一页、什么原因、改过几次、最后拒绝原文。
func automationNeedsManualPages(books map[string]any) []map[string]any {
	list := make([]map[string]any, 0)
	for bookID, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		nodes, _ := entry["nodes"].(map[string]any)
		for nodeName, nodeRaw := range nodes {
			node, _ := nodeRaw.(map[string]any)
			if node == nil {
				continue
			}
			pageLedger, _ := node["pageLedger"].(map[string]any)
			for videoID, pageRaw := range pageLedger {
				page, _ := pageRaw.(map[string]any)
				if page == nil {
					continue
				}
				if needsManual, _ := page["needsManual"].(bool); !needsManual {
					continue
				}
				item := map[string]any{
					"bookId":         bookID,
					"bookTitle":      automationStringValue(entry["title"], ""),
					"node":           nodeName,
					"videoId":        videoID,
					"label":          automationStringValue(page["label"], ""),
					"code":           automationStringValue(page["code"], ""),
					"message":        batchfactoryv11.UserFacingErrorText(automationStringValue(page["message"], "")),
					"repairAttempts": automationNumberValue(page["repairAttempts"], 0),
					"lastRefusalRaw": automationStringValue(page["lastRefusalRaw"], ""),
				}
				if fields, ok := page["repairedFields"].([]string); ok {
					item["repairedFields"] = append([]string{}, fields...)
				}
				list = append(list, item)
			}
		}
	}
	sort.Slice(list, func(i, j int) bool {
		left := automationStringValue(list[i]["bookId"], "") + "|" + automationStringValue(list[i]["videoId"], "")
		right := automationStringValue(list[j]["bookId"], "") + "|" + automationStringValue(list[j]["videoId"], "")
		return left < right
	})
	return list
}

// automationProgressStalled 判断节点是否静默停滞。超过阈值必须落终态，
// 不允许无限期挂 running。
func automationProgressStalled(lastProgressAt, now time.Time) bool {
	if lastProgressAt.IsZero() {
		return false
	}
	return now.Sub(lastProgressAt) >= automationStallTimeout
}

// errAutomationBookStalled 是书内停滞看门狗取消执行时携带的原因。它让所有
// 阶段错误路径统一走「跳过这本书」而不是按上游错误重试。
var errAutomationBookStalled = errors.New("书籍长时间没有进展")

// automationStallWatchInterval 是停滞看门狗检查进展的间隔。判定阈值仍是
// automationStallTimeout，这里只控制检查频率。
const automationStallWatchInterval = 30 * time.Second

// markAutomationBookStalled 把书内停滞落成可读状态：阶段、最后进展时间、
// 被跳过多久。前两次安排续跑让其它书继续，达到升级阈值后转人工处理。
func markAutomationBookStalled(entry map[string]any, stage string, lastProgress, now time.Time) bool {
	if entry == nil {
		return false
	}
	if stage == "" {
		stage = automationStringValue(entry["stage"], "unknown")
	}
	skippedFor := now.Sub(lastProgress).Round(time.Second)
	count := automationNumberValue(entry["stallSkipCount"], 0) + 1
	entry["stallSkipCount"] = count
	entry["stalledStage"] = stage
	entry["lastProgressAt"] = lastProgress.UTC().Format(time.RFC3339)
	entry["stalledAt"] = now.UTC().Format(time.RFC3339)
	entry["stalledFor"] = skippedFor.String()
	entry["stage"], entry["error"], entry["updatedAt"] = stage, errAutomationBookStalled.Error(), now

	message := fmt.Sprintf("%s 阶段 %s 没有进展，已跳过这本书让其它书继续", stage, skippedFor)
	if count < automationRetryEscalationThreshold {
		backoff := automationContinuationBackoff(count)
		entry["status"] = "failed"
		entry["message"] = fmt.Sprintf("%s；%d 秒后继续（第 %d/%d 次）",
			message, int(backoff.Seconds()), count, automationRetryEscalationThreshold)
		entry["retryRequested"] = true
		entry["retryAt"] = now.Add(backoff)
		appendAutomationBookLog(entry, now, "stalled", automationStringValue(entry["message"], ""))
		return true
	}
	entry["status"] = "failed"
	entry["message"] = fmt.Sprintf("%s；连续 %d 次停滞，需人工处理", message, count)
	entry["needsManual"] = true
	delete(entry, "retryRequested")
	delete(entry, "retryAt")
	appendAutomationBookLog(entry, now, "stalled", automationStringValue(entry["message"], ""))
	return true
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
