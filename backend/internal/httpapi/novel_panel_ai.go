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

func RegisterNovelPanelTextOpsCompat(root *http.ServeMux, middleware BridgeAuth, provider batchfactoryv11.DirectorProvider) {
	if root == nil || middleware.AuthStore == nil || provider == nil {
		return
	}
	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := BridgeIdentityFromContext(req.Context()); !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input map[string]any
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<20)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		path := req.URL.Path
		system, user := "", ""
		maxTokens := 5000
		switch {
		case strings.HasSuffix(path, "/optimize-character"), strings.HasSuffix(path, "/optimize-character-copy"):
			character, ok := input["character"].(map[string]any)
			if !ok {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少当前人物卡。"})
				return
			}
			system = "只返回合法 JSON，返回 {\"appearance\":\"完整可拍摄的外形描述\"}。不得编造输入之外的事实。"
			userBytes, _ := json.Marshal(map[string]any{"novel_text": input["novel_text"], "genre": input["genre"], "trailer_style": input["trailer_style"], "character": character})
			user = string(userBytes)
			maxTokens = 3000
		case strings.HasSuffix(path, "/optimize-all-characters"):
			characters, ok := input["characters"].([]any)
			if !ok || len(characters) == 0 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有可生成的人物卡。"})
				return
			}
			system = "只返回合法 JSON，返回 {\"characters\":[...]}，数组数量、顺序与输入完全一致；每项保留原 name 并补充 appearance。"
			userBytes, _ := json.Marshal(input)
			user = string(userBytes)
			maxTokens = 9000
		case strings.HasSuffix(path, "/optimize-style-copy"):
			system = "只返回合法 JSON，返回 {\"copy\":\"可直接使用的中文风格文案\"}，不得编造小说事实。"
			userBytes, _ := json.Marshal(input)
			user = string(userBytes)
			maxTokens = 3000
		case strings.HasSuffix(path, "/instruction-assist"):
			system = "你是 AI 指令编辑助手。只返回合法 JSON，返回 {\"suggestions\":\"...\",\"revised_instruction\":\"...\"}，不得改变 protocol_lock 约束。"
			userBytes, _ := json.Marshal(input)
			user = string(userBytes)
			maxTokens = 5000
		default:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "接口不存在"})
			return
		}
		completion, err := provider.Complete(req.Context(), batchfactoryv11.TextCompletionRequest{SystemPrompt: system, UserPrompt: user, MaxTokens: maxTokens, Temperature: 0.25})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "模型服务请求失败", "type": "upstream_error"})
			return
		}
		var result map[string]any
		if json.Unmarshal([]byte(stripJSONFence(completion)), &result) != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "模型返回结果格式无效"})
			return
		}
		if strings.HasSuffix(path, "/optimize-character") || strings.HasSuffix(path, "/optimize-character-copy") {
			if strings.TrimSpace(toString(result["appearance"])) == "" {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": "模型没有返回人物外形"})
				return
			}
		}
		if strings.HasSuffix(path, "/optimize-all-characters") {
			if _, ok := result["characters"].([]any); !ok {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": "模型返回的人物卡格式无效"})
				return
			}
		}
		writeJSON(w, http.StatusOK, result)
	}))
	for _, path := range []string{"/api/novel-panel/optimize-character", "/api/novel-panel/optimize-character-copy", "/api/novel-panel/optimize-all-characters", "/api/novel-panel/optimize-style-copy", "/api/novel-panel/instruction-assist"} {
		root.Handle("POST "+path, handler)
	}
}

func RegisterNovelPanelOutlineCompat(root *http.ServeMux, middleware BridgeAuth, provider batchfactoryv11.DirectorProvider, catalog *prompt.Catalog) {
	if root == nil || middleware.AuthStore == nil || provider == nil { return }
	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := BridgeIdentityFromContext(req.Context()); !ok { writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"}); return }
		var input map[string]any; if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<20)).Decode(&input) != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"}); return }
		novelText := strings.TrimSpace(toString(input["novel_text"])); if novelText == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先粘贴小说原文。"}); return }; if len([]rune(novelText)) > 500000 { novelText = string([]rune(novelText)[:500000]) }
		mode := "outline"; if strings.HasSuffix(req.URL.Path, "/regenerate-scene-outline") { mode = "regenerate" }; system := novelPanelPreset(catalog, "novel-outline") + "\n每个非空原文行至少生成一条分镜，返回合法 JSON，顶层必须包含 outline_shots 数组。"; if mode == "regenerate" { system += "这是单场景重生成，只返回当前场景对应分镜。" }
		payload := map[string]any{"novel_text": novelText, "generation_rules": input["generation_rules"], "must_cover_details": input["must_cover_details"], "shot_rhythm_requirements": input["shot_rhythm_requirements"], "scene": input["scene"], "characters": input["characters"], "outline_shots": input["outline_shots"]}; user, _ := json.Marshal(payload); completion, err := provider.Complete(req.Context(), batchfactoryv11.TextCompletionRequest{SystemPrompt: system, UserPrompt: string(user), MaxTokens: 16000, Temperature: 0.35}); if err != nil { writeJSON(w, http.StatusBadGateway, map[string]string{"error": "模型服务请求失败", "type": "upstream_error"}); return }; var result map[string]any; if json.Unmarshal([]byte(stripJSONFence(completion)), &result) != nil { writeJSON(w, http.StatusBadGateway, map[string]string{"error": "模型没有返回合法分镜 JSON"}); return }; if _, ok := result["outline_shots"].([]any); !ok { writeJSON(w, http.StatusBadGateway, map[string]string{"error": "模型没有返回分镜数组"}); return }; result["applied"] = true; if result["report"] == nil { result["report"] = map[string]any{"ok": true} }; writeJSON(w, http.StatusOK, result)
	}))
	for _, path := range []string{"/api/novel-panel/outline-scenes", "/api/novel-panel/regenerate-scene-outline", "/api/novel-panel/generate-scene-prompts"} { root.Handle("POST "+path, handler) }
}
