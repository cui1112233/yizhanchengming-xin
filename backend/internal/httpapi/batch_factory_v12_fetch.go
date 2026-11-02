package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

func registerV12FetchOriginalsRoute(mux *http.ServeMux, store batchfactoryv11.Store, workshopStore novelfetchworkshop.Store, fetcher *novelfetchworkshop.UpstreamFetcher, classification batchFetchClassification) {
	mux.HandleFunc("POST /api/batch-factory/v12/fetch-originals", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if fetcher == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "原文抓取服务未配置"})
			return
		}
		var input struct {
			Platform int      `json:"platform"`
			BookIDs  []string `json:"bookIds"`
			MaxTxt   int      `json:"maxTxt"`
		}
		if !decodeFetchOriginalsInput(w, r, &input) {
			return
		}
		if input.Platform <= 0 || input.MaxTxt < 100 || input.MaxTxt > 100000 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "平台或字数参数无效"})
			return
		}
		ids := make([]string, 0, len(input.BookIDs))
		seen := map[string]bool{}
		for _, raw := range input.BookIDs {
			id := strings.TrimSpace(raw)
			if id != "" && !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 || len(ids) > 50 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请提供 1–50 个书籍 ID"})
			return
		}
		results := make([]map[string]any, 0, len(ids))
		for _, bookID := range ids {
			result, err, attempts := fetchBatchSourceWithRetry(r.Context(), fetcher, bookID, strconv.Itoa(input.Platform), input.MaxTxt)
			if err != nil {
				results = append(results, map[string]any{"bookId": bookID, "platform": input.Platform, "status": "error", "data": nil, "error": err.Error(), "length": 0, "attempts": attempts, "retryable": retryableBatchFetchError(err)})
				continue
			}
			result, metadata, processErr := prepareBatchFetchedSource(r.Context(), workshopStore, owner, bookID, result, classification)
			if processErr != nil {
				results = append(results, map[string]any{
					"bookId": bookID, "platform": input.Platform, "status": "processing_error",
					"data": nil, "rawData": result.OriginalRaw, "error": processErr.Error(),
					"length": 0, "attempts": attempts, "sourceMetadata": metadata,
				})
				continue
			}
			results = append(results, map[string]any{
				"bookId": bookID, "platform": input.Platform, "status": "ok",
				"data": result.ProcessedText, "rawData": result.OriginalRaw,
				"error": nil, "length": result.ProcessedChars, "attempts": attempts,
				"bookinfo": result.BookInfo, "sourceMetadata": metadata,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"results": results})
	})
}

// decodeFetchOriginalsInput accepts the normal JSON request and a narrow URL
// fallback. Some already-open embedded browser tabs can retain an older fetch
// wrapper that loses a POST body after a hot update; the endpoint still has
// enough request data in the query to complete this idempotent read operation.
func decodeFetchOriginalsInput(w http.ResponseWriter, r *http.Request, target *struct {
	Platform int      `json:"platform"`
	BookIDs  []string `json:"bookIds"`
	MaxTxt   int      `json:"maxTxt"`
}) bool {
	if r.Body != nil {
		raw, err := io.ReadAll(io.LimitReader(r.Body, 256<<10))
		if err == nil && len(strings.TrimSpace(string(raw))) > 0 && json.Unmarshal(raw, target) == nil {
			return true
		}
	}
	query := r.URL.Query()
	platform, platformErr := strconv.Atoi(strings.TrimSpace(query.Get("platform")))
	maxTxt, maxTxtErr := strconv.Atoi(strings.TrimSpace(query.Get("maxTxt")))
	bookIDs := append([]string(nil), query["bookId"]...)
	for _, value := range query["bookIds"] {
		bookIDs = append(bookIDs, strings.Split(value, ",")...)
	}
	if platformErr == nil && maxTxtErr == nil && platform > 0 && maxTxt > 0 && len(bookIDs) > 0 {
		target.Platform = platform
		target.MaxTxt = maxTxt
		target.BookIDs = bookIDs
		return true
	}
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
	return false
}
