package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"qiantie/backend/internal/novelfetchworkshop"
	"strings"
)

func registerNovelFetchSummary(mux *http.ServeMux, store novelfetchworkshop.Store) {
	mux.HandleFunc("GET /api/novel-fetch-workshop/tasks/{bookId}/summary", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID := strings.TrimSpace(req.PathValue("bookId"))
		if !novelFetchBookIDPattern.MatchString(bookID) {
			writeJSON(w, 400, map[string]any{"error": "非法的书籍ID"})
			return
		}
		document, err := novelfetchworkshop.ReadDocumentSummary(req.Context(), store, identity.Username, bookID)
		if errors.Is(err, novelfetchworkshop.ErrNotFound) {
			writeJSON(w, 404, map[string]any{"error": "任务不存在"})
			return
		}
		if err != nil {
			writeJSON(w, 500, map[string]any{"error": "读取任务摘要失败"})
			return
		}
		raw, _ := json.Marshal(document)
		data := map[string]any{}
		json.Unmarshal(raw, &data)
		data["content_loaded"] = false
		writeJSON(w, 200, data)
	})
}
