package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

func TestScreenplayModelConfigAPIAtomicFieldsClearAndAliases(t *testing.T) {
	ctx := context.Background()
	store := screenplay.NewMemoryStore()
	_, _ = store.CreateProject(ctx, screenplay.Project{ID: "config-p", OwnerUsername: "alice"})
	mux := http.NewServeMux()
	now := time.Now()
	RegisterScreenplayCreativeRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, nil)
	path := "/api/screenplays/config-p/creative-config"
	read := signedJSONRequest(t, mux, now, "alice", "GET", path, nil)
	if read.Code != 200 || !strings.Contains(read.Body.String(), prompt.ScreenplayImageCoreDefault) || !strings.Contains(read.Body.String(), `"image_resolution":"2K"`) {
		t.Fatal(read.Code, read.Body.String())
	}
	body := map[string]any{"expected_revision": 0, "aspect_ratio": "21:9", "image_resolution": "4K", "asset_prompt_type": "western", "asset_process_prompt": "只改材质", "voice_design_model": "qwen-tts", "show_original_text": false, "character_image_mode": "prompt:213", "character_design_prompt_id": "213", "scene_image_mode": "none", "scene_design_prompt_id": prompt.CreativeScene289, "image_core_instruction": "自定义前缀"}
	response := signedJSONRequest(t, mux, now, "alice", "PUT", path, body)
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	p, e := store.GetProject(ctx, "alice", "config-p")
	if e != nil || p.Revision != 1 || p.AspectRatio != "21:9" || p.CreativeConfig.ImageResolution != "4K" || p.CreativeConfig.ShowOriginalText || p.CreativeConfig.VoiceDesignModel != "qwen-tts" || p.CreativeConfig.CharacterImageMode != "typedprompt:"+prompt.CreativeCharacter213 || p.CreativeConfig.SceneDesignPromptID != prompt.CreativeScene289 || p.CreativeConfig.SceneImageMode != "none" {
		t.Fatal(p, e)
	}
	if r := signedJSONRequest(t, mux, now, "alice", "PUT", path, body); r.Code != 409 {
		t.Fatal(r.Code, r.Body.String())
	}
	if r := signedJSONRequest(t, mux, now, "bob", "PUT", path, map[string]any{"expected_revision": 1, "image_resolution": "1K"}); r.Code != 404 {
		t.Fatal(r.Code, r.Body.String())
	}
	if r := signedJSONRequest(t, mux, now, "alice", "PUT", path, map[string]any{"expected_revision": 1, "aspect_ratio": "10:7"}); r.Code != 400 {
		t.Fatal(r.Code, r.Body.String())
	}
	response = signedJSONRequest(t, mux, now, "alice", "PUT", path, map[string]any{"expected_revision": 1, "image_core_instruction": "", "asset_process_prompt": ""})
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	read = signedJSONRequest(t, mux, now, "alice", "GET", path, nil)
	if !strings.Contains(read.Body.String(), `"image_core_instruction":""`) || !strings.Contains(read.Body.String(), `"asset_process_prompt":""`) || !strings.Contains(read.Body.String(), `"show_original_text":false`) || strings.Contains(read.Body.String(), prompt.ScreenplayImageCoreDefault) {
		t.Fatal(read.Body.String())
	}
}

func TestScreenplayModelConfigCreateFlowSinglePromptIDsOverrideDefaultModes(t *testing.T) {
	store := screenplay.NewMemoryStore()
	_, _ = store.CreateProject(context.Background(), screenplay.Project{ID: "create-flow", OwnerUsername: "alice"})
	mux := http.NewServeMux()
	now := time.Now()
	RegisterScreenplayCreativeRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, nil)
	response := signedJSONRequest(t, mux, now, "alice", "PUT", "/api/screenplays/create-flow/creative-config", map[string]any{"expected_revision": 0, "character_design_prompt_id": prompt.CreativeCharacter213, "scene_design_prompt_id": prompt.CreativeScene289})
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	p, e := store.GetProject(context.Background(), "alice", "create-flow")
	if e != nil || p.CreativeConfig.CharacterImageMode != "typedprompt:"+prompt.CreativeCharacter213 || p.CreativeConfig.SceneImageMode != "typedprompt:"+prompt.CreativeScene289 {
		t.Fatal(p.CreativeConfig, e)
	}
}

func TestScreenplayModelConfigFinalImageHTTPUsesResolutionAndCore(t *testing.T) {
	s, p := positioningFixture(t)
	p.CreativeConfig.ConfigVersion = 1
	p.CreativeConfig.ImageResolution = "4K"
	p.CreativeConfig.ImageCoreInstruction = "前置核心"
	p.CreativeConfig.AssetProcessPrompt = "资产专用不许普通叠加"
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body map[string]any
		if e := json.NewDecoder(r.Body).Decode(&body); e != nil {
			t.Fatal(e)
		}
		text, _ := body["prompt"].(string)
		if body["size"] != "4096x2304" || body["model"] != "seedream-test" || !strings.HasPrefix(text, "前置核心\n\n") || strings.Contains(text, "资产专用不许普通叠加") {
			t.Errorf("payload=%v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{map[string]any{"b64_json": base64.StdEncoding.EncodeToString(testImageToolPNG(t))}}})
	}))
	defer server.Close()
	catalog := novelfetchworkshop.NewMemoryStore()
	_ = catalog.PutConfig(context.Background(), p.OwnerUsername, map[string]any{"modelCatalog": []any{map[string]any{"id": "edit", "modelId": "seedream-test", "kind": "image", "enabled": true, "endpoint": server.URL + "/v1/images/generations", "credential": "test-only"}}})
	s.Images = func(ctx context.Context, owner string, in AgentImageGenerationInput) ([]string, int, error) {
		if in.Resolution != "4K" || in.AspectRatio != "16:9" {
			t.Fatal(in.Resolution, in.AspectRatio)
		}
		return configuredBatchImages(ctx, nil, catalog, owner, in)
	}
	if _, _, e := s.generatePositioning(context.Background(), p, "one", positioningRequest{ExpectedRevision: rev(0), IdempotencyKey: "http", AllowMissingRefs: true}); e != nil {
		t.Fatal(e)
	}
	if calls != 1 {
		t.Fatal(calls)
	}
}

func TestScreenplayAssetProcessOnlyDedicatedActionUsesPromptAndEditModel(t *testing.T) {
	s, p := positioningFixture(t)
	p.CreativeConfig.ConfigVersion = 1
	p.CreativeConfig.ImageResolution = "1K"
	p.CreativeConfig.ImageCoreInstruction = "核心在前"
	p.CreativeConfig.AssetProcessPrompt = "资产处理独占"
	ctx := context.Background()
	_, e := s.Projects.CreateAsset(ctx, screenplay.Asset{ID: "hero-process", OwnerUsername: p.OwnerUsername, ProjectID: p.ID, Type: screenplay.AssetCharacter, Name: "甲", Description: "真实角色", Prompt: "角色描述"})
	if e != nil {
		t.Fatal(e)
	}
	image, e := s.saveBytes(ctx, p, "image", "ref.png", "image/png", bytes.NewReader(testImageToolPNG(t)), "hero-process")
	if e != nil {
		t.Fatal(e)
	}
	s.ReferencesProvider = agentImageReferenceProvider(func(context.Context, agentworkspace.Asset, time.Duration) (string, error) {
		return "https://assets.example/ref", nil
	})
	calls := 0
	s.Images = func(_ context.Context, _ string, in AgentImageGenerationInput) ([]string, int, error) {
		calls++
		if in.ModelID != "gen" || in.Resolution != "1K" || in.Size != "576x1024" || !strings.HasPrefix(in.Prompt, "资产处理独占") || strings.Contains(in.Prompt, "核心在前") || len(in.ReferenceAssetIDs) != 1 || in.ReferenceAssetIDs[0] != image.AssetID {
			t.Fatalf("request=%+v", in)
		}
		return []string{"data:image/png;base64," + base64.StdEncoding.EncodeToString(testImageToolPNG(t))}, 200, nil
	}
	st, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if _, e = s.generate(ctx, p, screenplayMediaAction{Action: "asset_process", TargetID: "hero-process", ReferenceItemIDs: []string{image.ID}, IdempotencyKey: "process"}, st); e != nil {
		t.Fatal(e)
	}
	normal, e := s.creativeImagePrompt(ctx, p, "hero-process", "")
	if e != nil || strings.Contains(normal, "资产处理独占") || !strings.HasPrefix(normal, "核心在前") {
		t.Fatal(normal, e)
	}
	p.CreativeConfig.AssetProcessPrompt = ""
	if _, e = s.creativeAssetProcessPrompt(ctx, p, "hero-process", ""); e == nil || calls != 1 {
		t.Fatal(e, calls)
	}
}
