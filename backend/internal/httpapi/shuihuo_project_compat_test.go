package httpapi

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/prompt"
)

func TestShuihuoProjectsListIncludesOnlyMarkedBatches(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: store})

	ordinary, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "批量工厂项目"})
	if err != nil {
		t.Fatal(err)
	}
	shuihuo, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "创作漫剧项目",
		InitialBatchSettings: batchfactoryv11.SettingsPatch{
			"shuihuoProject": []byte("true"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/shuihuo-production/projects", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("list=%d body=%s", response.Code, response.Body.String())
	}
	body := decodeBody[map[string][]map[string]any](t, response)
	projects := body["projects"]
	if len(projects) != 1 || projects[0]["id"] != shuihuo.ID {
		t.Fatalf("projects=%v, want only marked project %q and no unmarked batch %q", projects, shuihuo.ID, ordinary.ID)
	}
}

func TestShuihuoAssetTypeAndTemplateCanBeUpdated(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: batchfactoryv11.NewMemoryStore()})

	createdType := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/asset-types", map[string]any{"name": "旧名称", "category": "character"})
	if createdType.Code != http.StatusCreated {
		t.Fatalf("create type=%d body=%s", createdType.Code, createdType.Body.String())
	}
	typeBody := decodeBody[map[string]map[string]any](t, createdType)
	typeID, _ := typeBody["assetType"]["id"].(string)
	if typeID == "" {
		t.Fatalf("created type=%v", typeBody)
	}

	updatedType := signedJSONRequest(t, api, now, "alice", http.MethodPut, "/api/shuihuo-production/asset-types/"+typeID, map[string]any{"name": "新名称", "category": "scene"})
	if updatedType.Code != http.StatusOK {
		t.Fatalf("update type=%d body=%s", updatedType.Code, updatedType.Body.String())
	}
	updatedTypeBody := decodeBody[map[string]map[string]any](t, updatedType)
	if updatedTypeBody["assetType"]["name"] != "新名称" || updatedTypeBody["assetType"]["category"] != "scene" {
		t.Fatalf("updated type=%v", updatedTypeBody)
	}

	createdTemplate := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/asset-templates", map[string]any{"assetTypeId": typeID, "name": "旧模板", "prompt": "旧提示"})
	if createdTemplate.Code != http.StatusCreated {
		t.Fatalf("create template=%d body=%s", createdTemplate.Code, createdTemplate.Body.String())
	}
	templateBody := decodeBody[map[string]map[string]any](t, createdTemplate)
	templateID, _ := templateBody["assetTemplate"]["id"].(string)
	if templateID == "" {
		t.Fatalf("created template=%v", templateBody)
	}

	updatedTemplate := signedJSONRequest(t, api, now, "alice", http.MethodPut, "/api/shuihuo-production/asset-templates/"+templateID, map[string]any{"name": "新模板", "prompt": "新提示"})
	if updatedTemplate.Code != http.StatusOK {
		t.Fatalf("update template=%d body=%s", updatedTemplate.Code, updatedTemplate.Body.String())
	}
	updatedTemplateBody := decodeBody[map[string]map[string]any](t, updatedTemplate)
	if updatedTemplateBody["assetTemplate"]["name"] != "新模板" || updatedTemplateBody["assetTemplate"]["prompt"] != "新提示" {
		t.Fatalf("updated template=%v", updatedTemplateBody)
	}
}

func TestShuihuoAssetImageOverrides(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{
		BridgeSecret:   "secret",
		Now:            func() time.Time { return now },
		Slice:          3,
		Store:          batchfactoryv11.NewMemoryStore(),
		LocalArtifacts: localartifact.NewStore(t.TempDir(), 1<<20),
	})

	createdProject := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/projects", map[string]any{"name": "素材覆盖项目"})
	if createdProject.Code != http.StatusCreated {
		t.Fatalf("create project=%d body=%s", createdProject.Code, createdProject.Body.String())
	}
	projectID, _ := decodeBody[map[string]any](t, createdProject)["id"].(string)
	if projectID == "" {
		t.Fatalf("created project=%s", createdProject.Body.String())
	}

	createdAsset := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/projects/"+projectID+"/assets/image", map[string]any{
		"category": "character", "name": "林晚", "prompt": "短发女医生", "imageModelId": "seedream-4.5", "imageAspectRatio": "1:1",
		"dataUrl": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=",
	})
	if createdAsset.Code != http.StatusCreated {
		t.Fatalf("create asset=%d body=%s", createdAsset.Code, createdAsset.Body.String())
	}

	initialRead := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/shuihuo-production/projects/"+projectID, nil)
	if initialRead.Code != http.StatusOK {
		t.Fatalf("initial read=%d body=%s", initialRead.Code, initialRead.Body.String())
	}
	initialAssets, _ := decodeBody[map[string]any](t, initialRead)["assets"].([]any)
	if len(initialAssets) != 1 {
		t.Fatalf("initial assets=%s", initialRead.Body.String())
	}
	initialAsset, _ := initialAssets[0].(map[string]any)
	if initialAsset["imageModelId"] != "seedream-4.5" || initialAsset["imageAspectRatio"] != "1:1" {
		t.Fatalf("initial asset=%v", initialAsset)
	}
	assetID, _ := initialAsset["id"].(string)
	revision, _ := initialAsset["revision"].(float64)
	if assetID == "" || revision == 0 {
		t.Fatalf("initial asset=%v", initialAsset)
	}

	updatedAsset := signedJSONRequest(t, api, now, "alice", http.MethodPut, "/api/shuihuo-production/assets/"+assetID, map[string]any{
		"name": "林晚", "prompt": "短发女医生", "expectedRevision": revision, "imageModelId": "", "imageAspectRatio": "",
	})
	if updatedAsset.Code != http.StatusOK {
		t.Fatalf("update asset=%d body=%s", updatedAsset.Code, updatedAsset.Body.String())
	}
	updated, _ := decodeBody[map[string]any](t, updatedAsset)["asset"].(map[string]any)
	if updated["imageModelId"] != "" || updated["imageAspectRatio"] != "" {
		t.Fatalf("updated asset=%v", updated)
	}

	clearedRead := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/shuihuo-production/projects/"+projectID, nil)
	if clearedRead.Code != http.StatusOK {
		t.Fatalf("cleared read=%d body=%s", clearedRead.Code, clearedRead.Body.String())
	}
	clearedAssets, _ := decodeBody[map[string]any](t, clearedRead)["assets"].([]any)
	if len(clearedAssets) != 1 {
		t.Fatalf("cleared assets=%s", clearedRead.Body.String())
	}
	cleared, _ := clearedAssets[0].(map[string]any)
	if cleared["imageModelId"] != "" || cleared["imageAspectRatio"] != "" {
		t.Fatalf("cleared asset=%v", cleared)
	}
}

func TestShuihuoPresetSlotsExposePublishedProductionPresets(t *testing.T) {
	now := time.Unix(1700000000, 0)
	presetPath := filepath.Join(t.TempDir(), "presets.json")
	content := `[
		{"id":"shuihuo-image-prompt","module":"shuihuo-production","name":"画面提示词","kind":"base","version":2,"status":"published","protocolLock":{"slot":"shuihuo.prompt.image"}},
		{"id":"shuihuo-video-prompt","module":"shuihuo-production","name":"视频提示词","kind":"base","version":3,"status":"published","protocolLock":{"slot":"shuihuo.prompt.video"}},
		{"id":"shuihuo-draft","module":"shuihuo-production","name":"草稿","kind":"base","version":1,"status":"draft","protocolLock":{"slot":"shuihuo.prompt.draft"}},
		{"id":"other-module","module":"script","name":"其他模块","kind":"base","version":1,"status":"published","protocolLock":{"slot":"shuihuo.prompt.other"}}
	]`
	if err := os.WriteFile(presetPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: batchfactoryv11.NewMemoryStore(), PresetCatalog: prompt.NewCatalog(presetPath)})
	recorder := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/shuihuo-production/preset-slots", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("preset slots=%d body=%s", recorder.Code, recorder.Body.String())
	}
	body := decodeBody[map[string][]map[string]any](t, recorder)
	presets := body["presets"]
	if len(presets) != 2 {
		t.Fatalf("presets=%v", presets)
	}
	if presets[0]["id"] != "shuihuo-image-prompt" || presets[0]["slot"] != "shuihuo.prompt.image" || presets[1]["id"] != "shuihuo-video-prompt" || presets[1]["slot"] != "shuihuo.prompt.video" {
		t.Fatalf("presets=%v", presets)
	}
}
