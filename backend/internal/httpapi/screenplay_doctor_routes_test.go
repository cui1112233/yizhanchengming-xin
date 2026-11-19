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

func TestScreenplayRoutesDoctorReview(t *testing.T) {
	now := time.Date(2026, 11, 19, 2, 30, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, screenplay.Project{ID: "doctor-project", OwnerUsername: "alice", Title: "医生审核"}); err != nil {
		t.Fatal(err)
	}
	if _, err := screenplay.SaveManualDocument(ctx, store, "alice", "doctor-project", "# 第1集\n\n第一场\n林川走进巷口。"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveStoryboard(ctx, screenplay.Storyboard{
		ProjectID: "doctor-project", OwnerUsername: "alice", EpisodeNo: 1, Engine: screenplay.EngineXiaoluoV7,
		Shots: []screenplay.StoryboardShot{
			{ID: "dr-1", Seq: 1, Characters: "林川", Scene: "巷口", ImagePrompt: "巷口夜景", VideoDesc: "镜头推进"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	storyboardReply := `{"issues":[{"id":"ghost","severity":"blocker","location":"分镜1","dimension":"资产引用","problem":"人物「白桃」不在剧本中"}],"suggested_actions":["patch_shot_field"],"summary":"1 个阻断"}`
	provider := &batchTranslateProvider{result: storyboardReply}
	resolver := func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, resolver, func() time.Time { return now })
	base := "/api/comic-screenplays/projects/doctor-project"

	// 剧本医生：审核剧本全文，provider 收到剧本正文。
	screenplayReply := `{"issues":[{"id":"pacing","severity":"major","episode":1,"location":"第1场","dimension":"节奏推进","problem":"开篇铺垫过长"}],"suggested_actions":["rewrite_episode_1"],"summary":"无阻断，建议打磨开篇"}`
	provider.result = screenplayReply
	scriptReview := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/doctor/screenplay", map[string]any{"model": "text-main"})
	if scriptReview.Code != http.StatusOK {
		t.Fatalf("screenplay doctor=%d %s", scriptReview.Code, scriptReview.Body.String())
	}
	var payload struct {
		Review screenplay.DoctorReview `json:"review"`
	}
	if err := json.Unmarshal(scriptReview.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Review.Doctor != "screenplay" || !payload.Review.Pass || payload.Review.BlockNextStep {
		t.Fatalf("screenplay review wrong: %+v", payload.Review)
	}
	if len(payload.Review.Issues) != 1 || payload.Review.Issues[0].Severity != screenplay.DoctorSeverityMajor {
		t.Fatalf("issues wrong: %+v", payload.Review.Issues)
	}
	if !strings.Contains(provider.request.UserPrompt, "林川走进巷口") {
		t.Fatalf("provider did not receive script: %q", provider.request.UserPrompt)
	}

	// 分镜医生：对照剧本审核分镜，blocker → pass=false。
	provider.result = storyboardReply
	storyboardReview := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/storyboards/xiaoluo_v7_15s/1/doctor/storyboard", map[string]any{"model": "text-main"})
	if storyboardReview.Code != http.StatusOK {
		t.Fatalf("storyboard doctor=%d %s", storyboardReview.Code, storyboardReview.Body.String())
	}
	if err := json.Unmarshal(storyboardReview.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Review.Doctor != "storyboard" || payload.Review.Pass || !payload.Review.BlockNextStep {
		t.Fatalf("storyboard review wrong: %+v", payload.Review)
	}
	if !strings.Contains(provider.request.UserPrompt, "巷口夜景") {
		t.Fatalf("provider did not receive shots: %q", provider.request.UserPrompt)
	}

	// 缺少分镜的剧集报错。
	missing := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/storyboards/xiaoluo_v7_15s/2/doctor/storyboard", map[string]any{"model": "text-main"})
	if missing.Code == http.StatusOK {
		t.Fatalf("missing storyboard should error: %s", missing.Body.String())
	}

	// 越权访问 404。
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodPost, base+"/doctor/screenplay", map[string]any{"model": "text-main"})
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign doctor=%d %s", foreign.Code, foreign.Body.String())
	}
}
