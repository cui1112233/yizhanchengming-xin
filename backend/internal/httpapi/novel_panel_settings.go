package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	panelDefaultTimeout = 120
	panelMinTimeout     = 10
	panelMaxTimeout     = 900
)

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

func RegisterNovelPanelSettingsCompat(root *http.ServeMux, middleware BridgeAuth, store *UserConfigStore) {
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
		if req.Method == http.MethodGet && req.URL.Path == "/api/settings" {
			writeJSON(w, http.StatusOK, panelPublicSettings(store, identity.Username))
			return
		}
		if req.Method == http.MethodGet && req.URL.Path == "/api/runtime-config" {
			saved := readPanelSettings(store, identity.Username)
			writeJSON(w, http.StatusOK, map[string]any{"ai_timeout_seconds": saved["ai_timeout_seconds"]})
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
		writeJSON(w, http.StatusOK, map[string]any{"message": "设置已保存。", "settings": panelPublicSettings(store, identity.Username)})
	}))
	root.Handle("GET /api/settings", handler)
	root.Handle("POST /api/settings", handler)
	root.Handle("GET /api/runtime-config", handler)
	root.Handle("POST /api/runtime-config", handler)
}
