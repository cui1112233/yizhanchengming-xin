package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

// registerV12MetadataCompat keeps the two metadata repair calls used by the
// novel-fetch handoff on Go. Classification uses the same upstream-first,
// AI-fallback chain as novel fetch, so a manually supplied正文 is classified
// before batch production starts.
func registerV12MetadataCompat(mux *http.ServeMux, auth BridgeAuth, store batchfactoryv11.Store, fetcher *novelfetchworkshop.UpstreamFetcher, classification batchFetchClassification) {
	if mux == nil || store == nil {
		return
	}
	classify := func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		batchID := strings.TrimSpace(req.PathValue("batchId"))
		batch, err := store.GetBatch(req.Context(), identity.Username, batchID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "批量工程不存在"})
			return
		}
		results := make([]map[string]any, 0, len(batch.Books))
		for _, book := range batch.Books {
			info := novelfetchworkshop.BookInfo{
				WorkTitle: book.Title,
				Category:  firstMetadataString(book.SourceMetadata, "category"),
				Genre:     firstMetadataString(book.SourceMetadata, "genre"),
				Themes:    firstMetadataString(book.SourceMetadata, "themes", "theme"),
				Labels:    firstMetadataString(book.SourceMetadata, "labels", "tags"),
			}
			// 与单书分类入口一致：客户端断开（499）不得取消已开始的分类，
			// 摘除取消信号保证结果总能落库。
			runCtx, runCancel := context.WithTimeout(context.WithoutCancel(req.Context()), 4*time.Minute)
			meta := finalizeFetchedBatchClassification(classifyFetchedBatchBook(runCtx, identity.Username, info, book.SourceText, book.SourceMetadata, classification))
			decision := novelfetchworkshop.GenderDecision{
				Gender: stringValue(meta["gender"]), Source: novelfetchworkshop.GenderSource(stringValue(meta["genderSource"])), Reason: stringValue(meta["genderReason"]),
			}
			style := firstMetadataString(meta, "style", "styleName", "style_name")
			updated, updateErr := classifyBookMetadataWithRetry(runCtx, store, identity.Username, batchID, book.ID, meta)
			runCancel()
			status := "classified"
			if updateErr != nil || stringValue(meta["classifyStatus"]) != "classified" {
				status = "failed"
			}
			row := map[string]any{"bookId": book.ID, "status": status, "classification": map[string]any{"gender": decision.Gender, "genderSource": decision.Source, "genderReason": decision.Reason, "style": style}}
			if updateErr != nil {
				row["error"] = updateErr.Error()
			} else if status == "failed" {
				row["error"] = stringValue(meta["classifyError"])
			} else {
				_ = updated
			}
			results = append(results, row)
		}
		writeJSON(w, http.StatusOK, map[string]any{"results": results})
	}

	backfill := func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if fetcher == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "原文获取服务未配置"})
			return
		}
		batchID := strings.TrimSpace(req.PathValue("batchId"))
		batch, err := store.GetBatch(req.Context(), identity.Username, batchID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "批量工程不存在"})
			return
		}
		results := make([]map[string]any, 0, len(batch.Books))
		for _, book := range batch.Books {
			row := map[string]any{"bookId": book.ID}
			if isBatchCustomCopy(book) {
				_, err := reuseCustomCopyOriginal(req.Context(), store, identity.Username, batchID, book)
				row["status"] = "existing"
				if err != nil {
					row["status"] = "failed"
					row["error"] = err.Error()
				}
				results = append(results, row)
				continue
			}
			platform := firstMetadataString(book.SourceMetadata, "platformId", "platform_id")
			if platform == "" {
				platform = book.Platform
			}
			if strings.TrimSpace(book.SourceText) != "" {
				row["status"] = "existing"
				results = append(results, row)
				continue
			}
			fetched, fetchErr := fetcher.Fetch(req.Context(), firstNonEmpty(book.BookID, book.ID), platform, 4000)
			if fetchErr != nil {
				row["status"] = "failed"
				row["error"] = fetchErr.Error()
				results = append(results, row)
				continue
			}
			meta := copyAnyMap(book.SourceMetadata)
			if fetched.BookInfo.Category != "" {
				meta["category"] = fetched.BookInfo.Category
			}
			if fetched.BookInfo.Genre != "" {
				meta["genre"] = fetched.BookInfo.Genre
			}
			processed := novelfetchworkshop.ProcessRawText(fetched.OriginalRaw, 4000)
			meta = finalizeFetchedBatchClassification(classifyFetchedBatchBook(req.Context(), identity.Username, fetched.BookInfo, processed.ProcessedText, meta, classification))
			_, captureErr := store.CaptureBookSource(req.Context(), identity.Username, batchID, book.ID, batchfactoryv11.CaptureBookSourceInput{SourceText: processed.ProcessedText, RawSourceText: fetched.OriginalRaw, SourceMetadata: meta, ExpectedRevision: book.Revision})
			if captureErr != nil {
				row["status"] = "failed"
				row["error"] = captureErr.Error()
			} else {
				row["status"] = "backfilled"
			}
			results = append(results, row)
		}
		writeJSON(w, http.StatusOK, map[string]any{"results": results})
	}

	mux.Handle("POST /api/batch-factory/v12/batches/{batchId}/classify-fetched-metadata", auth.Middleware(http.HandlerFunc(classify)))
	mux.Handle("POST /api/batch-factory/v12/batches/{batchId}/backfill-121-metadata", auth.Middleware(http.HandlerFunc(backfill)))
}

func copyAnyMap(input map[string]any) map[string]any {
	out := make(map[string]any, len(input)+4)
	for key, value := range input {
		out[key] = value
	}
	return out
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "<nil>" {
		return ""
	}
	return text
}

func firstMetadataString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
				return strings.TrimSpace(text)
			}
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
