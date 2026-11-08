package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt"
)

func TestPrivateCustomizationComicCreatesSinglePastedBook(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: batchfactoryv11.NewMemoryStore()})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/private-comic/tasks", map[string]any{
		"title":      "雨夜归人",
		"sourceText": "林晚走进无人的雨夜，远处的灯光在积水里慢慢晃动。",
		"initialBatchSettings": map[string]any{
			"productionMode": "comic",
			"aiPromptConfig": map[string]any{"comic": map[string]any{"style": map[string]any{"presetId": "batch-comic-style-01"}}},
		},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Batch batchfactoryv11.Batch `json:"batch"`
		Book  batchfactoryv11.Book  `json:"book"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Batch.ID == "" || payload.Book.ID == "" {
		t.Fatalf("missing created ids: %+v", payload)
	}
	if payload.Book.Title != "雨夜归人" || payload.Book.SourceText != "林晚走进无人的雨夜，远处的灯光在积水里慢慢晃动。" {
		t.Fatalf("book content not preserved: %+v", payload.Book)
	}
	if len(payload.Batch.Books) != 1 {
		t.Fatalf("want one book, got %d", len(payload.Batch.Books))
	}
}

func TestPrivateCustomizationComicRejectsEmptySource(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: batchfactoryv11.NewMemoryStore()})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/private-comic/tasks", map[string]any{"title": "空稿", "sourceText": "  "})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPrivateCustomizationComicRejectsTooShortSource(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: batchfactoryv11.NewMemoryStore()})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/private-comic/tasks", map[string]any{"title": "短稿", "sourceText": "太短"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPrivateCustomizationComicHydratesPublishedComicSelections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	data := `[{"id":"batch-comic-assets","module":"batch-factory","name":"漫画资产","kind":"base","version":1,"revision":1,"status":"published","body":"漫画资产正文","protocolLock":{"format":"json","slot":"batch.comic-assets"}},{"id":"batch-comic-page","module":"batch-factory","name":"漫画页","kind":"base","version":1,"revision":1,"status":"published","body":"漫画页正文","protocolLock":{"format":"json","slot":"batch.comic-page"}},{"id":"batch-comic-style-01","module":"batch-factory","name":"韩漫","kind":"base","version":1,"revision":1,"status":"published","body":"韩漫画风正文","protocolLock":{"format":"style","slot":"batch.comic-style"}}]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: batchfactoryv11.NewMemoryStore(), PresetCatalog: prompt.NewCatalog(path)})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/private-comic/tasks", map[string]any{
		"sourceText": "林晚走进无人的雨夜，远处的灯光在积水里慢慢晃动。",
		"initialBatchSettings": map[string]any{"aiPromptConfig": map[string]any{"comic": map[string]any{
			"presetId": "batch-comic-page",
			"style":    map[string]any{"presetId": "batch-comic-style-01"},
		}}},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Batch batchfactoryv11.Batch `json:"batch"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(payload.Batch.SettingsState.Patch["aiPromptConfig"], &config); err != nil {
		t.Fatal(err)
	}
	comic := config["comic"].(map[string]any)
	if comic["body"] != "漫画页正文" || comic["style"].(map[string]any)["body"] != "韩漫画风正文" {
		t.Fatalf("comic selections were not hydrated: %#v", comic)
	}
}

func TestPrivateCustomizationComicUsesPublishedSystemPromptsWhenNoBatchPresetWasChosen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	data := `[{"id":"batch-comic-assets","module":"batch-factory","name":"漫画资产","kind":"base","version":3,"revision":1,"status":"published","body":"后台漫画资产正文","protocolLock":{"format":"json","slot":"batch.comic-assets"}},{"id":"batch-comic-page","module":"batch-factory","name":"漫画分页","kind":"base","version":4,"revision":1,"status":"published","body":"后台漫画分页正文","protocolLock":{"format":"json","slot":"batch.comic-page"}}]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: batchfactoryv11.NewMemoryStore(), PresetCatalog: prompt.NewCatalog(path)})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/private-comic/tasks", map[string]any{
		"sourceText":           "赵家堂屋的门帘是新换的，红布面压着一道黄边，屋里没有一个人。",
		"initialBatchSettings": map[string]any{"aiPromptConfig": map[string]any{"comic": map[string]any{"enabled": true}}},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Batch batchfactoryv11.Batch `json:"batch"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(payload.Batch.SettingsState.Patch["aiPromptConfig"], &config); err != nil {
		t.Fatal(err)
	}
	comic := config["comic"].(map[string]any)
	assets := comic["assets"].(map[string]any)
	if assets["presetId"] != "batch-comic-assets" || assets["body"] != "后台漫画资产正文" {
		t.Fatalf("private comic asset prompt did not use the backend default: %#v", assets)
	}
	if comic["presetId"] != "batch-comic-page" || comic["body"] != "后台漫画分页正文" {
		t.Fatalf("private comic page prompt did not use the backend default: %#v", comic)
	}
}

func TestPrivateCustomizationComicSettingsRepairLegacyTaskWithoutBatchPreset(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	data := `[{"id":"batch-comic-assets","module":"batch-factory","name":"漫画资产","kind":"base","version":3,"revision":1,"status":"published","body":"后台漫画资产正文","protocolLock":{"format":"json","slot":"batch.comic-assets"}},{"id":"batch-comic-page","module":"batch-factory","name":"漫画分页","kind":"base","version":4,"revision":1,"status":"published","body":"后台漫画分页正文","protocolLock":{"format":"json","slot":"batch.comic-page"}}]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	store := batchfactoryv11.NewMemoryStore()
	intake, err := store.CreateIntake(context.Background(), "alice", batchfactoryv11.NovelFetchIntakeInput{
		Books:    []batchfactoryv11.CreateBookInput{{ID: "legacy-private-book", BookID: "legacy-private-book", Title: "旧单本", Platform: "private_customization", SourceText: "这是一个旧版创建的单本漫画任务，里面还没有保存系统提示词。", TxtText: "这是一个旧版创建的单本漫画任务，里面还没有保存系统提示词。"}},
		Metadata: map[string]any{"sourceMode": "private_customization"},
	})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := store.CreateBatchFromIntake(context.Background(), "alice", intake.ID, batchfactoryv11.CreateBatchInput{
		Title: "旧单本",
		InitialBatchSettings: batchfactoryv11.SettingsPatch{
			"productionMode": json.RawMessage(`"comic"`),
			"aiPromptConfig": json.RawMessage(`{"comic":{"enabled":true,"scope":"private_customization"}}`),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, PresetCatalog: prompt.NewCatalog(path)})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPut, "/api/batch-factory/v12/batches/"+batch.ID+"/settings", map[string]any{
		"expectedRevision": batch.Revision,
		"patch": map[string]any{
			"productionMode": "comic",
			"aiPromptConfig": map[string]any{"comic": map[string]any{"enabled": true, "scope": "private_customization"}},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	updated, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(updated.SettingsState.Patch["aiPromptConfig"], &config); err != nil {
		t.Fatal(err)
	}
	comic := config["comic"].(map[string]any)
	assets := comic["assets"].(map[string]any)
	if assets["body"] != "后台漫画资产正文" || comic["body"] != "后台漫画分页正文" {
		t.Fatalf("legacy private comic prompts were not repaired: %#v", comic)
	}
}
