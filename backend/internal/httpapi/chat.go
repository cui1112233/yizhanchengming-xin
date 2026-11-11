package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt"
)

type goChatRequest struct {
	PromptType   string           `json:"promptType"`
	Mode         string           `json:"mode"`
	Format       string           `json:"format"`
	Duration     string           `json:"duration"`
	NovelText    string           `json:"novelText"`
	Characters   any              `json:"characters"`
	Scenes       any              `json:"scenes"`
	VisualStyle  string           `json:"visualStyle"`
	Protagonists any              `json:"protagonists"`
	Constraints  map[string]any   `json:"constraints"`
	Description  string           `json:"descriptionMode"`
	MustCover    string           `json:"mustCoverDetails"`
	Rhythm       string           `json:"shotRhythmRequirements"`
	Messages     []map[string]any `json:"messages"`
	MaxTokens    int              `json:"max_tokens"`
	Temperature  float64          `json:"temperature"`
	TextModelID  string           `json:"textModelId"`
}

func registerChatRoute(root *http.ServeMux, auth BridgeAuth, provider batchfactoryv11.DirectorProvider, resolveProvider func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error), catalog *prompt.Catalog, personal *prompt.PersonalStore) {
	root.Handle("/api/chat", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		if provider == nil && resolveProvider == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "文本模型服务暂不可用"})
			return
		}
		var body goChatRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数无效"})
			return
		}
		identity, _ := BridgeIdentityFromContext(r.Context())
		// 用户在选择模型下拉里指定了模型（含 channel:/custom: 渠道）时，按该模型构造 provider。
		requestProvider := provider
		if resolver := resolveProvider; resolver != nil && strings.TrimSpace(body.TextModelID) != "" {
			resolved, resolveErr := resolver(r.Context(), identity.Username, strings.TrimSpace(body.TextModelID))
			if resolveErr != nil {
				writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "所选文本模型不可用：" + resolveErr.Error()})
				return
			}
			if resolved == nil {
				writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "所选文本模型不可用"})
				return
			}
			requestProvider = resolved
		}
		if requestProvider == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "文本模型服务暂不可用"})
			return
		}
		system, user, disableJSON, err := buildGoChatMessages(body, identity.Username, catalog, personal)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		maxTokens := body.MaxTokens
		if maxTokens < 1 {
			maxTokens = 4096
		}
		if maxTokens > 32768 {
			maxTokens = 32768
		}
		temperature := body.Temperature
		if temperature < 0 || temperature > 2 {
			temperature = 0.7
		}
		content, err := requestProvider.Complete(r.Context(), batchfactoryv11.TextCompletionRequest{SystemPrompt: system, UserPrompt: user, MaxTokens: maxTokens, Temperature: temperature, DisableJSONResponse: disableJSON})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": "文本模型请求失败", "type": "upstream_error"})
			return
		}
		if body.PromptType == "entity_enrich" {
			var enrichment any
			if err := json.Unmarshal([]byte(stripJSONFence(content)), &enrichment); err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": "实体补全结果不是合法 JSON"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"enrichment": enrichment})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"choices": []any{map[string]any{"message": map[string]string{"role": "assistant", "content": content}}}})
	})))
}

func buildGoChatMessages(body goChatRequest, username string, catalog *prompt.Catalog, personal *prompt.PersonalStore) (string, string, bool, error) {
	if body.PromptType == "" && len(body.Messages) > 0 {
		var parts []string
		for _, message := range body.Messages {
			role, _ := message["role"].(string)
			content, _ := message["content"].(string)
			if strings.TrimSpace(content) != "" {
				parts = append(parts, role+": "+content)
			}
		}
		return "", strings.Join(parts, "\n\n"), true, nil
	}
	if strings.TrimSpace(body.NovelText) == "" {
		return "", "", false, errors.New("需要完整原文")
	}
	duration := body.Duration
	if duration != "15s" {
		duration = "10s"
	}
	format := body.Format
	if format == "" {
		format = "screenplay"
	}
	mode := body.Mode
	if mode == "" {
		mode = "continuous"
	}
	lookup := func(id string) string {
		if catalog == nil {
			return ""
		}
		record, found, _ := catalog.PublishedByID(id)
		if found {
			return record.Body
		}
		return ""
	}
	system := lookup(map[string]string{"continuous": "script-continuous", "hook": "script-hook", "segmented": "script-segmented"}[mode])
	if body.PromptType == "extract" {
		system = lookup("script-extract")
	}
	if body.PromptType == "script" {
		formatID := map[string]string{"screenplay": "script-format-screenplay", "storyboard": "script-format-storyboard", "shortdrama": "script-format-shortdrama", "shotlist": "script-format-shotlist", "q版": "script-format-q版"}[format]
		system = strings.Join([]string{system, lookup(formatID), lookup("script-general")}, "\n\n---\n\n")
	}
	if body.PromptType == "quick_director" {
		system = strings.Join([]string{lookup("script-quick-director-storyboard"), lookup("script-director-storyboard-master")}, "\n\n---\n\n")
	}
	if body.PromptType == "entity_enrich" {
		system = lookup("script-extract") + "\n\n只返回 JSON：fields、evidence、suggestions、uncertainties。"
	}
	if strings.TrimSpace(system) == "" {
		system = "请根据输入完成任务，只输出最终结果。"
	}
	constraintText := goConstraintText(body.Constraints, username, catalog, personal)
	if constraintText != "" {
		system += "\n\n## 当前约束\n" + constraintText
	}
	user := "## 小说原文\n" + truncateText(body.NovelText, 180000)
	if body.Characters != nil {
		encoded, _ := json.Marshal(body.Characters)
		user += "\n\n## 人物信息\n" + string(encoded)
	}
	if body.Scenes != nil {
		encoded, _ := json.Marshal(body.Scenes)
		user += "\n\n## 场景信息\n" + string(encoded)
	}
	if body.VisualStyle != "" {
		user += "\n\n## 统一视觉风格\n" + truncateText(body.VisualStyle, 20000)
	}
	if body.PromptType == "quick_director" {
		user += "\n\n## 必须覆盖\n" + body.MustCover + "\n\n## 节奏要求\n" + body.Rhythm
	}
	user += "\n\n当前单元最大时长：" + duration
	return system, user, body.PromptType == "extract" || body.PromptType == "entity_enrich", nil
}

func goConstraintText(values map[string]any, username string, catalog *prompt.Catalog, personal *prompt.PersonalStore) string {
	var parts []string
	for _, category := range []string{"prefix", "quality", "restriction", "negative"} {
		value, _ := values[category].(map[string]any)
		if enabled, _ := value["enabled"].(bool); !enabled {
			continue
		}
		text := ""
		if source, _ := value["source"].(string); source == "personal" && personal != nil {
			if id, _ := value["personalPromptId"].(string); id != "" {
				if item, found, _ := personal.GetOwned(username, id); found {
					text = item.Body
				}
			}
		} else if source, _ := value["source"].(string); source == "draft" {
			text, _ = value["body"].(string)
		} else if catalog != nil {
			if id, _ := value["presetId"].(string); id != "" {
				if record, found, _ := catalog.PublishedByID(id); found {
					text = record.Body
				}
			}
		}
		if strings.TrimSpace(text) != "" {
			parts = append(parts, category+"："+text)
		}
	}
	return strings.Join(parts, "\n\n")
}

func stripJSONFence(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "```json")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(value, "```")
	return strings.TrimSpace(value)
}
