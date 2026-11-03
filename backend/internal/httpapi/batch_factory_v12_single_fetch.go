package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

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

func prepareBatchFetchedSource(ctx context.Context, workshopStore novelfetchworkshop.Store, owner, bookID string, result novelfetchworkshop.FetchAndProcessResult, classification batchFetchClassification) (novelfetchworkshop.FetchAndProcessResult, map[string]any, error) {
	metadata := map[string]any{
		"originalRawStatus":        "done",
		"originalRawChars":         result.OriginalRawChars,
		"upstreamCode":             result.Code,
		"upstreamMessage":          result.Message,
		"sourceMode":               "novel_fetch",
		"maxTxt":                   result.MaxTxt,
		"processingConfigRevision": "novel-fetch-rules-v1",
	}
	if result.BookInfo.Category != "" {
		metadata["category"] = result.BookInfo.Category
	}
	if result.BookInfo.Genre != "" {
		metadata["genre"] = result.BookInfo.Genre
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
	processed, trace, sensitiveItems, processErr := processWorkshopConfiguredText(
		ctx, workshopStore, owner, bookID, result.ProcessedText, "original",
		novelFetchSensitiveAIEnabled(settings), provider,
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
	metadata["fetchRetryable"] = retryableBatchFetchError(err)
	metadata["fetchAttempts"] = attempts
	metadata["fetchFailedAt"] = time.Now().UTC()
	_, _ = store.UpdateBookMetadata(ctx, owner, batchID, book.ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: book.Revision})
}

func retryableBatchFetchError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "timed out") ||
		strings.Contains(message, "request failed") ||
		strings.Contains(message, "http status 5") ||
		strings.Contains(message, "returned invalid json")
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

func registerV12SingleFetchRoute(mux *http.ServeMux, store batchfactoryv11.Store, workshopStore novelfetchworkshop.Store, fetcher *novelfetchworkshop.UpstreamFetcher, classification batchFetchClassification) {
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
		var result novelfetchworkshop.FetchAndProcessResult
		if retryProcessing {
			result = batchFetchResultFromStoredRaw(*book, maxTxt)
		} else {
			var fetchErr error
			result, fetchErr, attempts = fetchBatchSourceWithRetry(r.Context(), fetcher, bookID, platform, maxTxt)
			if fetchErr != nil {
				persistBatchFetchFailure(r.Context(), store, owner, batch.ID, *book, fetchErr, attempts)
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": fetchErr.Error()})
				return
			}
		}
		result, metadata, processErr := prepareBatchFetchedSource(r.Context(), workshopStore, owner, bookID, result, classification)
		productionText := result.ProcessedText
		if processErr != nil {
			productionText = ""
		}
		updated, err := captureBatchFetchedSource(r.Context(), store, owner, batch.ID, book.ID, batchfactoryv11.CaptureBookSourceInput{SourceText: productionText, RawSourceText: result.OriginalRaw, ReplaceSource: true, SourceTitle: book.Title, SourceMetadata: metadata, ExpectedRevision: book.Revision})
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
	case "黑岩", "黑岩付费", "黑岩付费书城":
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
