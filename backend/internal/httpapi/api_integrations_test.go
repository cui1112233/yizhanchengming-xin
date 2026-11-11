package httpapi

import "testing"

func TestAPIIntegrationsViewHidesSecrets(t *testing.T) {
	stored := map[string]any{
		"values": map[string]any{
			"ttsServiceUrl": "http://tts.example/v1/audio/speech",
			"ttsApiKey":     "sk-secret-value",
			"ttsSentenceLength": float64(50),
		},
		"pinned": []any{"ttsApiKey", "notARealField"},
		"channelSelection": map[string]any{
			"text":          map[string]any{"enabled": []any{"deepseek", "doubao", "deepseek", ""}},
			"unknown-group": "x",
			"tts":           "  ",
		},
		"customChannels": []any{
			map[string]any{"id": "ch_1", "name": "自建", "baseUrl": "https://api.example.com", "apiKey": "sk-custom", "enabled": true, "group": "text"},
		},
	}
	view := apiIntegrationsView(stored)
	values := view["values"].(map[string]any)
	if _, leaked := values["ttsApiKey"]; leaked {
		t.Fatal("secret value must not be returned to the browser")
	}
	if values["ttsServiceUrl"] != "http://tts.example/v1/audio/speech" {
		t.Fatal("non-secret value should be returned")
	}
	configured := view["configured"].(map[string]bool)
	if !configured["ttsApiKey"] {
		t.Fatal("secret should be marked configured")
	}
	pinned := view["pinned"].([]string)
	if len(pinned) != 1 || pinned[0] != "ttsApiKey" {
		t.Fatalf("pinned should drop unknown keys, got %v", pinned)
	}
	channel := view["customChannels"].([]map[string]any)[0]
	if _, leaked := channel["apiKey"]; leaked {
		t.Fatal("custom channel key must not be returned")
	}
	if channel["hasApiKey"] != true {
		t.Fatal("custom channel should report hasApiKey")
	}
	if channel["group"] != "text" {
		t.Fatal("custom channel group should be returned")
	}
	selection := view["channelSelection"].(map[string]any)
	textSelection, ok := selection["text"].(map[string]any)
	if !ok {
		t.Fatalf("known group selection should be normalized to an object, got %v", selection)
	}
	enabled := textSelection["enabled"].([]string)
	if len(enabled) != 2 || enabled[0] != "deepseek" || enabled[1] != "doubao" {
		t.Fatalf("enabled list should be deduped and ordered, got %v", enabled)
	}
	if _, hasActive := textSelection["active"]; hasActive {
		t.Fatal("channel selection must not carry a default active channel")
	}
	if _, ok := selection["unknown-group"]; ok {
		t.Fatal("unknown group selection must be dropped")
	}
	if _, ok := selection["tts"]; ok {
		t.Fatal("blank channel id must be dropped")
	}
}

func TestAPIIntegrationsChannelSelectionKeepsExplicitEmpty(t *testing.T) {
	selection := apiIntegrationChannelSelection(map[string]any{
		"channelSelection": map[string]any{
			"text":  map[string]any{"enabled": []any{}},
			"video": "",
		},
	})
	textSelection, ok := selection["text"].(map[string]any)
	if !ok {
		t.Fatal("explicit empty selection must be kept so the empty state survives a refresh")
	}
	if enabled := textSelection["enabled"].([]string); len(enabled) != 0 {
		t.Fatalf("explicit empty selection should stay empty, got %v", enabled)
	}
	if _, ok := selection["video"]; ok {
		t.Fatal("a blank legacy string must not create a group entry")
	}
}

func TestSanitizeIntegrationChannelKeepsExistingKey(t *testing.T) {
	previous := map[string]any{"id": "ch_1", "apiKey": "sk-keep"}
	channel, ok := sanitizeIntegrationChannel(map[string]any{
		"id": "ch_1", "name": "自建", "baseUrl": "https://api.example.com", "enabled": false,
	}, previous)
	if !ok {
		t.Fatal("channel should be accepted")
	}
	if channel["apiKey"] != "sk-keep" {
		t.Fatal("omitted apiKey must keep the stored one")
	}
	if channel["enabled"] != false {
		t.Fatal("enabled flag should be preserved")
	}
	if _, ok := sanitizeIntegrationChannel(map[string]any{"name": "  "}, nil); ok {
		t.Fatal("channel without a name must be rejected")
	}
	if _, ok := sanitizeIntegrationChannel(map[string]any{"name": "自建", "concurrency": float64(999)}, nil); ok {
		t.Fatal("concurrency above 100 must be rejected")
	}
}

func TestSanitizeIntegrationChannelClearsKeyOnRequest(t *testing.T) {
	previous := map[string]any{"id": "ch_1", "apiKey": "sk-old"}
	channel, ok := sanitizeIntegrationChannel(map[string]any{
		"id": "ch_1", "name": "自建", "clearApiKey": true,
	}, previous)
	if !ok {
		t.Fatal("channel should be accepted")
	}
	if channel["apiKey"] != "" {
		t.Fatal("clearApiKey must remove the stored key")
	}
	// 只改其他字段时，仍然保留已存 Key，避免误删。
	kept, _ := sanitizeIntegrationChannel(map[string]any{"id": "ch_1", "name": "改名"}, previous)
	if kept["apiKey"] != "sk-old" {
		t.Fatal("editing other fields must keep the stored key")
	}
}

func TestSanitizeIntegrationChannelKeepsApiFormatAndFullURL(t *testing.T) {
	channel, ok := sanitizeIntegrationChannel(map[string]any{
		"name": "自建", "apiFormat": "custom", "fullUrl": true, "baseUrl": "https://api.example.com/chat/completions",
	}, nil)
	if !ok {
		t.Fatal("channel should be accepted")
	}
	if channel["apiFormat"] != "custom" || channel["fullUrl"] != true {
		t.Fatalf("apiFormat/fullUrl should be stored, got %v %v", channel["apiFormat"], channel["fullUrl"])
	}
	// 老数据只提交部分字段时，沿用已存的格式与完整地址开关，不被重置。
	kept, ok := sanitizeIntegrationChannel(map[string]any{"name": "改名"},
		map[string]any{"apiFormat": "custom", "fullUrl": true})
	if !ok || kept["apiFormat"] != "custom" || kept["fullUrl"] != true {
		t.Fatalf("partial update should keep apiFormat/fullUrl, got %v", kept)
	}
	if _, ok := sanitizeIntegrationChannel(map[string]any{"name": "自建", "apiFormat": "unknown"}, nil); ok {
		t.Fatal("unknown apiFormat must be rejected")
	}
}

func TestCoerceIntegrationValueRejectsWrongTypes(t *testing.T) {
	if _, ok := coerceIntegrationValue("ttsSentenceLength", "abc"); ok {
		t.Fatal("non-numeric number field must be rejected")
	}
	if value, ok := coerceIntegrationValue("ttsSentenceLength", "50"); !ok || value.(float64) != 50 {
		t.Fatal("numeric string should be coerced")
	}
	if _, ok := coerceIntegrationValue("jimengBrowserEnabled", "yes"); ok {
		t.Fatal("boolean field must reject strings")
	}
}
