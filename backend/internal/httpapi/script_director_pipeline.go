package httpapi

import (
	"encoding/json"
	"net/http"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt"
)

// registerScriptDirectorPipelineCompat keeps the old browser contract alive,
// but executes it entirely inside Go. This lets older clients migrate without
// bringing the retired Express server back for one endpoint.
func registerScriptDirectorPipelineCompat(root *http.ServeMux, auth BridgeAuth, provider batchfactoryv11.DirectorProvider, catalog *prompt.Catalog, personal *prompt.PersonalStore) {
	root.Handle("/api/script/director-pipeline", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		if provider == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "文本模型服务暂不可用"})
			return
		}
		var body goChatRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数无效"})
			return
		}
		body.PromptType = "quick_director"
		identity, _ := BridgeIdentityFromContext(r.Context())
		system, user, _, err := buildGoChatMessages(body, identity.Username, catalog, personal)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		maxTokens := body.MaxTokens
		if maxTokens < 1 {
			maxTokens = 12000
		}
		if maxTokens > 32768 {
			maxTokens = 32768
		}
		temperature := body.Temperature
		if temperature < 0 || temperature > 2 {
			temperature = 0.2
		}
		content, err := provider.Complete(r.Context(), batchfactoryv11.TextCompletionRequest{SystemPrompt: system, UserPrompt: user, MaxTokens: maxTokens, Temperature: temperature})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "applied": false, "error": "导演流水线上游模型请求失败", "type": "upstream_error"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "applied": true, "output_text": content})
	})))
}
