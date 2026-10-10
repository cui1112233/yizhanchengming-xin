package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"qiantie/backend/internal/auth"
)

var defaultUserConfig = map[string]any{
	"provider":                "openai",
	"baseUrl":                 "https://api.openai.com/v1",
	"model":                   "gpt-4o-mini",
	"modelCatalog":            []any{},
	"modelCatalogVersion":     1,
	"storageRoot":             "",
	"productionRetentionDays": 3,
	"image": map[string]any{
		"mode": "openai_compatible", "provider": "openai_compatible", "displayName": "", "baseUrl": "", "model": "",
	},
	"video": map[string]any{
		"provider": "yd_video", "displayName": "中转亚迪",
	},
	"pet": map[string]any{
		"id": "stacky", "displayName": "CM", "description": "CM，前贴的桌面宠物。", "spriteVersionNumber": 2, "spritesheetPath": "/pets/stacky/spritesheet.webp",
	},
	"tts":           map[string]any{"voice": "zh-CN-XiaoxiaoNeural", "style": "general", "speed": 1.8, "pitch": 10},
	"notifications": map[string]any{"soundEnabled": true, "petVisible": true, "soundVolume": 60},
	"avatar":        nil,
}

type UserConfigStore struct {
	BaseDir string
	mu      sync.Mutex
}

func NewUserConfigStore(baseDir string) *UserConfigStore {
	return &UserConfigStore{BaseDir: strings.TrimSpace(baseDir)}
}

func cloneConfigMap(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		if nested, ok := value.(map[string]any); ok {
			result[key] = cloneConfigMap(nested)
		} else if list, ok := value.([]any); ok {
			result[key] = append([]any(nil), list...)
		} else {
			result[key] = value
		}
	}
	return result
}

func mergeConfigMap(base, patch map[string]any) map[string]any {
	result := cloneConfigMap(base)
	for key, value := range patch {
		if nested, ok := value.(map[string]any); ok {
			previous, _ := result[key].(map[string]any)
			if previous == nil {
				previous = map[string]any{}
			}
			result[key] = mergeConfigMap(previous, nested)
			continue
		}
		result[key] = value
	}
	return result
}

func (s *UserConfigStore) configPath(username string) string {
	return filepath.Join(s.BaseDir, username, "api-config.json")
}

func (s *UserConfigStore) read(username string) map[string]any {
	config := cloneConfigMap(defaultUserConfig)
	if s == nil || s.BaseDir == "" {
		return config
	}
	raw, err := os.ReadFile(s.configPath(username))
	if err != nil {
		return config
	}
	var saved map[string]any
	if json.Unmarshal(raw, &saved) != nil {
		return config
	}
	return mergeConfigMap(config, saved)
}

func (s *UserConfigStore) write(username string, config map[string]any) error {
	if s == nil || s.BaseDir == "" {
		return fmt.Errorf("用户配置存储未配置")
	}
	dir := filepath.Dir(s.configPath(username))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.configPath(username) + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.configPath(username))
}

func publicUserConfig(config map[string]any) map[string]any {
	result := cloneConfigMap(config)
	apiKey, _ := result["apiKey"].(string)
	delete(result, "apiKey")
	result["hasApiKey"] = strings.TrimSpace(apiKey) != ""
	if image, ok := result["image"].(map[string]any); ok {
		key, _ := image["apiKey"].(string)
		delete(image, "apiKey")
		image["hasApiKey"] = strings.TrimSpace(key) != ""
	}
	if video, ok := result["video"].(map[string]any); ok {
		apiKey, _ := video["apiKey"].(string)
		ydKey, _ := video["ydApiKey"].(string)
		h3Key, _ := video["h3ApiKey"].(string)
		delete(video, "apiKey")
		delete(video, "ydApiKey")
		delete(video, "h3ApiKey")
		video["hasApiKey"] = strings.TrimSpace(apiKey) != "" || strings.TrimSpace(ydKey) != "" || strings.TrimSpace(h3Key) != ""
		video["ydHasApiKey"] = strings.TrimSpace(apiKey) != "" || strings.TrimSpace(ydKey) != ""
		video["h3HasApiKey"] = strings.TrimSpace(h3Key) != ""
	}
	if catalog, ok := result["modelCatalog"].([]any); ok {
		for _, raw := range catalog {
			model, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			key, _ := model["credential"].(string)
			delete(model, "credential")
			model["hasApiKey"] = strings.TrimSpace(key) != ""
		}
	}
	return result
}

func preferenceOnlyPatch(input map[string]any) map[string]any {
	allowed := map[string]any{}
	for _, key := range []string{"storageRoot", "productionRetentionDays", "pet", "tts", "notifications", "avatar", "scriptDefaults"} {
		if value, ok := input[key]; ok {
			allowed[key] = value
		}
	}
	return allowed
}

func registerUserConfigRoutes(root *http.ServeMux, middleware BridgeAuth, store *UserConfigStore, accounts auth.Store) {
	if root == nil || middleware.AuthStore == nil || store == nil || accounts == nil {
		return
	}
	root.Handle("GET /api/config", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		store.mu.Lock()
		config := store.read(identity.Username)
		store.mu.Unlock()
		writeJSON(w, http.StatusOK, publicUserConfig(config))
	})))
	root.Handle("POST /api/config", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var patch map[string]any
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 256<<10)).Decode(&patch) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		if !identity.IsOwner {
			patch = preferenceOnlyPatch(patch)
		}
		store.mu.Lock()
		config := mergeConfigMap(store.read(identity.Username), patch)
		err := store.write(identity.Username, config)
		store.mu.Unlock()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "配置保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, publicUserConfig(config))
	})))
}
