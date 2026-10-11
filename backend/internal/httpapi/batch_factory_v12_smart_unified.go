package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
)

// registerV12SmartUnifiedRefresh keeps the optional visual-baseline refresh
// on Go. It uses the same text provider as the Go director and persists the
// normalized style on the book, so a later director/H3 run can reuse it.
func registerV12SmartUnifiedRefresh(mux *http.ServeMux, auth BridgeAuth, store batchfactoryv11.Store, provider batchfactoryv11.DirectorProvider) {
	if mux == nil || store == nil {
		return
	}
	mux.Handle("POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/smart-unified/refresh", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if provider == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"smartUnified": map[string]any{"available": false, "skipped": true, "reason": "文本模型服务未配置"}})
			return
		}
		batch, err := store.GetBatch(r.Context(), owner, r.PathValue("batchId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		var book batchfactoryv11.Book
		found := false
		for _, candidate := range batch.Books {
			if candidate.ID == r.PathValue("bookId") {
				book, found = candidate, true
				break
			}
		}
		if !found {
			writeStoreError(w, batchfactoryv11.ErrNotFound)
			return
		}
		source := strings.TrimSpace(book.SourceText)
		if source == "" {
			source = strings.TrimSpace(book.WorkingFrontContent)
		}
		if source == "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"smartUnified": map[string]any{"available": false, "skipped": true, "reason": "当前书没有可分析的原文"}})
			return
		}
		content, err := provider.Complete(r.Context(), batchfactoryv11.TextCompletionRequest{
			SystemPrompt: smartUnifiedFallbackSystemPrompt,
			UserPrompt:   "## 完整原文\n" + truncateText(source, 180000) + "\n\n请根据原文建立全片级视觉基线，只返回系统预设定义的 JSON。",
			Temperature:  0.25,
			MaxTokens:    2600,
		})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"smartUnified": map[string]any{"available": false, "skipped": false, "reason": "统一视觉分析模型请求失败"}})
			return
		}
		style, err := normalizeSmartUnifiedResult(content)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"smartUnified": map[string]any{"available": false, "skipped": false, "reason": err.Error()}})
			return
		}
		encoded, _ := json.Marshal(style)
		latest, err := store.GetBatch(r.Context(), owner, r.PathValue("batchId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, candidate := range latest.Books {
			if candidate.ID != r.PathValue("bookId") {
				continue
			}
			_, err = store.SaveSettings(r.Context(), owner, batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBook, BatchID: r.PathValue("batchId"), BookID: r.PathValue("bookId")}, batchfactoryv11.SettingsUpdate{
				Patch: batchfactoryv11.SettingsPatch{"h3StyleAnalysis": mustJSONRaw(string(encoded))}, ExpectedRevision: candidate.Revision,
			})
			break
		}
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"smartUnified": map[string]any{"available": true, "skipped": false, "reason": "", "attempts": []any{}, "provider": nil, "style": style}})
	})))
}

func mustJSONRaw(value string) json.RawMessage {
	return json.RawMessage(value)
}
