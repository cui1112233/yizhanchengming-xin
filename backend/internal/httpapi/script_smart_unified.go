package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
)

const smartUnifiedFallbackSystemPrompt = `你负责建立全片级视觉基线。只返回 JSON：{"final_genre":"XXXX短剧","trailer_style":"完整统一摄影规格","story_era":"时代/世界层","negative_prompt":"","picture_limit_prompt":"","quality_constraint_prompt":""}。不要输出解释、Markdown 或人物分镜。`

func registerScriptSmartUnifiedRoute(root *http.ServeMux, auth BridgeAuth, provider batchfactoryv11.DirectorProvider) {
	root.Handle("/api/script/smart-unified-style", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var input struct {
			NovelText  string `json:"novelText"`
			Characters any    `json:"characters"`
			Scenes     any    `json:"scenes"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&input); err != nil || strings.TrimSpace(input.NovelText) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "智能统一视觉分析需要完整原文。"})
			return
		}
		if provider == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "文本模型服务暂不可用。"})
			return
		}
		systemPrompt := smartUnifiedFallbackSystemPrompt
		characters, _ := json.Marshal(input.Characters)
		scenes, _ := json.Marshal(input.Scenes)
		userPrompt := "## 完整原文\n" + truncateText(input.NovelText, 180000) + "\n\n"
		if string(characters) != "null" && string(characters) != "{}" {
			userPrompt += "## 已确认人物卡\n" + string(characters) + "\n\n"
		}
		if string(scenes) != "null" && string(scenes) != "{}" {
			userPrompt += "## 已确认场景卡\n" + string(scenes) + "\n\n"
		}
		userPrompt += "请根据原文建立全片级视觉基线，只返回系统预设定义的 JSON。"
		content, err := provider.Complete(r.Context(), batchfactoryv11.TextCompletionRequest{SystemPrompt: systemPrompt, UserPrompt: userPrompt, Temperature: 0.25, MaxTokens: 2600})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "统一视觉分析模型请求失败。"})
			return
		}
		style, err := normalizeSmartUnifiedResult(content)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "smartUnifiedStyle": style})
	})))
}

func truncateText(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) > max {
		return value[:max]
	}
	return value
}

func normalizeSmartUnifiedResult(raw string) (map[string]any, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(raw, "```"), "```json"))
	var value map[string]any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, errors.New("统一视觉分析必须返回合法 JSON")
	}
	fields := value
	if nested, ok := value["fields"].(map[string]any); ok {
		fields = nested
	}
	genre := firstString(fields, "final_genre", "genre")
	trailer := firstString(fields, "trailer_style")
	era := firstString(fields, "story_era")
	if genre == "" || trailer == "" || era == "" {
		return nil, fmt.Errorf("统一视觉分析缺少必要字段")
	}
	prompt := genre + "；" + trailer + "；" + era + "。"
	return map[string]any{"fields": fields, "prompt": prompt}, nil
}

func firstString(value map[string]any, keys ...string) string {
	for _, key := range keys {
		if text, ok := value[key].(string); ok && strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text)
		}
	}
	return ""
}
