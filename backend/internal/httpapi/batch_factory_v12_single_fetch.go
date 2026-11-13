package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

// 批量工厂自己的敏感词处理配置：批次级设置为主，单书设置按现有继承体系覆盖。
// 旧批次没有这三个字段时按默认值读取（关闭 AI 修复、继承全局模型、失败自动打码并继续）。
// 显式写入 sensitiveAIEnabled=true 的老批次仍保持开启，默认值只作用于未设置的情况。
type batchSensitiveConfig struct {
	Enabled   bool
	Models    []string
	OnFailure string
}

const (
	batchSensitiveOnFailureReview = "review"
	batchSensitiveOnFailureFail   = "fail"
)

func defaultBatchSensitiveConfig() batchSensitiveConfig {
	return batchSensitiveConfig{Enabled: false, OnFailure: batchSensitiveOnFailureReview}
}

func resolveBatchSensitiveConfig(layers ...batchfactoryv11.SettingsPatch) batchSensitiveConfig {
	cfg := defaultBatchSensitiveConfig()
	effective := batchfactoryv11.ResolveSettings(layers...)
	if raw, ok := effective["sensitiveAIEnabled"]; ok && len(raw) > 0 {
		var enabled bool
		if json.Unmarshal(raw, &enabled) == nil {
			cfg.Enabled = enabled
		}
	}
	if raw, ok := effective["sensitiveAIModels"]; ok && len(raw) > 0 {
		var models []string
		if json.Unmarshal(raw, &models) == nil {
			for _, model := range models {
				if trimmed := strings.TrimSpace(model); trimmed != "" {
					cfg.Models = append(cfg.Models, trimmed)
				}
			}
		} else {
			var single string
			if json.Unmarshal(raw, &single) == nil {
				for _, model := range strings.Split(single, ",") {
					if trimmed := strings.TrimSpace(model); trimmed != "" {
						cfg.Models = append(cfg.Models, trimmed)
					}
				}
			}
		}
	}
	if raw, ok := effective["sensitiveAIOnFailure"]; ok && len(raw) > 0 {
		var strategy string
		if json.Unmarshal(raw, &strategy) == nil && strings.EqualFold(strings.TrimSpace(strategy), batchSensitiveOnFailureFail) {
			cfg.OnFailure = batchSensitiveOnFailureFail
		}
	}
	return cfg
}

// batchSensitiveOutcome 记录本次处理的降级结果，供界面显示“有 N 处未处理已保留原文”。
type batchSensitiveOutcome struct {
	ReviewRequired bool
	ReviewCount    int
	AIModels       []string
}

// processBatchConfiguredText 是批量工厂的三级降级链路：
// 1) AI 修复（开关打开且配置了可用模型，按模型顺序逐个尝试）；
// 2) 确定性规则脱敏（规则结果已不含命中词则视为完成）；
// 3) 仍不安全时打码并标记待人工确认，绝不把未处理的原样内容当成合规内容。
func processBatchConfiguredText(ctx context.Context, store novelfetchworkshop.Store, owner, bookID, text, scope string, cfg batchSensitiveConfig, provider novelfetchworkshop.RewriteProvider) (string, []map[string]any, []map[string]any, batchSensitiveOutcome, error) {
	outcome := batchSensitiveOutcome{}
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return text, nil, nil, outcome, err
	}
	processed, trace, items := novelfetchworkshop.ProcessConfiguredText(text, scope, settings)
	if len(items) == 0 {
		return processed, trace, items, outcome, nil
	}
	if cfg.Enabled && provider != nil {
		models := append([]string(nil), cfg.Models...)
		if len(models) == 0 {
			if model := firstText(mapOrEmpty(settings["sensitive_ai"])["model"], mapOrEmpty(mapOrEmpty(settings["appConfig"])["sensitive_ai"])["model"], settings["text_model_id"], settings["textModelId"]); model != "" {
				models = []string{model}
			}
		}
		for _, model := range models {
			ai := mapOrEmpty(settings["sensitive_ai"])
			if len(ai) == 0 {
				ai = mapOrEmpty(mapOrEmpty(settings["appConfig"])["sensitive_ai"])
			}
			attemptSettings := make(map[string]any, len(settings)+1)
			for key, value := range settings {
				attemptSettings[key] = value
			}
			attemptAI := make(map[string]any, len(ai)+2)
			for key, value := range ai {
				attemptAI[key] = value
			}
			attemptAI["enabled"] = true
			attemptAI["model"] = model
			attemptSettings["sensitive_ai"] = attemptAI
			before := processed
			processed, items = applySensitiveAIRepair(ctx, provider, attemptSettings, bookID, processed, items)
			failed := 0
			for _, item := range items {
				if strings.EqualFold(strings.TrimSpace(fmt.Sprint(item["status"])), "failed") {
					failed++
				}
			}
			trace = append(trace, map[string]any{"stage": "sensitive_ai", "title": "敏感词 AI 修复", "model": model, "changed": before != processed, "chars": len([]rune(processed)), "failed": failed})
			if failed == 0 {
				outcome.AIModels = append(outcome.AIModels, model)
				break
			}
		}
	}
	unresolved := 0
	keywords := make([]string, 0, len(items))
	for _, item := range items {
		keyword := strings.TrimSpace(fmt.Sprint(item["keyword"]))
		if keyword == "<nil>" {
			keyword = ""
		}
		finalText := strings.TrimSpace(fmt.Sprint(item["after"]))
		if finalText == "<nil>" {
			finalText = ""
		}
		// 命中词已不在替换结果里才算真正脱敏（AI 结果由 AI 层自行校验）。
		if keyword != "" && !containsSensitiveKeyword(finalText, keyword) {
			if !strings.EqualFold(strings.TrimSpace(fmt.Sprint(item["status"])), "done") {
				item["status"] = "done"
				item["source"] = "rules"
			}
			continue
		}
		if keyword == "" && strings.EqualFold(strings.TrimSpace(fmt.Sprint(item["status"])), "done") {
			continue
		}
		// 规则结果仍含命中词：打码兜底，同时保留“未处理”标记供人工确认。
		item["status"] = "failed"
		item["source"] = "mask"
		item["masked"] = true
		if keyword != "" {
			keywords = append(keywords, keyword)
		}
		unresolved++
	}
	if unresolved > 0 {
		if cfg.OnFailure == batchSensitiveOnFailureFail {
			return processed, trace, items, outcome, fmt.Errorf("敏感词 AI 修复失败 %d 处，已停止后续处理", unresolved)
		}
		for _, keyword := range keywords {
			processed = strings.ReplaceAll(processed, keyword, strings.Repeat("＊", len([]rune(keyword))))
		}
		outcome.ReviewRequired = true
		outcome.ReviewCount = unresolved
	}
	return processed, trace, items, outcome, nil
}

type batchFetchClassification struct {
	AIGenderJudge         novelfetchworkshop.AIGenderJudge
	AIGenderJudgeResolver func(context.Context, string) (novelfetchworkshop.AIGenderJudge, error)
	RewriteProvider       novelfetchworkshop.RewriteProvider
	RewriteResolver       func(context.Context, string) (novelfetchworkshop.RewriteProvider, error)
}

func captureBatchFetchedSource(ctx context.Context, store batchfactoryv11.Store, owner, batchID, bookID string, input batchfactoryv11.CaptureBookSourceInput) (batchfactoryv11.Book, error) {
	updated, err := store.CaptureBookSource(ctx, owner, batchID, bookID, input)
	if !errors.Is(err, batchfactoryv11.ErrConflict) {
		return updated, err
	}
	latest, getErr := store.GetBatch(ctx, owner, batchID)
	if getErr != nil {
		return batchfactoryv11.Book{}, getErr
	}
	for index := range latest.Books {
		book := &latest.Books[index]
		if book.ID != bookID {
			continue
		}
		// A concurrent fetch already won. Reuse it rather than replacing a valid
		// source with the slower request's response.
		if strings.TrimSpace(book.SourceText) != "" {
			return *book, nil
		}
		input.ExpectedRevision = book.Revision
		return store.CaptureBookSource(ctx, owner, batchID, bookID, input)
	}
	return batchfactoryv11.Book{}, batchfactoryv11.ErrNotFound
}

// fetchedBookSourceTitle 优先使用上游 bookinfo.book_name；旧数据没有书名时
// 保留原标题，交给 store 的占位名判断决定是否覆盖。
func fetchedBookSourceTitle(result novelfetchworkshop.FetchAndProcessResult, fallback string) string {
	if title := strings.TrimSpace(result.BookInfo.WorkTitle); title != "" {
		return title
	}
	return fallback
}

func prepareBatchFetchedSource(ctx context.Context, workshopStore novelfetchworkshop.Store, owner, bookID string, result novelfetchworkshop.FetchAndProcessResult, classification batchFetchClassification, sensitive batchSensitiveConfig) (novelfetchworkshop.FetchAndProcessResult, map[string]any, error) {
	metadata := map[string]any{
		"originalRawStatus":        "done",
		"originalRawChars":         result.OriginalRawChars,
		"upstreamCode":             result.Code,
		"upstreamMessage":          result.Message,
		"sourceMode":               "novel_fetch",
		"maxTxt":                   result.MaxTxt,
		"processingConfigRevision": "novel-fetch-rules-v1",
		"sensitiveAIEnabled":       sensitive.Enabled,
		"sensitiveAIOnFailure":     sensitive.OnFailure,
	}
	if len(sensitive.Models) > 0 {
		metadata["sensitiveAIModels"] = sensitive.Models
	}
	if result.BookInfo.Category != "" {
		metadata["category"] = result.BookInfo.Category
	}
	if result.BookInfo.Genre != "" {
		metadata["genre"] = result.BookInfo.Genre
	}
	// 抓正文的上游响应里本来就带 bookinfo.book_name；顺手回填书名，
	// 避免书城导入的书一直停留在「小说 <id>」这类占位名。
	if title := strings.TrimSpace(result.BookInfo.WorkTitle); title != "" {
		metadata["bookName"] = title
	}
	if workshopStore == nil {
		metadata["originalStatus"] = "done"
		metadata["queueStage"] = "classification"
		metadata["sourceContentVersion"] = "original"
		metadata["processingStatus"] = "not_configured"
		metadata["processedChars"] = result.ProcessedChars
		metadata["processingFingerprint"] = novelFetchProcessingFingerprint(nil)
		metadata = finalizeFetchedBatchClassification(classifyFetchedBatchBook(ctx, owner, result.BookInfo, result.ProcessedText, metadata, classification))
		return result, metadata, nil
	}
	settings, err := workshopStore.GetConfig(ctx, owner)
	if err != nil {
		metadata["originalStatus"] = "processing_failed"
		metadata["queueStage"] = "original_processing"
		metadata["processingStatus"] = "failed"
		metadata["processRetryable"] = true
		metadata["processError"] = err.Error()
		return result, metadata, err
	}
	provider := providerForBatchClassification(ctx, owner, classification)
	metadata["processingFingerprint"] = novelFetchProcessingFingerprint(settings)
	processed, trace, sensitiveItems, outcome, processErr := processBatchConfiguredText(
		ctx, workshopStore, owner, bookID, result.ProcessedText, "original", sensitive, provider,
	)
	result.ProcessedText = processed
	result.ProcessedChars = len([]rune(processed))
	metadata["originalRawChars"] = result.OriginalRawChars
	metadata["processedChars"] = result.ProcessedChars
	metadata["processingStages"] = summarizeBatchProcessingStages(trace)
	metadata["sensitiveStatus"] = "done"
	replacedCount := 0
	for _, item := range sensitiveItems {
		if !strings.EqualFold(strings.TrimSpace(fmt.Sprint(item["status"])), "failed") {
			replacedCount++
		}
	}
	metadata["sensitiveReplaced"] = replacedCount
	metadata["sensitiveHitCount"] = len(sensitiveItems)
	setSensitiveDetails(metadata, sensitiveItems, len(sensitiveItems))
	metadata["sourceContentVersion"] = "sensitive"
	metadata["processingVersion"] = "novel-fetch-rules-v1"
	if processErr != nil {
		metadata["originalStatus"] = "processing_failed"
		metadata["queueStage"] = "original_processing"
		metadata["processingStatus"] = "failed"
		metadata["processRetryable"] = true
		metadata["processError"] = processErr.Error()
		metadata["sensitiveStatus"] = "partial"
		return result, metadata, processErr
	}
	// 三级降级：未能安全处理的命中原样保留正文并打码，标记待人工确认后继续后续阶段，
	// 不再把整本书报废成“原文待获取”。
	metadata["sensitiveReviewRequired"] = outcome.ReviewRequired
	metadata["sensitiveReviewCount"] = outcome.ReviewCount
	if len(outcome.AIModels) > 0 {
		metadata["sensitiveAIModelsApplied"] = outcome.AIModels
	}
	if outcome.ReviewRequired {
		metadata["sensitiveStatus"] = "review_required"
		metadata["sensitiveReviewMessage"] = fmt.Sprintf("有 %d 处敏感词未能替换，已自动打码，不阻断自动化", outcome.ReviewCount)
	}
	metadata["originalStatus"] = "done"
	metadata["queueStage"] = "classification"
	metadata["processingStatus"] = "done"
	metadata["processRetryable"] = false
	metadata = finalizeFetchedBatchClassification(classifyFetchedBatchBook(ctx, owner, result.BookInfo, result.ProcessedText, metadata, classification))
	return result, metadata, nil
}

func summarizeBatchProcessingStages(trace []map[string]any) []any {
	stages := make([]any, 0, len(trace))
	for _, item := range trace {
		stages = append(stages, map[string]any{
			"stage": item["stage"], "title": item["title"],
			"chars": item["chars"], "changed": item["changed"],
		})
	}
	return stages
}

func batchFetchResultFromStoredRaw(book batchfactoryv11.Book, maxTxt int) novelfetchworkshop.FetchAndProcessResult {
	if maxTxt <= 0 {
		maxTxt = 4000
	}
	processed := novelfetchworkshop.ProcessRawText(book.TxtText, maxTxt)
	return novelfetchworkshop.FetchAndProcessResult{
		OriginalRaw: processed.OriginalRaw, ProcessedText: processed.ProcessedText,
		OriginalRawChars: processed.OriginalRawChars, ProcessedChars: processed.ProcessedChars, MaxTxt: processed.MaxTxt,
		BookInfo: novelfetchworkshop.BookInfo{
			Category: firstMetadataString(book.SourceMetadata, "category"),
			Genre:    firstMetadataString(book.SourceMetadata, "genre"),
			Themes:   firstMetadataString(book.SourceMetadata, "themes", "theme"),
			Labels:   firstMetadataString(book.SourceMetadata, "labels", "tags"),
		},
	}
}

// fetchBatchSourceWithRetry is shared by manual fetch, bulk fetch, and the
// production auto-fetch fallback. A transient upstream failure must not make
// those three entry points behave differently.
func fetchBatchSourceWithRetry(ctx context.Context, fetcher *novelfetchworkshop.UpstreamFetcher, bookID, platform string, maxTxt int) (novelfetchworkshop.FetchAndProcessResult, error, int) {
	if fetcher == nil {
		return novelfetchworkshop.FetchAndProcessResult{}, fmt.Errorf("原文抓取服务未配置"), 0
	}
	var result novelfetchworkshop.FetchAndProcessResult
	var err error
	attempts := 0
	for attempt := 0; attempt < 2; attempt++ {
		attempts++
		result, err = fetcher.FetchAndProcess(ctx, bookID, platform, maxTxt)
		if err == nil || ctx.Err() != nil || !retryableBatchFetchError(err) {
			break
		}
	}
	return result, err, attempts
}

func persistBatchFetchFailure(ctx context.Context, store batchfactoryv11.Store, owner, batchID string, book batchfactoryv11.Book, err error, attempts int) {
	if store == nil || book.ID == "" || err == nil {
		return
	}
	metadata := copyAnyMap(book.SourceMetadata)
	metadata["originalStatus"] = "failed"
	metadata["originalRawStatus"] = "failed"
	metadata["queueStage"] = "original"
	metadata["fetchError"] = err.Error()
	technicalErr := errors.Unwrap(err)
	if technicalErr != nil {
		metadata["fetchTechnicalError"] = technicalErr.Error()
	} else {
		delete(metadata, "fetchTechnicalError")
		technicalErr = err
	}
	// 可重试判定要看原始错误：technicalErr 可能只是 JSON 解码的
	// "invalid character '<'"，丢掉外层“上游返回异常”的语义。
	retryable := retryableBatchFetchError(err)
	if !retryable {
		retryable = retryableBatchFetchError(technicalErr)
	}
	metadata["fetchRetryable"] = retryable
	metadata["fetchAttempts"] = attempts
	metadata["fetchFailedAt"] = time.Now().UTC()
	// 失败登记本身也可能因为重试期间 revision 已变化而写不进去；先回读最新
	// revision，写失败时再回读重试一次，避免失败原因静默丢失。
	latest := book
	if batch, readErr := store.GetBatch(ctx, owner, batchID); readErr == nil {
		for _, candidate := range batch.Books {
			if candidate.ID == book.ID {
				latest = candidate
				break
			}
		}
	}
	for attempt := 0; attempt < 2; attempt++ {
		if _, writeErr := store.UpdateBookMetadata(ctx, owner, batchID, book.ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: latest.Revision}); writeErr == nil {
			return
		} else if !errors.Is(writeErr, batchfactoryv11.ErrConflict) {
			return
		}
		batch, readErr := store.GetBatch(ctx, owner, batchID)
		if readErr != nil {
			return
		}
		for _, candidate := range batch.Books {
			if candidate.ID == book.ID {
				latest = candidate
				break
			}
		}
	}
}

type batchFetchUserError struct {
	message string
	cause   error
}

func (e *batchFetchUserError) Error() string { return e.message }
func (e *batchFetchUserError) Unwrap() error { return e.cause }

// batchFetchDisplayError turns the upstream's technical 400 into an action the
// user can actually take. “获取书籍信息失败”既可能是书号/书城不匹配，也可能是
// 上游某个书城渠道临时不可用，所以文案要同时给出“稍后重试”和“换书号/贴正文”。
func batchFetchDisplayError(err error, book batchfactoryv11.Book) error {
	if err == nil || !strings.Contains(err.Error(), "获取书籍信息失败") {
		return err
	}
	bookID := strings.TrimSpace(book.BookID)
	if bookID == "" {
		bookID = strings.TrimSpace(book.ID)
	}
	platformName := firstMetadataString(book.SourceMetadata, "platformName")
	if platformName == "" {
		platformName = strings.TrimSpace(book.Platform)
	}
	return &batchFetchUserError{
		message: fmt.Sprintf("书号 %s 在%s无法读取：可能是书号与书名/书城不匹配、该书已下架，或该平台渠道暂时不可用。可稍后重试；若持续失败请更换书号，或打开“编辑内容”粘贴正文后重试", bookID, platformName),
		cause:   err,
	}
}

func retryableBatchFetchError(err error) bool {
	if err == nil {
		return false
	}
	// 展示层错误（batchFetchUserError）会把上游原因包在 cause 里；只测外层
	// 文案会把可自愈的上游故障误判成永久失败，所以两层都要判断。
	if retryableBatchFetchMessage(err.Error()) {
		return true
	}
	if cause := errors.Unwrap(err); cause != nil {
		return retryableBatchFetchMessage(cause.Error())
	}
	return false
}

func retryableBatchFetchMessage(message string) bool {
	message = strings.ToLower(message)
	return strings.Contains(message, "timed out") ||
		strings.Contains(message, "request failed") ||
		strings.Contains(message, "http status 5") ||
		strings.Contains(message, "returned invalid json") ||
		strings.Contains(message, "不是 json") ||
		// 121 的“获取书籍信息失败”既可能是书号/书城不匹配，也可能是上游某个
		// 书城渠道临时故障（例如番茄免费渠道 PHP 告警）。它只重试一次，代价
		// 很低，但能把渠道恢复后的书自动救回来。
		strings.Contains(message, "获取书籍信息失败")
}

func classifyFetchedBatchBook(ctx context.Context, owner string, bookInfo novelfetchworkshop.BookInfo, sourceText string, metadata map[string]any, classification batchFetchClassification) map[string]any {
	meta := copyAnyMap(metadata)
	info := bookInfo
	info.Text = sourceText
	// Keep the upstream book-info facets in the same metadata record that the
	// batch factory detail view and publication form read. Older records only
	// kept category/genre, which made the detail view look empty even though
	// the upstream response had themes and labels.
	if strings.TrimSpace(info.Themes) != "" {
		meta["themes"] = strings.TrimSpace(info.Themes)
	}
	if strings.TrimSpace(info.Labels) != "" {
		meta["labels"] = strings.TrimSpace(info.Labels)
	}
	if firstMetadataString(meta, "tags") == "" {
		facets := make([]string, 0, 2)
		if value := strings.TrimSpace(info.Themes); value != "" {
			facets = append(facets, value)
		}
		if value := strings.TrimSpace(info.Labels); value != "" {
			facets = append(facets, value)
		}
		if len(facets) > 0 {
			meta["tags"] = strings.Join(facets, "、")
		}
	}
	judge := classification.AIGenderJudge
	if classification.AIGenderJudgeResolver != nil {
		if resolved, err := classification.AIGenderJudgeResolver(ctx, owner); err == nil && resolved != nil {
			judge = resolved
		}
	}
	manualGender := ""
	if strings.EqualFold(strings.TrimSpace(stringValue(meta["genderSource"])), "manual") {
		manualGender = stringValue(meta["gender"])
	}
	classificationResult := novelfetchworkshop.ClassifyBook(ctx, manualGender, info, judge, providerForBatchClassification(ctx, owner, classification))
	decision := classificationResult.Gender
	if decision.Gender != "" {
		meta["gender"] = decision.Gender
		meta["genderSource"] = string(decision.Source)
		meta["genderReason"] = decision.Reason
		meta["genderConfidence"] = decision.Confidence
	}
	style := classificationResult.Style
	meta["styleStatus"] = "processing"
	recordNovelFetchStyleDecision(meta, style, nil)
	return meta
}

func providerForBatchClassification(ctx context.Context, owner string, classification batchFetchClassification) novelfetchworkshop.RewriteProvider {
	provider := classification.RewriteProvider
	if classification.RewriteResolver != nil {
		if resolved, err := classification.RewriteResolver(ctx, owner); err == nil && resolved != nil {
			provider = resolved
		}
	}
	return provider
}

// Fetching正文 and classifying its metadata are separate stages. Keep the
// classification result explicit so a failed AI判断 can be retried without
// fetching the正文 again. Classification always produces a usable pair: a
// missing value is derived from the other one, and an unknown pair falls back
// to a default so production is never blocked by a confirmation step.
func finalizeFetchedBatchClassification(metadata map[string]any) map[string]any {
	meta := copyAnyMap(metadata)
	gender := validBatchGender(firstMetadataString(meta, "gender", "gender_name"))
	style := validBatchStyle(firstMetadataString(meta, "style", "styleName", "style_name"))
	note := firstMetadataString(meta, "styleError")
	needsReview := false
	if gender == "" && style != "" {
		gender = genderForBatchStyle(style)
		meta["genderSource"] = "fallback"
		meta["genderReason"] = "平台与 AI 未给出明确男女频，已按风格采用默认男女频"
		meta["genderConfidence"] = 0.5
		needsReview = true
	}
	if style == "" {
		if gender == "" {
			gender = "女频"
			meta["genderSource"] = "fallback"
			meta["genderReason"] = "未能判断男女频，已采用默认女频"
			meta["genderConfidence"] = 0.5
			needsReview = true
		}
		style = defaultBatchStyleForGender(gender)
		meta["styleSource"] = "fallback"
		meta["styleReason"] = "未能判断风格，已采用默认风格"
		meta["styleConfidence"] = 0.5
		needsReview = true
	}
	if strings.EqualFold(stringValue(meta["styleSource"]), "fallback") || strings.EqualFold(stringValue(meta["genderSource"]), "fallback") {
		needsReview = true
	}
	if note != "" {
		meta["classifyNote"] = note
	}
	meta["gender"] = gender
	meta["style"] = style
	meta["styleStatus"] = "classified"
	delete(meta, "styleError")
	meta["classifyStatus"] = "classified"
	meta["classifyError"] = ""
	meta["classifyRetryable"] = false
	meta["classifyNeedsReview"] = needsReview
	if needsReview {
		meta["classifyReason"] = fmt.Sprintf("已自动采用%s / %s，可在单书配置中修改", gender, style)
	} else {
		meta["classifyReason"] = fmt.Sprintf("已识别为%s / %s", gender, style)
	}
	return meta
}

func validBatchGender(value string) string {
	switch strings.TrimSpace(value) {
	case "男频", "女频":
		return strings.TrimSpace(value)
	}
	return ""
}

func validBatchStyle(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "待判定" || value == "未判定" {
		return ""
	}
	return value
}

func defaultBatchStyleForGender(gender string) string {
	if gender == "男频" {
		return "男频都市"
	}
	return "现代通用"
}

func genderForBatchStyle(style string) string {
	switch style {
	case "男频都市", "玄幻", "历史":
		return "男频"
	default:
		return "女频"
	}
}

func registerV12PublishClassificationRoute(mux *http.ServeMux, store batchfactoryv11.Store, classification batchFetchClassification) {
	if mux == nil || store == nil {
		return
	}
	mux.HandleFunc("POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/classify-publish-metadata", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			Force bool `json:"force"`
		}
		if r.ContentLength != 0 && !decodeJSON(w, r, &input) {
			return
		}
		batch, err := store.GetBatch(r.Context(), owner, strings.TrimSpace(r.PathValue("batchId")))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		var book *batchfactoryv11.Book
		for index := range batch.Books {
			if batch.Books[index].ID == strings.TrimSpace(r.PathValue("bookId")) {
				book = &batch.Books[index]
				break
			}
		}
		if book == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "小说不存在"})
			return
		}
		gender := validBatchGender(firstMetadataString(book.SourceMetadata, "gender", "gender_name"))
		style := validBatchStyle(firstMetadataString(book.SourceMetadata, "style", "styleName", "style_name"))
		if !input.Force && gender != "" && style != "" {
			writeJSON(w, http.StatusOK, map[string]any{"book": *book, "reused": true})
			return
		}
		info := novelfetchworkshop.BookInfo{
			Category: firstMetadataString(book.SourceMetadata, "category"),
			Genre:    firstMetadataString(book.SourceMetadata, "genre"),
			Themes:   firstMetadataString(book.SourceMetadata, "themes", "theme"),
			Labels:   firstMetadataString(book.SourceMetadata, "labels", "tags"),
		}
		metadata := finalizeFetchedBatchClassification(classifyFetchedBatchBook(r.Context(), owner, info, book.SourceText, book.SourceMetadata, classification))
		updated, err := store.UpdateBookMetadata(r.Context(), owner, batch.ID, book.ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: book.Revision})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"book": updated, "reused": false})
	})
}

func registerV12SingleFetchRoute(mux *http.ServeMux, store batchfactoryv11.Store, workshopStore novelfetchworkshop.Store, fetcher *novelfetchworkshop.UpstreamFetcher, giantResolver GiantMaterialResolver, classification batchFetchClassification) {
	if mux == nil || store == nil || fetcher == nil {
		return
	}
	mux.HandleFunc("POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/fetch-original", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		batch, err := store.GetBatch(r.Context(), owner, r.PathValue("batchId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		var book *batchfactoryv11.Book
		for index := range batch.Books {
			if batch.Books[index].ID == r.PathValue("bookId") {
				book = &batch.Books[index]
				break
			}
		}
		if book == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "小说不存在"})
			return
		}
		maxTxt := 4000
		var input struct {
			MaxTxt   int    `json:"maxTxt"`
			Platform string `json:"platform"`
			Force    bool   `json:"force"`
		}
		if r.ContentLength != 0 && !decodeJSON(w, r, &input) {
			return
		}
		if isBatchCustomCopy(*book) {
			original, err := reuseCustomCopyOriginal(r.Context(), store, owner, batch.ID, *book)
			if err != nil {
				writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": "reused", "reused": true, "book": original, "length": len([]rune(original.SourceText)), "sourceMetadata": original.SourceMetadata})
			return
		}
		retryProcessing := strings.EqualFold(stringValue(book.SourceMetadata["processingStatus"]), "failed") && strings.TrimSpace(book.TxtText) != ""
		if !input.Force && !retryProcessing && strings.TrimSpace(book.SourceText) != "" {
			writeJSON(w, http.StatusOK, map[string]any{
				"ok": true, "status": "reused", "reused": true, "book": *book,
				"length": len([]rune(book.SourceText)), "sourceMetadata": book.SourceMetadata,
			})
			return
		}
		if input.MaxTxt != 0 {
			maxTxt = input.MaxTxt
		} else if retryProcessing && targetNumber(book.SourceMetadata["maxTxt"]) > 0 {
			maxTxt = int(targetNumber(book.SourceMetadata["maxTxt"]))
		}
		if maxTxt < 100 || maxTxt > 100000 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "字数参数无效"})
			return
		}
		platform := resolveSingleFetchPlatform(input.Platform, *book)
		if platform == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少平台参数"})
			return
		}
		bookID := strings.TrimSpace(book.BookID)
		if bookID == "" {
			bookID = strings.TrimSpace(book.ID)
		}
		if bookID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少小说编号"})
			return
		}
		attempts := 0
		resolvedBookID := bookID
		fallbackFrom := ""
		var result novelfetchworkshop.FetchAndProcessResult
		if retryProcessing {
			result = batchFetchResultFromStoredRaw(*book, maxTxt)
		} else {
			var fetchErr error
			result, resolvedBookID, fallbackFrom, fetchErr, attempts = fetchBatchSourceWithGiantCandidate(r.Context(), fetcher, giantResolver, *book, bookID, platform, maxTxt)
			if fetchErr != nil {
				displayErr := batchFetchDisplayError(fetchErr, *book)
				persistBatchFetchFailure(r.Context(), store, owner, batch.ID, *book, displayErr, attempts)
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": displayErr.Error()})
				return
			}
			bookID = resolvedBookID
		}
		result, metadata, processErr := prepareBatchFetchedSource(r.Context(), workshopStore, owner, bookID, result, classification, resolveBatchSensitiveConfig(batch.SettingsState.Patch, book.SettingsState.Patch))
		if fallbackFrom != "" {
			metadata["platformBookId"] = resolvedBookID
			metadata["resolvedPlatformBookId"] = resolvedBookID
			metadata["platformCandidateFallbackFrom"] = fallbackFrom
		}
		if !retryProcessing {
			markBatchDirectReadSuccess(metadata)
		}
		productionText := result.ProcessedText
		if processErr != nil {
			productionText = ""
		}
		updated, err := captureBatchFetchedSource(r.Context(), store, owner, batch.ID, book.ID, batchfactoryv11.CaptureBookSourceInput{SourceText: productionText, RawSourceText: result.OriginalRaw, ReplaceSource: true, ReplaceTitle: true, SourceBookID: bookID, SourceTitle: fetchedBookSourceTitle(result, book.Title), SourceMetadata: metadata, ExpectedRevision: book.Revision})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		if processErr != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"ok": false, "status": "processing_error", "error": processErr.Error(), "book": updated,
				"rawData": result.OriginalRaw, "sourceMetadata": metadata,
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": "done", "book": updated, "rawData": result.OriginalRaw, "length": result.ProcessedChars, "originalLength": result.OriginalRawChars, "maxTxt": result.MaxTxt, "platform": parsePlatformNumber(platform), "bookinfo": result.BookInfo, "sourceMetadata": metadata})
	})
}

// fetchBatchSourceWithGiantCandidate handles Qingyu materials that point at
// multiple same-title works. The first work is not always readable by 121.
// We only try same-title, same-platform siblings after the selected ID gets the
// deterministic "book info missing" response, so unrelated books never cross.
func fetchBatchSourceWithGiantCandidate(ctx context.Context, fetcher *novelfetchworkshop.UpstreamFetcher, resolver GiantMaterialResolver, book batchfactoryv11.Book, bookID, platform string, maxTxt int) (novelfetchworkshop.FetchAndProcessResult, string, string, error, int) {
	result, err, attempts := fetchBatchSourceWithRetry(ctx, fetcher, bookID, platform, maxTxt)
	if err == nil {
		return result, bookID, "", nil, attempts
	}
	giantMaterialID := firstMetadataString(book.SourceMetadata, "giantMaterialId", "qingyuMaterialId")
	if resolver == nil || giantMaterialID == "" || !strings.Contains(err.Error(), "获取书籍信息失败") {
		return result, bookID, "", err, attempts
	}
	material, resolveErr := resolver.Resolve(ctx, giantMaterialID)
	if resolveErr != nil {
		return result, bookID, "", err, attempts
	}
	platformCode := firstMetadataString(book.SourceMetadata, "platformCode")
	titleKey := giantCandidateTitleKey(book.Title)
	// Qingyu sometimes groups a numbered sequence such as "5风误归期…" with
	// the readable base work "风误归期…". Try the unnumbered base first, then
	// numbered siblings, avoiding several known-bad paid calls before the useful ID.
	for pass := 0; pass < 2; pass++ {
		for _, candidate := range material.Books {
			candidateID := strings.TrimSpace(candidate.PlatformBookID)
			if candidateID == "" || candidateID == bookID || giantCandidateTitleKey(candidate.Title) != titleKey {
				continue
			}
			if (pass == 0) == giantCandidateHasSequencePrefix(candidate.Title) {
				continue
			}
			if candidate.PlatformID != "" && strings.TrimSpace(candidate.PlatformID) != platform {
				continue
			}
			if platformCode != "" && candidate.PlatformCode != "" && !strings.EqualFold(strings.TrimSpace(candidate.PlatformCode), platformCode) {
				continue
			}
			candidateResult, candidateErr, candidateAttempts := fetchBatchSourceWithRetry(ctx, fetcher, candidateID, platform, maxTxt)
			attempts += candidateAttempts
			if candidateErr == nil {
				return candidateResult, candidateID, bookID, nil, attempts
			}
			if !strings.Contains(candidateErr.Error(), "获取书籍信息失败") {
				return candidateResult, candidateID, bookID, candidateErr, attempts
			}
		}
	}
	return result, bookID, "", err, attempts
}

func giantCandidateHasSequencePrefix(raw string) bool {
	value := strings.TrimSpace(raw)
	return value != "" && value[0] >= '0' && value[0] <= '9'
}

func giantCandidateTitleKey(raw string) string {
	value := strings.TrimSpace(raw)
	for len(value) > 0 && value[0] >= '0' && value[0] <= '9' {
		value = value[1:]
	}
	value = strings.TrimSpace(value)
	var normalized strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			normalized.WriteRune(unicode.ToLower(r))
		}
	}
	return normalized.String()
}

func markBatchDirectReadSuccess(metadata map[string]any) {
	if metadata == nil {
		return
	}
	metadata["originalReadStage"] = "direct"
	metadata["originalReadError"] = ""
	metadata["contentPending"] = false
}

func parsePlatformNumber(value string) any {
	if number, err := strconv.Atoi(value); err == nil {
		return number
	}
	return value
}

// Giant-material placeholder books keep the display name (e.g. "七猫") in
// Platform while the numeric 121 platform ID lives in metadata.platformId.
// The upstream fetch API only accepts the numeric ID, so a non-numeric
// platform falls back to the metadata ID before the request goes out.
func resolveSingleFetchPlatform(inputPlatform string, book batchfactoryv11.Book) string {
	platform := strings.TrimSpace(inputPlatform)
	if platform == "" {
		platform = strings.TrimSpace(book.Platform)
	}
	if numericPlatformID(platform) != "" {
		return platform
	}
	if candidate := firstMetadataString(book.SourceMetadata, "platformId", "platform_id"); numericPlatformID(candidate) != "" {
		return strings.TrimSpace(candidate)
	}
	return normalizeSingleFetchPlatformAlias(platform)
}

func normalizeSingleFetchPlatformAlias(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "yg", "阳光", "点众", "点众付费", "yg-阳光-点众付费":
		return "4"
	case "番茄付费", "番茄付费书城":
		return "2"
	case "常读", "常读付费", "常读付费书城":
		return "2"
	case "番茄免费", "番茄免费书城":
		return "7"
	case "七猫", "七猫付费", "七猫付费书城":
		return "3"
	case "知乎", "知乎付费", "知乎付费书城":
		return "15"
	case "掌阅", "掌阅付费", "掌阅付费书城":
		return "20"
	case "卓越", "卓越付费", "卓越付费书城":
		return "26"
	case "九州", "九州书城":
		return "29"
	case "掌文", "掌文付费", "掌文付费书城":
		return "31"
	case "hy", "黑岩", "黑岩付费", "黑岩付费书城":
		return "1"
	default:
		return strings.TrimSpace(value)
	}
}

func numericPlatformID(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	for _, r := range trimmed {
		if r < '0' || r > '9' {
			return ""
		}
	}
	return trimmed
}
