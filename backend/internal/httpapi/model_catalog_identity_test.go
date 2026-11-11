package httpapi

import (
	"context"
	"testing"
)

func TestModelIdentityKeyNormalizesEndpointAndModel(t *testing.T) {
	cases := []struct {
		baseURL string
		modelID string
		want    string
	}{
		{"https://API.Example.com/v1/", " GPT-Image-2.5 ", "https://api.example.com/v1|gpt-image-2.5"},
		{"https://api.example.com/v1", "gpt-image-2.5", "https://api.example.com/v1|gpt-image-2.5"},
		{"", "gpt-image-2", ""},
		{"https://api.example.com/v1", "  ", ""},
	}
	for _, item := range cases {
		if got := modelIdentityKey(item.baseURL, item.modelID); got != item.want {
			t.Fatalf("modelIdentityKey(%q, %q) = %q, want %q", item.baseURL, item.modelID, got, item.want)
		}
	}
}

func TestIntegrationChannelModelsSkipsChannelAlreadyInCatalog(t *testing.T) {
	// 渠道地址多一个末尾斜杠、模型名大小写不同，仍应判定为同一个真实模型。
	userConfig := newIntegrationTestUserConfig(t, map[string]any{
		"values": map[string]any{},
		"customChannels": []any{map[string]any{
			"id": "ch_dup", "name": "重复渠道", "group": "image",
			"baseUrl": "https://GrsAI.ai/", "model": "GPT-Image-2",
			"apiKey": "sk-test", "enabled": true,
		}},
		"channelSelection": map[string]any{"image": map[string]any{"enabled": []any{"ch_dup"}}},
	})
	existing := []map[string]any{{
		"id": "m-gpt-image-2", "kind": "image",
		"baseUrl": "https://grsai.ai", "modelId": "gpt-image-2",
	}}

	got := integrationChannelModels(context.Background(), "image", userConfig, "tester", existing)
	if len(got) != 0 {
		t.Fatalf("同源渠道不应再进下拉，got %d 条：%v", len(got), got)
	}
}

func TestManagedModelAvailableIgnoresManualSwitch(t *testing.T) {
	cases := []struct {
		name  string
		model map[string]any
		want  bool
	}{
		{"填了凭证就可用，不依赖 enabled", map[string]any{"id": "m1", "credential": "sk-test"}, true},
		{"空凭证不可用", map[string]any{"id": "m2", "enabled": true}, false},
		{"enabled=false 的旧数据不出现", map[string]any{"id": "m3", "credential": "sk-test", "enabled": false}, false},
		{"[off] 已下线不出现", map[string]any{"id": "m4", "displayName": "[off]gpt-image-2.5", "credential": "sk-test", "enabled": true}, false},
		{"本地执行器无需凭证", map[string]any{"id": "local-doubao-executor-video", "provider": "local_executor"}, true},
	}
	for _, item := range cases {
		if got := managedModelAvailable(item.model); got != item.want {
			t.Fatalf("%s：managedModelAvailable = %v，want %v", item.name, got, item.want)
		}
	}
}

func TestPublicManagedModelCarriesAvailableFlag(t *testing.T) {
	public := publicManagedModel(map[string]any{"id": "m1", "credential": "sk-test", "enabled": true})
	if public["available"] != true || public["hasCredential"] != true {
		t.Fatalf("publicManagedModel 应带 available/hasCredential，got %v", public)
	}
	if _, leaked := public["credential"]; leaked {
		t.Fatal("publicManagedModel 不能返回明文 credential")
	}
}

func TestIntegrationChannelModelsIgnoresLegacySelection(t *testing.T) {
	// 预置渠道填了 Key、但旧 channelSelection 是空的：按“填了就生效”应进下拉。
	userConfig := newIntegrationTestUserConfig(t, map[string]any{
		"values":           map[string]any{"deepseekOfficialApiKey": "sk-test"},
		"channelSelection": map[string]any{},
	})

	got := integrationChannelModels(context.Background(), "text", userConfig, "tester", nil)
	if len(got) != 1 || got[0]["id"] != "channel:deepseek" {
		t.Fatalf("预置渠道应只按凭证进入下拉，got %v", got)
	}
}

func TestIntegrationChannelModelsRestoresLegacyEmptyGroup(t *testing.T) {
	// 旧版自定义渠道 group 为空，只能从 channelSelection 推断分组；
	// 页面与下拉必须得出同一个分组，否则会出现“页面看不到但下拉能选”。
	userConfig := newIntegrationTestUserConfig(t, map[string]any{
		"values": map[string]any{},
		"customChannels": []any{map[string]any{
			"id": "ch_legacy", "name": "旧自定义文本", "group": "",
			"baseUrl": "https://api.example.com/v1", "model": "gpt-4o", "apiKey": "sk-test",
		}},
		"channelSelection": map[string]any{"text": map[string]any{"enabled": []any{"ch_legacy"}}},
	})

	if got := integrationChannelModels(context.Background(), "text", userConfig, "tester", nil); len(got) != 1 {
		t.Fatalf("空 group 的旧渠道应按 channelSelection 归到 text，got %v", got)
	}
	if got := integrationChannelModels(context.Background(), "image", userConfig, "tester", nil); len(got) != 0 {
		t.Fatalf("同一条渠道不应出现在其他分组，got %v", got)
	}
}

func TestPlatformPresetsRemainAvailableForDropdown(t *testing.T) {
	// 三组页面不再重复展示平台预设，但它们仍必须出现在 /api/models 下拉里。
	for _, id := range []string{"yd2-mini-video", "minimax-h3-video", "seedance-2-0-official", "fanke-video-api"} {
		model := map[string]any{"id": id, "kind": "video", "credential": "sk-test", "enabled": true}
		if !managedModelAvailable(model) {
			t.Fatalf("%s 应保持可用（下拉可选项），不能因为页面去重而丢失", id)
		}
	}
	local := map[string]any{"id": "local-doubao-executor-video", "kind": "video", "provider": "local_executor", "enabled": true}
	if !managedModelAvailable(local) {
		t.Fatal("本地执行器没有 Key 也必须保留下拉可选项")
	}
}

func TestIntegrationChannelModelsKeepsDifferentEndpoint(t *testing.T) {
	userConfig := newIntegrationTestUserConfig(t, map[string]any{
		"values": map[string]any{},
		"customChannels": []any{map[string]any{
			"id": "ch_other", "name": "另一个地址", "group": "image",
			"baseUrl": "https://mirror.example.com", "model": "gpt-image-2",
			"apiKey": "sk-test", "enabled": true,
		}},
		"channelSelection": map[string]any{"image": map[string]any{"enabled": []any{"ch_other"}}},
	})
	existing := []map[string]any{{
		"id": "m-gpt-image-2", "kind": "image",
		"baseUrl": "https://grsai.ai", "modelId": "gpt-image-2",
	}}

	got := integrationChannelModels(context.Background(), "image", userConfig, "tester", existing)
	if len(got) != 1 || got[0]["id"] != "custom:ch_other" {
		t.Fatalf("不同地址的渠道应保留，got %v", got)
	}
}
