package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

const (
	panelDefaultTimeout = 400
	panelMinTimeout     = 30
	panelMaxTimeout     = 600
)

var panelImageDefaults = map[string]any{
	"base_url": "", "model": "", "generate_path": "/images/generations", "edit_path": "/images/edits",
	"size": "1536x1024", "aspect_ratio": "16:9", "supports_reference": true, "timeout_seconds": 400,
	"extra_json": map[string]any{},
}

func panelTimeout(value any) int {
	var parsed int
	switch current := value.(type) {
	case float64:
		parsed = int(current)
	case int:
		parsed = current
	case json.Number:
		parsed, _ = strconv.Atoi(string(current))
	case string:
		_, _ = fmt.Sscan(strings.TrimSpace(current), &parsed)
	}
	if parsed < panelMinTimeout || parsed > panelMaxTimeout {
		if parsed == 0 {
			return panelDefaultTimeout
		}
		if parsed < panelMinTimeout {
			return panelMinTimeout
		}
		return panelMaxTimeout
	}
	return parsed
}

func panelSettingsPath(store *UserConfigStore, username string) string {
	return filepath.Join(store.BaseDir, username, "novel-panel", "settings.json")
}

func readPanelSettings(store *UserConfigStore, username string) map[string]any {
	result := map[string]any{"ai_mode": "remote", "ai_timeout_seconds": panelDefaultTimeout, "text_model_id": ""}
	if store == nil || strings.TrimSpace(store.BaseDir) == "" {
		return result
	}
	raw, err := os.ReadFile(panelSettingsPath(store, username))
	if err != nil {
		return result
	}
	var saved map[string]any
	if json.Unmarshal(raw, &saved) == nil {
		for key, value := range saved {
			result[key] = value
		}
	}
	result["ai_mode"] = "remote"
	if value, ok := result["ai_mode"].(string); ok && value == "local" {
		result["ai_mode"] = "local"
	}
	result["ai_timeout_seconds"] = panelTimeout(result["ai_timeout_seconds"])
	if value, ok := result["text_model_id"].(string); ok {
		result["text_model_id"] = strings.TrimSpace(value)
	} else {
		result["text_model_id"] = ""
	}
	return result
}

func writePanelSettings(store *UserConfigStore, username string, settings map[string]any) error {
	if store == nil || strings.TrimSpace(store.BaseDir) == "" {
		return os.ErrInvalid
	}
	dir := filepath.Dir(panelSettingsPath(store, username))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	tmp := panelSettingsPath(store, username) + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, panelSettingsPath(store, username))
}

func panelPublicSettings(store *UserConfigStore, username string) map[string]any {
	config := store.read(username)
	saved := readPanelSettings(store, username)
	apiKey, _ := config["apiKey"].(string)
	model, _ := config["model"].(string)
	baseURL, _ := config["baseUrl"].(string)
	return map[string]any{"ai_mode": saved["ai_mode"], "base_url": baseURL, "model": model, "api_key_configured": strings.TrimSpace(apiKey) != "", "text_model_id": saved["text_model_id"], "ai_timeout_seconds": saved["ai_timeout_seconds"]}
}

func panelPublicSettingsWithCatalog(ctx context.Context, store *UserConfigStore, username string, catalogStore novelfetchworkshop.Store) map[string]any {
	result := panelPublicSettings(store, username)
	if catalogStore == nil {
		return result
	}
	settings, err := catalogStore.GetConfig(ctx, username)
	if err != nil {
		return result
	}
	selected := strings.TrimSpace(fmt.Sprint(result["text_model_id"]))
	var fallback map[string]any
	for _, raw := range modelCatalog(settings) {
		if raw["kind"] != "text" || raw["enabled"] != true {
			continue
		}
		if fallback == nil {
			fallback = raw
		}
		if selected != "" && (raw["id"] == selected || raw["modelId"] == selected) {
			fallback = raw
			break
		}
	}
	if fallback != nil {
		if id := strings.TrimSpace(fmt.Sprint(fallback["id"])); id != "" && id != "<nil>" {
			result["text_model_id"] = id
		}
		if baseURL := strings.TrimSpace(fmt.Sprint(fallback["baseUrl"])); baseURL != "" && baseURL != "<nil>" {
			result["base_url"] = baseURL
		}
		if modelID := strings.TrimSpace(fmt.Sprint(fallback["modelId"])); modelID != "" && modelID != "<nil>" {
			result["model"] = modelID
		}
		credential := strings.TrimSpace(fmt.Sprint(fallback["credential"]))
		result["api_key_configured"] = (credential != "" && credential != "<nil>") || fallback["hasCredential"] == true
	}
	return result
}

func imageSettingsPath(store *UserConfigStore, username string) string {
	return filepath.Join(store.BaseDir, username, "novel-panel", "image-settings.json")
}

func readPanelImageSettings(store *UserConfigStore, username string) map[string]any {
	result := cloneConfigMap(panelImageDefaults)
	if store == nil || strings.TrimSpace(store.BaseDir) == "" {
		return result
	}
	raw, err := os.ReadFile(imageSettingsPath(store, username))
	if err != nil {
		return result
	}
	var saved map[string]any
	if json.Unmarshal(raw, &saved) == nil {
		result = mergeConfigMap(result, saved)
	}
	parsed := panelTimeout(result["timeout_seconds"])
	if parsed < 30 {
		parsed = 30
	}
	if parsed > 400 {
		parsed = 400
	}
	result["timeout_seconds"] = parsed
	return result
}

func publicPanelImageSettings(store *UserConfigStore, username string, input map[string]any) map[string]any {
	result := cloneConfigMap(input)
	key, _ := result["api_key"].(string)
	delete(result, "api_key")
	result["api_key_configured"] = strings.TrimSpace(key) != ""
	if !result["api_key_configured"].(bool) {
		existing, _ := readPanelImageSettings(store, username)["api_key"].(string)
		result["api_key_configured"] = strings.TrimSpace(existing) != ""
	}
	return result
}

func writePanelImageSettings(store *UserConfigStore, username string, settings map[string]any) error {
	dir := filepath.Dir(imageSettingsPath(store, username))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	tmp := imageSettingsPath(store, username) + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, imageSettingsPath(store, username))
}

func RegisterNovelPanelSettingsCompat(root *http.ServeMux, middleware BridgeAuth, store *UserConfigStore, provider batchfactoryv11.DirectorProvider, catalogStores ...novelfetchworkshop.Store) {
	if root == nil || middleware.AuthStore == nil || store == nil {
		return
	}
	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		store.mu.Lock()
		defer store.mu.Unlock()
		var catalogStore novelfetchworkshop.Store
		if len(catalogStores) > 0 {
			catalogStore = catalogStores[0]
		}
		if req.Method == http.MethodGet && req.URL.Path == "/api/settings" {
			writeJSON(w, http.StatusOK, panelPublicSettingsWithCatalog(req.Context(), store, identity.Username, catalogStore))
			return
		}
		if req.Method == http.MethodGet && req.URL.Path == "/api/runtime-config" {
			saved := readPanelSettings(store, identity.Username)
			writeJSON(w, http.StatusOK, map[string]any{"ai_timeout_seconds": saved["ai_timeout_seconds"]})
			return
		}
		if req.Method == http.MethodGet && req.URL.Path == "/api/image-settings" {
			writeJSON(w, http.StatusOK, map[string]any{"settings": publicPanelImageSettings(store, identity.Username, readPanelImageSettings(store, identity.Username))})
			return
		}
		var patch map[string]any
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&patch) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		saved := readPanelSettings(store, identity.Username)
		if req.URL.Path == "/api/runtime-config" {
			if value, exists := patch["ai_timeout_seconds"]; exists {
				saved["ai_timeout_seconds"] = panelTimeout(value)
			}
			if err := writePanelSettings(store, identity.Username, saved); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "设置保存失败"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"message": "AI等待上限已保存。", "runtime_config": map[string]any{"ai_timeout_seconds": saved["ai_timeout_seconds"]}})
			return
		}
		if req.URL.Path == "/api/image-settings" {
			current := readPanelImageSettings(store, identity.Username)
			for _, key := range []string{"base_url", "model", "generate_path", "edit_path", "size", "aspect_ratio"} {
				if value, ok := patch[key]; ok {
					current[key] = strings.TrimSpace(fmt.Sprint(value))
				}
			}
			if value, ok := patch["supports_reference"]; ok {
				current["supports_reference"] = value == true || fmt.Sprint(value) == "true"
			}
			if value, ok := patch["timeout_seconds"]; ok {
				parsed := panelTimeout(value)
				if parsed < 30 {
					parsed = 30
				}
				if parsed > 400 {
					parsed = 400
				}
				current["timeout_seconds"] = parsed
			}
			if value, ok := patch["extra_json"].(map[string]any); ok {
				current["extra_json"] = value
			}
			if clear, ok := patch["clear_api_key"].(bool); ok && clear {
				delete(current, "api_key")
			} else if value, ok := patch["api_key"].(string); ok && strings.TrimSpace(value) != "" {
				current["api_key"] = strings.TrimSpace(value)
			}
			if err := writePanelImageSettings(store, identity.Username, current); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "图片设置保存失败"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "图片AI设置已保存。", "settings": publicPanelImageSettings(store, identity.Username, current)})
			return
		}
		if req.URL.Path == "/api/image-settings/test" {
			candidate := readPanelImageSettings(store, identity.Username)
			for _, key := range []string{"base_url", "model", "api_key", "generate_path", "edit_path"} {
				if value, ok := patch[key]; ok {
					candidate[key] = value
				}
			}
			missing := []string{}
			for _, key := range []string{"base_url", "model", "api_key"} {
				if strings.TrimSpace(fmt.Sprint(candidate[key])) == "" {
					missing = append(missing, key)
				}
			}
			public := publicPanelImageSettings(store, identity.Username, candidate)
			if len(missing) > 0 {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "图片AI配置缺少：" + strings.Join(missing, "、"), "code": "IMAGE_SETTINGS_INCOMPLETE", "details": map[string]any{"missing": missing}, "settings": public})
				return
			}
			parsed, err := url.Parse(strings.TrimRight(strings.TrimSpace(fmt.Sprint(candidate["base_url"])), "/"))
			if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "图片 API 地址格式不正确。请填写 http:// 或 https:// 开头的有效地址。", "code": "IMAGE_AI_URL_INVALID", "settings": public})
				return
			}
			started := time.Now()
			ctx, cancel := context.WithTimeout(req.Context(), 12*time.Second)
			defer cancel()
			probeReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
			probeReq.Header.Set("Accept", "application/json")
			response, probeErr := http.DefaultClient.Do(probeReq)
			diagnostics := map[string]any{"host": parsed.Hostname(), "elapsed_ms": time.Since(started).Milliseconds()}
			if response != nil {
				diagnostics["http_status"] = response.StatusCode
				_ = response.Body.Close()
			}
			if probeErr != nil {
				writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "已保存图片AI设置；无法连通图片服务地址，生成前请确认服务可用。", "diagnostics": diagnostics, "settings": public})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "图片AI服务地址可访问。", "diagnostics": diagnostics, "settings": public})
			return
		}
		if value, ok := patch["ai_mode"].(string); ok && value == "local" {
			saved["ai_mode"] = "local"
		} else if value, ok := patch["ai_mode"].(string); ok && value == "remote" {
			saved["ai_mode"] = "remote"
		}
		if value, ok := patch["text_model_id"].(string); ok {
			saved["text_model_id"] = strings.TrimSpace(value)
		}
		if value, exists := patch["ai_timeout_seconds"]; exists {
			saved["ai_timeout_seconds"] = panelTimeout(value)
		}
		if err := writePanelSettings(store, identity.Username, saved); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "设置保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"message": "设置已保存。", "settings": panelPublicSettingsWithCatalog(req.Context(), store, identity.Username, catalogStore)})
	}))
	root.Handle("GET /api/settings", handler)
	root.Handle("POST /api/settings", handler)
	root.Handle("GET /api/runtime-config", handler)
	root.Handle("POST /api/runtime-config", handler)
	root.Handle("GET /api/image-settings", handler)
	root.Handle("POST /api/image-settings", handler)
	root.Handle("POST /api/image-settings/test", handler)
	root.Handle("POST /api/settings/test", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if provider == nil || (reflect.ValueOf(provider).Kind() == reflect.Ptr && reflect.ValueOf(provider).IsNil()) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "文本 AI 服务暂不可用", "code": "TEXT_AI_UNAVAILABLE"})
			return
		}
		if req.Body != nil {
			_ = json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&map[string]any{})
		}
		ctx, cancel := context.WithTimeout(req.Context(), 20*time.Second)
		defer cancel()
		_, err := provider.Complete(ctx, batchfactoryv11.TextCompletionRequest{SystemPrompt: "你是连接测试服务，只需返回 pong。", UserPrompt: "ping", MaxTokens: 8, Temperature: 0, DisableJSONResponse: true})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error(), "code": "TEXT_AI_TEST_FAILED"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "连接成功。", "username": identity.Username})
	})))
	for _, item := range []struct{ method, pattern, target string }{
		{http.MethodGet, "/api/novel-panel/settings", "/api/settings"},
		{http.MethodPost, "/api/novel-panel/settings", "/api/settings"},
		{http.MethodGet, "/api/novel-panel/runtime-config", "/api/runtime-config"},
		{http.MethodPost, "/api/novel-panel/runtime-config", "/api/runtime-config"},
		{http.MethodGet, "/api/novel-panel/image-settings", "/api/image-settings"},
		{http.MethodPost, "/api/novel-panel/image-settings", "/api/image-settings"},
		{http.MethodPost, "/api/novel-panel/image-settings/test", "/api/image-settings/test"},
		{http.MethodPost, "/api/novel-panel/settings/test", "/api/settings/test"},
	} {
		root.Handle(item.method+" "+item.pattern, panelSettingsAliasHandler(handler))
	}
}

func panelSettingsAliasHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		clone := req.Clone(req.Context())
		urlCopy := *req.URL
		urlCopy.Path = strings.Replace(urlCopy.Path, "/api/novel-panel/", "/api/", 1)
		urlCopy.RawPath = ""
		clone.URL = &urlCopy
		handler.ServeHTTP(w, clone)
	})
}
