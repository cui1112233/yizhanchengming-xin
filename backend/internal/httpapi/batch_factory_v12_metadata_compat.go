package httpapi

import (
	"fmt"
	"net/http"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

// registerV12MetadataCompat keeps the two metadata repair calls used by the
// novel-fetch handoff on Go. They are deliberately deterministic: category
// facts returned by 121 are applied first, and an unresolved gender remains
// 待确认 instead of being guessed in the batch factory.
func registerV12MetadataCompat(mux *http.ServeMux, auth BridgeAuth, store batchfactoryv11.Store, fetcher *novelfetchworkshop.UpstreamFetcher) {
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
			meta := copyAnyMap(book.SourceMetadata)
			decision := novelfetchworkshop.ResolveGender(req.Context(), stringValue(meta["gender"]), novelfetchworkshop.BookInfo{
				Category: stringValue(meta["category"]), Genre: stringValue(meta["genre"]),
			}, nil)
			if decision.Gender != "" {
				meta["gender"] = decision.Gender
				meta["genderSource"] = string(decision.Source)
				meta["genderReason"] = decision.Reason
			}
			style := firstMetadataString(meta, "style", "genre", "category")
			if style != "" {
				meta["style"] = style
			}
			updated, updateErr := store.UpdateBookMetadata(req.Context(), identity.Username, batchID, book.ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: meta, ExpectedRevision: book.Revision})
			status := "classified"
			if updateErr != nil {
				status = "failed"
			}
			row := map[string]any{"bookId": book.ID, "status": status, "classification": map[string]any{"gender": decision.Gender, "genderSource": decision.Source, "genderReason": decision.Reason, "style": style}}
			if updateErr != nil {
				row["error"] = updateErr.Error()
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
			decision := novelfetchworkshop.ResolveGender(req.Context(), stringValue(meta["gender"]), fetched.BookInfo, nil)
			if decision.Gender != "" {
				meta["gender"], meta["genderSource"], meta["genderReason"] = decision.Gender, string(decision.Source), decision.Reason
			}
			processed := novelfetchworkshop.ProcessRawText(fetched.OriginalRaw, 4000)
			_, captureErr := store.CaptureBookSource(req.Context(), identity.Username, batchID, book.ID, batchfactoryv11.CaptureBookSourceInput{SourceText: processed.ProcessedText, SourceMetadata: meta, ExpectedRevision: book.Revision})
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
