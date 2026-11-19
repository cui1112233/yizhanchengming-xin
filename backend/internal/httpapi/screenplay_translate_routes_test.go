package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
)

type batchTranslateProvider struct {
	result  string
	request batchfactoryv11.TextCompletionRequest
}

func (p *batchTranslateProvider) Complete(_ context.Context, input batchfactoryv11.TextCompletionRequest) (string, error) {
	p.request = input
	return p.result, nil
}

func TestScreenplayRoutesBatchTranslate(t *testing.T) {
	now := time.Date(2026, 11, 18, 23, 40, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, screenplay.Project{ID: "translate-project", OwnerUsername: "alice", Title: "批量翻译"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveStoryboard(ctx, screenplay.Storyboard{
		ProjectID: "translate-project", OwnerUsername: "alice", EpisodeNo: 1, Engine: screenplay.EngineXiaoluoV7,
		Shots: []screenplay.StoryboardShot{
			{ID: "bt-1", Seq: 1, ImagePrompt: "夜晚的街道", VideoDesc: "镜头推进", Line: "你好"},
			{ID: "bt-2", Seq: 2, ImagePrompt: "", VideoDesc: "", Line: "只有台词"},
			{ID: "bt-3", Seq: 3, ImagePrompt: "卧室清晨", VideoDesc: "", Line: ""},
		},
	}); err != nil {
		t.Fatal(err)
	}
	provider := &batchTranslateProvider{result: "night street"}
	resolver := func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, resolver, func() time.Time { return now })
	base := "/api/comic-screenplays/projects/translate-project/storyboards/xiaoluo_v7_15s/1/translate"

	prepare := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/prepare", map[string]any{
		"field": "image", "direction": "zh2en", "scope": "all",
	})
	if prepare.Code != http.StatusOK {
		t.Fatalf("prepare=%d %s", prepare.Code, prepare.Body.String())
	}
	var prepared struct {
		Items []screenplay.BatchTranslateItem `json:"items"`
		Total int                             `json:"total"`
	}
	if err := json.Unmarshal(prepare.Body.Bytes(), &prepared); err != nil {
		t.Fatal(err)
	}
	// bt-2 画面提示词为空 → 回退台词；bt-3 有提示词。
	if prepared.Total != 3 || len(prepared.Items) != 3 {
		t.Fatalf("prepare total=%d %s", prepared.Total, prepare.Body.String())
	}
	if !prepared.Items[1].FromSubtitle || prepared.Items[1].Text != "只有台词" {
		t.Fatalf("字幕回退错误: %+v", prepared.Items[1])
	}
	if containsText(prepare.Body.String(), `"storyboard"`) {
		t.Fatalf("prepare 不应写库：%s", prepare.Body.String())
	}

	translated := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/item", map[string]any{
		"shot_id": "bt-1", "field": "image", "direction": "zh2en", "model": "text-main",
	})
	if translated.Code != http.StatusOK {
		t.Fatalf("item=%d %s", translated.Code, translated.Body.String())
	}
	if !containsText(translated.Body.String(), `"text":"night street"`) {
		t.Fatalf("item text = %s", translated.Body.String())
	}
	if provider.request.SystemPrompt == "" || !strings.Contains(provider.request.UserPrompt, "夜晚的街道") {
		t.Fatalf("模型输入不完整: system=%q user=%q", provider.request.SystemPrompt, provider.request.UserPrompt)
	}
	stored, err := store.GetStoryboard(ctx, "alice", "translate-project", 1, screenplay.EngineXiaoluoV7)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Shots[0].ImagePrompt != "night street" {
		t.Fatalf("画面提示词未写回: %q", stored.Shots[0].ImagePrompt)
	}
	if stored.Shots[0].VideoDesc != "镜头推进" {
		t.Fatalf("未选字段被改动: %q", stored.Shots[0].VideoDesc)
	}

	empty := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/prepare", map[string]any{
		"field": "video", "direction": "zh2en", "scope": "custom", "shot_start": 90, "shot_end": 100,
	})
	if empty.Code != http.StatusOK || !containsText(empty.Body.String(), "没有需要翻译的分镜") {
		t.Fatalf("empty prepare=%d %s", empty.Code, empty.Body.String())
	}

	badField := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/item", map[string]any{
		"shot_id": "bt-1", "field": "audio", "direction": "zh2en",
	})
	if badField.Code == http.StatusOK {
		t.Fatalf("非法字段应报错：%s", badField.Body.String())
	}

	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodPost, base+"/item", map[string]any{
		"shot_id": "bt-1", "field": "image", "direction": "zh2en",
	})
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign item=%d %s", foreign.Code, foreign.Body.String())
	}
}
