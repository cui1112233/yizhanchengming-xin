package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

type integrationChannelRecorder struct {
	mu            sync.Mutex
	path          string
	authorization string
	body          map[string]any
}

func (r *integrationChannelRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	raw, _ := io.ReadAll(req.Body)
	r.mu.Lock()
	r.path = req.URL.Path
	r.authorization = req.Header.Get("Authorization")
	_ = json.Unmarshal(raw, &r.body)
	r.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"pong"}}],"data":[{"url":"https://example.test/generated.png"}]}`))
}

func newIntegrationTestUserConfig(t *testing.T, apiIntegrations map[string]any) *UserConfigStore {
	t.Helper()
	store := NewUserConfigStore(t.TempDir())
	if err := store.write("tester", map[string]any{apiIntegrationsSettingKey: apiIntegrations}); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestIntegrationChannelTextProviderUsesCustomChannel(t *testing.T) {
	recorder := &integrationChannelRecorder{}
	server := httptest.NewServer(recorder)
	defer server.Close()

	userConfig := newIntegrationTestUserConfig(t, map[string]any{
		"values": map[string]any{},
		"customChannels": []any{map[string]any{
			"id": "ch_text", "name": "测试文本", "group": "text",
			"baseUrl": server.URL + "/v1", "model": "demo-model", "apiKey": "sk-test", "enabled": true,
		}},
		"channelSelection": map[string]any{"text": map[string]any{"enabled": []any{"ch_text"}}},
	})

	provider, matched, err := ResolveIntegrationTextProvider(context.Background(), userConfig, "tester", "custom:ch_text")
	if !matched || err != nil {
		t.Fatalf("matched=%v err=%v", matched, err)
	}
	content, err := provider.Complete(context.Background(), batchfactoryv11.TextCompletionRequest{
		SystemPrompt: "sys", UserPrompt: "hi", Temperature: 0, MaxTokens: 16,
	})
	if err != nil || content != "pong" {
		t.Fatalf("content=%q err=%v", content, err)
	}
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if recorder.path != "/v1/chat/completions" {
		t.Fatalf("request path=%q", recorder.path)
	}
	if recorder.authorization != "Bearer sk-test" {
		t.Fatalf("authorization=%q", recorder.authorization)
	}
	if model := recorder.body["model"]; model != "demo-model" {
		t.Fatalf("model=%v", model)
	}
	messages, _ := recorder.body["messages"].([]any)
	if len(messages) != 2 {
		t.Fatalf("messages=%v", recorder.body["messages"])
	}
}

func TestIntegrationChannelPresetResolvesOfficialEndpoint(t *testing.T) {
	userConfig := newIntegrationTestUserConfig(t, map[string]any{
		"values":           map[string]any{"deepseekOfficialApiKey": "sk-bad"},
		"channelSelection": map[string]any{"text": map[string]any{"enabled": []any{"deepseek"}}},
	})
	provider, matched, err := ResolveIntegrationTextProvider(context.Background(), userConfig, "tester", "channel:deepseek")
	if !matched || err != nil {
		t.Fatalf("matched=%v err=%v", matched, err)
	}
	resolved, ok := provider.(*integrationTextProvider)
	if !ok {
		t.Fatalf("unexpected provider %T", provider)
	}
	if resolved.Endpoint != "https://api.deepseek.com/v1/chat/completions" || resolved.Model != "deepseek-chat" || resolved.APIKey != "sk-bad" {
		t.Fatalf("endpoint=%q model=%q key=%q", resolved.Endpoint, resolved.Model, resolved.APIKey)
	}

	withoutKey := newIntegrationTestUserConfig(t, map[string]any{
		"values":           map[string]any{},
		"channelSelection": map[string]any{"text": map[string]any{"enabled": []any{"deepseek"}}},
	})
	_, matched, err = ResolveIntegrationTextProvider(context.Background(), withoutKey, "tester", "channel:deepseek")
	if !matched || err == nil || !strings.Contains(err.Error(), "API Key") {
		t.Fatalf("matched=%v err=%v", matched, err)
	}
}

func TestAgentImageGeneratorRunsCustomChannel(t *testing.T) {
	recorder := &integrationChannelRecorder{}
	server := httptest.NewServer(recorder)
	defer server.Close()

	userConfig := newIntegrationTestUserConfig(t, map[string]any{
		"values": map[string]any{},
		"customChannels": []any{map[string]any{
			"id": "ch_img", "name": "测试出图", "group": "image",
			"baseUrl": server.URL + "/v1", "model": "img-model", "apiKey": "sk-img", "enabled": true,
		}},
		"channelSelection": map[string]any{"image": map[string]any{"enabled": []any{"ch_img"}}},
	})

	generator := newAgentImageGenerator(nil, nil, userConfig, nil, nil, nil, nil)
	images, status, err := generator(context.Background(), "tester", AgentImageGenerationInput{
		ModelID: "custom:ch_img", Prompt: "一只猫", Count: 1, AspectRatio: "1:1",
	})
	if err != nil || status != http.StatusOK || len(images) != 1 {
		t.Fatalf("status=%d images=%v err=%v", status, images, err)
	}
	if images[0] != "https://example.test/generated.png" {
		t.Fatalf("images=%v", images)
	}
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if recorder.path != "/v1/images/generations" {
		t.Fatalf("request path=%q", recorder.path)
	}
	if recorder.authorization != "Bearer sk-img" {
		t.Fatalf("authorization=%q", recorder.authorization)
	}
	if model := recorder.body["model"]; model != "img-model" {
		t.Fatalf("model=%v", model)
	}
	if prompt := recorder.body["prompt"]; prompt != "一只猫" {
		t.Fatalf("prompt=%v", prompt)
	}
}

func TestIntegrationChannelPresetLiveRejectsBadKey(t *testing.T) {
	if os.Getenv("V88_LIVE_CHANNEL_PROBE") != "1" {
		t.Skip("set V88_LIVE_CHANNEL_PROBE=1 to probe the official endpoint with an invalid key")
	}
	userConfig := newIntegrationTestUserConfig(t, map[string]any{
		"values":           map[string]any{"deepseekOfficialApiKey": "sk-v88-invalid-probe"},
		"channelSelection": map[string]any{"text": map[string]any{"enabled": []any{"deepseek"}}},
	})
	provider, matched, err := ResolveIntegrationTextProvider(context.Background(), userConfig, "tester", "channel:deepseek")
	if !matched || err != nil {
		t.Fatalf("matched=%v err=%v", matched, err)
	}
	_, err = provider.Complete(context.Background(), batchfactoryv11.TextCompletionRequest{
		SystemPrompt: "ping", UserPrompt: "ping", MaxTokens: 4, DisableJSONResponse: true,
	})
	if err == nil {
		t.Fatal("expected the official endpoint to reject an invalid key")
	}
	if !strings.Contains(err.Error(), "HTTP 401") && !strings.Contains(err.Error(), "HTTP 403") {
		t.Fatalf("expected 401/403 from the official endpoint, got: %v", err)
	}
}

func TestIntegrationChannelImagePresetLiveRejectsBadKey(t *testing.T) {
	if os.Getenv("V88_LIVE_CHANNEL_PROBE") != "1" {
		t.Skip("set V88_LIVE_CHANNEL_PROBE=1 to probe the official endpoint with an invalid key")
	}
	userConfig := newIntegrationTestUserConfig(t, map[string]any{
		"values":           map[string]any{"volcanoSeedreamApi": "sk-v88-invalid-probe"},
		"channelSelection": map[string]any{"image": map[string]any{"enabled": []any{"volcanoSeedream"}}},
	})
	target, matched, err := integrationChannelTargetFor(context.Background(), userConfig, "tester", "channel:volcanoSeedream")
	if !matched || err != nil {
		t.Fatalf("matched=%v err=%v", matched, err)
	}
	_, _, err = generateIntegrationChannelImage(context.Background(), target, AgentImageGenerationInput{
		Prompt: "probe", Count: 1, AspectRatio: "1:1",
	})
	if err == nil {
		t.Fatal("expected the official endpoint to reject an invalid key")
	}
	if !strings.Contains(err.Error(), "HTTP 401") && !strings.Contains(err.Error(), "HTTP 403") {
		t.Fatalf("expected 401/403 from the official endpoint, got: %v", err)
	}
}

func TestAgentChatTextModelsIncludesEnabledChannels(t *testing.T) {
	userConfig := newIntegrationTestUserConfig(t, map[string]any{
		"values": map[string]any{
			"deepseekOfficialApiKey": "sk-demo",
		},
		"customChannels": []any{map[string]any{
			"id": "ch_text", "name": "测试文本", "group": "text",
			"baseUrl": "https://example.test/v1", "model": "demo-model", "apiKey": "sk-test", "enabled": true,
		}},
		"channelSelection": map[string]any{"text": map[string]any{"enabled": []any{"deepseek", "ch_text"}}},
	})
	models := agentChatTextModels(context.Background(), nil, userConfig, "tester")
	byID := map[string]string{}
	for _, model := range models {
		byID[model["id"].(string)] = model["name"].(string)
	}
	if !strings.HasPrefix(byID["channel:deepseek"], "深度探索 DeepSeek") {
		t.Fatalf("preset channel missing: %#v", models)
	}
	if !strings.HasPrefix(byID["custom:ch_text"], "测试文本") {
		t.Fatalf("custom channel missing: %#v", models)
	}
}

// 页面选中 custom: 图片渠道时，必须按 API 接口配置出图，而不是回落到旧 modelCatalog。
func TestResolveShuihuoImageTargetUsesCustomImageChannel(t *testing.T) {
	userConfig := newIntegrationTestUserConfig(t, map[string]any{
		"values": map[string]any{},
		"customChannels": []any{map[string]any{
			"id": "ch_img", "name": "测试图片", "group": "image",
			"baseUrl": "https://img.example.test/v1", "model": "gpt-image-2", "apiKey": "sk-image", "enabled": true,
		}},
	})
	target, err := resolveShuihuoImageTargetWithChannels(context.Background(), userConfig, "tester",
		"custom:ch_img", "custom:ch_img", "1:1", "personal_api", nil)
	if err != nil {
		t.Fatal(err)
	}
	if target.ModelID != "gpt-image-2" {
		t.Fatalf("ModelID=%q, want channel model", target.ModelID)
	}
	if got := fmt.Sprint(target.CatalogModel["baseUrl"]); got != "https://img.example.test/v1" {
		t.Fatalf("baseUrl=%q, want channel baseUrl", got)
	}
	if got := fmt.Sprint(target.CatalogModel["credential"]); got != "sk-image" {
		t.Fatalf("credential=%q, want channel key", got)
	}
	if target.Provider != batchfactoryv11.VideoProviderPersonalAPI {
		t.Fatalf("provider=%q, want personal_api", target.Provider)
	}
}

// 旧作品保存的是 custom-<模型名>，新配置保存的是 custom:<渠道 ID>。
// 组员继承管理者渠道后，旧编号也必须能按真实模型名找到同一条渠道。
func TestResolveShuihuoImageTargetUsesLegacyCustomImageAlias(t *testing.T) {
	userConfig := NewUserConfigStore(t.TempDir())
	if err := userConfig.write("manager", map[string]any{apiIntegrationsSettingKey: map[string]any{
		"values": map[string]any{},
		"customChannels": []any{map[string]any{
			"id": "ch_current_image", "name": "当前图片模型", "group": "image",
			"baseUrl": "https://img.example.test/v1", "model": "gpt-image-2", "apiKey": "sk-image", "enabled": true,
		}},
	}}); err != nil {
		t.Fatal(err)
	}
	target, err := resolveShuihuoImageTargetWithChannels(context.Background(), userConfig, "manager",
		"custom-gpt-image-2", "custom-gpt-image-2", "9:16", "personal_api", nil)
	if err != nil {
		t.Fatal(err)
	}
	if target.ModelID != "gpt-image-2" {
		t.Fatalf("ModelID=%q, want legacy alias to resolve current channel model", target.ModelID)
	}
	if got := fmt.Sprint(target.CatalogModel["baseUrl"]); got != "https://img.example.test/v1" {
		t.Fatalf("baseUrl=%q, want current channel baseUrl", got)
	}
}

// 把文本渠道误选成图片模型时必须明确报错，不能悄悄用旧目录出图。
func TestResolveShuihuoImageTargetRejectsNonImageChannel(t *testing.T) {
	userConfig := newIntegrationTestUserConfig(t, map[string]any{
		"values": map[string]any{},
		"customChannels": []any{map[string]any{
			"id": "ch_text", "name": "测试文本", "group": "text",
			"baseUrl": "https://text.example.test/v1", "model": "demo-model", "apiKey": "sk-text", "enabled": true,
		}},
	})
	_, err := resolveShuihuoImageTargetWithChannels(context.Background(), userConfig, "tester",
		"custom:ch_text", "custom:ch_text", "1:1", "personal_api", nil)
	if err == nil || !strings.Contains(err.Error(), "不是图片渠道") {
		t.Fatalf("err=%v, want non-image channel rejection", err)
	}
}

// 非渠道 id 仍走旧 modelCatalog，保证现有目录模型不受影响。
func TestResolveShuihuoImageTargetKeepsLegacyCatalog(t *testing.T) {
	catalog := []map[string]any{{
		"id": "legacy-image", "kind": "image", "enabled": true,
		"provider": "personal_api", "modelId": "legacy-model",
		"baseUrl": "https://legacy.example.test/v1", "credential": "sk-legacy",
	}}
	target, err := resolveShuihuoImageTargetWithChannels(context.Background(), nil, "tester",
		"legacy-image", "legacy-image", "1:1", "personal_api", catalog)
	if err != nil {
		t.Fatal(err)
	}
	if target.ModelID != "legacy-model" {
		t.Fatalf("ModelID=%q, want legacy model", target.ModelID)
	}
}

func TestShuihuoImageReadinessAcceptsCustomImageChannel(t *testing.T) {
	ctx := context.Background()
	userConfig := newIntegrationTestUserConfig(t, map[string]any{
		"values": map[string]any{},
		"customChannels": []any{
			map[string]any{
				"id": "ch_image", "name": "图片渠道", "group": "image",
				"baseUrl": "https://image.example.test/v1", "model": "gpt-image-2", "apiKey": "sk-image", "enabled": true,
			},
			map[string]any{
				"id": "ch_video", "name": "视频渠道", "group": "video",
				"baseUrl": "https://video.example.test/v1", "model": "video-model", "apiKey": "sk-video", "enabled": true,
			},
		},
		"channelSelection": map[string]any{
			"image": map[string]any{"enabled": []any{"ch_image"}},
			"video": map[string]any{"enabled": []any{"ch_video"}},
		},
	})
	if !shuihuoImageChannelReady(ctx, userConfig, "tester") {
		t.Fatal("configured custom image channel must make image readiness pass")
	}

	missingKey := newIntegrationTestUserConfig(t, map[string]any{
		"values": map[string]any{},
		"customChannels": []any{map[string]any{
			"id": "ch_image", "name": "图片渠道", "group": "image",
			"baseUrl": "https://image.example.test/v1", "model": "gpt-image-2", "apiKey": "", "enabled": true,
		}},
		"channelSelection": map[string]any{"image": map[string]any{"enabled": []any{"ch_image"}}},
	})
	if shuihuoImageChannelReady(ctx, missingKey, "tester") {
		t.Fatal("image channel without api key must not pass readiness")
	}

	videoOnly := newIntegrationTestUserConfig(t, map[string]any{
		"values": map[string]any{},
		"customChannels": []any{map[string]any{
			"id": "ch_video", "name": "视频渠道", "group": "video",
			"baseUrl": "https://video.example.test/v1", "model": "video-model", "apiKey": "sk-video", "enabled": true,
		}},
		"channelSelection": map[string]any{"video": map[string]any{"enabled": []any{"ch_video"}}},
	})
	if shuihuoImageChannelReady(ctx, videoOnly, "tester") {
		t.Fatal("video-only channel must not satisfy image readiness")
	}
}
