package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt"
)

func novelPanelPreset(catalog *prompt.Catalog, id string) string {
	if catalog != nil {
		if record, found, _ := catalog.PublishedByID(id); found {
			return record.Body
		}
	}
	return "请分析用户提供的小说原文，只返回合法 JSON，不要输出 Markdown。"
}

func RegisterNovelPanelAnalysisCompat(root *http.ServeMux, middleware BridgeAuth, provider batchfactoryv11.DirectorProvider, catalog *prompt.Catalog) {
	if root == nil || middleware.AuthStore == nil || provider == nil {
		return
	}
	root.Handle("POST /api/novel-panel/analyze", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := BridgeIdentityFromContext(req.Context()); !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input map[string]any
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<20)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		novelText := strings.TrimSpace(toString(input["novel_text"]))
		if novelText == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先粘贴小说原文。"})
			return
		}
		if len([]rune(novelText)) > 500000 {
			novelText = string([]rune(novelText)[:500000])
		}
		completion, err := provider.Complete(req.Context(), batchfactoryv11.TextCompletionRequest{SystemPrompt: novelPanelPreset(catalog, "novel-analysis"), UserPrompt: novelText, MaxTokens: 12000, Temperature: 0.25})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "模型服务请求失败", "type": "upstream_error"})
			return
		}
		var result map[string]any
		if json.Unmarshal([]byte(stripJSONFence(completion)), &result) != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "模型分析结果格式无效"})
			return
		}
		if _, ok := result["characters"].([]any); !ok {
			result["characters"] = []any{}
		}
		if _, ok := result["scene_options"].([]any); !ok {
			result["scene_options"] = []any{}
		}
		writeJSON(w, http.StatusOK, result)
	})))
}
