package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"qiantie/backend/internal/novelfetchworkshop"
)

const (
	referenceAssetDefaultTimeout = 400
	referenceAssetMinTimeout     = 30
	referenceAssetMaxTimeout     = 600
)

var referenceAssetImageDefaults = map[string]any{
	"base_url": "", "model": "", "generate_path": "/images/generations", "edit_path": "/images/edits",
	"size": "1536x1024", "aspect_ratio": "16:9", "supports_reference": true, "timeout_seconds": referenceAssetDefaultTimeout,
	"extra_json": map[string]any{},
}

func referenceAssetTimeout(value any) int {
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
	if parsed < referenceAssetMinTimeout || parsed > referenceAssetMaxTimeout {
		if parsed == 0 {
			return referenceAssetDefaultTimeout
		}
		if parsed < referenceAssetMinTimeout {
			return referenceAssetMinTimeout
		}
		return referenceAssetMaxTimeout
	}
	return parsed
}

func referenceAssetImageSettingsPath(store *UserConfigStore, username string) (string, error) {
	if store == nil || strings.TrimSpace(store.BaseDir) == "" {
		return "", fmt.Errorf("用户配置存储未配置")
	}
	return UserScopedPath(store.BaseDir, username, "reference-assets", "image-settings.json")
}

func legacyReferenceAssetImageSettingsPath(store *UserConfigStore, username string) (string, error) {
	if store == nil || strings.TrimSpace(store.BaseDir) == "" {
		return "", fmt.Errorf("用户配置存储未配置")
	}
	return UserScopedPath(store.BaseDir, username, "novel-panel", "image-settings.json")
}

func readReferenceAssetImageSettings(store *UserConfigStore, username string) map[string]any {
	result := cloneConfigMap(referenceAssetImageDefaults)
	if store == nil || strings.TrimSpace(store.BaseDir) == "" {
		return result
	}
	path, err := referenceAssetImageSettingsPath(store, username)
	if err != nil {
		return result
	}
	raw, err := os.ReadFile(path)
	if err != nil && os.IsNotExist(err) {
		legacyPath, legacyErr := legacyReferenceAssetImageSettingsPath(store, username)
		if legacyErr != nil {
			return result
		}
		raw, err = os.ReadFile(legacyPath)
	}
	if err == nil {
		var saved map[string]any
		if json.Unmarshal(raw, &saved) == nil {
			result = mergeConfigMap(result, saved)
		}
	}
	parsed := referenceAssetTimeout(result["timeout_seconds"])
	if parsed > 400 {
		parsed = 400
	}
	result["timeout_seconds"] = parsed
	return result
}

func readReferenceAssetImageSettingsWithCatalog(ctx context.Context, store *UserConfigStore, username string, catalogStore novelfetchworkshop.Store) map[string]any {
	return readReferenceAssetImageSettingsForModel(ctx, store, username, catalogStore, "")
}

// readReferenceAssetImageSettingsForModel resolves the exact image model the
// user selected in the Script page. Without a selection it retains the legacy
// first-available model behavior for older callers.
func readReferenceAssetImageSettingsForModel(ctx context.Context, store *UserConfigStore, username string, catalogStore novelfetchworkshop.Store, selectedModelID string) map[string]any {
	result := readReferenceAssetImageSettings(store, username)
	selectedModelID = strings.TrimSpace(selectedModelID)
	if target, matched, err := integrationChannelTargetFor(ctx, store, username, selectedModelID); matched {
		if err == nil && target.Kind == "image" {
			result["base_url"], result["model"], result["api_key"] = target.BaseURL, target.ModelID, target.Credential
			return result
		}
	}
	if selectedModelID == "" && strings.TrimSpace(fmt.Sprint(result["base_url"])) != "" && strings.TrimSpace(fmt.Sprint(result["model"])) != "" && strings.TrimSpace(fmt.Sprint(result["api_key"])) != "" {
		return result
	}
	if catalogStore == nil {
		return result
	}
	settings, err := catalogStore.GetConfig(ctx, username)
	if err != nil {
		return result
	}
	for _, candidate := range modelCatalog(settings) {
		if selectedModelID != "" && strings.TrimSpace(fmt.Sprint(candidate["id"])) != selectedModelID {
			continue
		}
		if candidate["kind"] != "image" || candidate["enabled"] != true {
			continue
		}
		baseURL := strings.TrimSpace(fmt.Sprint(candidate["baseUrl"]))
		modelID := strings.TrimSpace(fmt.Sprint(candidate["modelId"]))
		credential := strings.TrimSpace(fmt.Sprint(candidate["credential"]))
		if baseURL == "<nil>" {
			baseURL = ""
		}
		if modelID == "<nil>" {
			modelID = ""
		}
		if credential == "<nil>" {
			credential = ""
		}
		if baseURL == "" || modelID == "" || credential == "" {
			continue
		}
		result["base_url"], result["model"], result["api_key"] = baseURL, modelID, credential
		if capabilities, ok := candidate["capabilities"].(map[string]any); ok {
			result["supports_reference"] = capabilities["supportsReferenceImages"] == true
		}
		break
	}
	return result
}
