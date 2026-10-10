package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"qiantie/backend/internal/novelfetchworkshop"
)

func registerV12FetchOriginalsRoute(mux *http.ServeMux, fetcher *novelfetchworkshop.UpstreamFetcher) {
	mux.HandleFunc("POST /api/batch-factory/v12/fetch-originals", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := bridgeOwner(r); !ok {
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
		if !decodeJSON(w, r, &input) {
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
			result, err := fetcher.FetchAndProcess(r.Context(), bookID, strconv.Itoa(input.Platform), input.MaxTxt)
			if err != nil {
				results = append(results, map[string]any{"bookId": bookID, "platform": input.Platform, "status": "error", "data": nil, "error": err.Error(), "length": 0})
				continue
			}
			metadata := map[string]any{}
			if result.BookInfo.Category != "" {
				metadata["category"] = result.BookInfo.Category
			}
			if result.BookInfo.Genre != "" {
				metadata["genre"] = result.BookInfo.Genre
			}
			decision := novelfetchworkshop.ResolveGender(r.Context(), "", result.BookInfo, nil)
			if decision.Gender == "男频" || decision.Gender == "女频" {
				metadata["gender"] = decision.Gender
				metadata["genderSource"] = string(decision.Source)
			}
			results = append(results, map[string]any{
				"bookId": bookID, "platform": input.Platform, "status": "ok",
				"data": result.ProcessedText, "rawData": result.OriginalRaw,
				"error": nil, "length": result.ProcessedChars, "attempts": 1,
				"bookinfo": result.BookInfo, "sourceMetadata": metadata,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"results": results})
	})
}
