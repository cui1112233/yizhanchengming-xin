package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
)

// 管理后台发布「剧本医生系统词」覆盖版本后，剧本医生路由运行时改用覆盖正文
//（V88 内置运行词 v88_doctor_screenplay 的 RuntimeBody 通道）。
func TestDoctorRouteUsesPublishedPromptOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	raw := `[{"id":"luoshui.v88_doctor_screenplay","module":"luoshui-plugin","name":"剧本医生系统词","version":1,"revision":2,"status":"published","body":"管理员定制的剧本医生提示词"}]`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	prompt.SetLuoshuiOverrideCatalog(prompt.NewCatalog(path))
	t.Cleanup(func() { prompt.SetLuoshuiOverrideCatalog(nil) })

	now := time.Date(2026, 11, 19, 7, 0, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, screenplay.Project{ID: "doctor-override", OwnerUsername: "alice", Title: "医生覆盖"}); err != nil {
		t.Fatal(err)
	}
	if _, err := screenplay.SaveManualDocument(ctx, store, "alice", "doctor-override", "# 第1集\n\n第一场\n林川走进巷口。"); err != nil {
		t.Fatal(err)
	}
	provider := &batchTranslateProvider{result: `{"issues":[],"suggested_actions":[],"summary":"ok"}`}
	resolver := func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, resolver, func() time.Time { return now })

	r := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/doctor-override/doctor/screenplay", map[string]any{"model": "text-main"})
	if r.Code != http.StatusOK {
		t.Fatalf("doctor=%d %s", r.Code, r.Body.String())
	}
	if provider.request.SystemPrompt != "管理员定制的剧本医生提示词" {
		t.Fatalf("医生路由未使用覆盖提示词: %q", provider.request.SystemPrompt)
	}
	var payload struct {
		Review screenplay.DoctorReview `json:"review"`
	}
	if err := json.Unmarshal(r.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Review.Pass {
		t.Fatalf("review wrong: %+v", payload.Review)
	}
}

// 故事概要同样走覆盖通道：发布 v88_synopsis 覆盖后，概要生成使用覆盖正文。
func TestSynopsisRouteUsesPublishedPromptOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	raw := `[{"id":"luoshui.v88_synopsis","module":"luoshui-plugin","name":"故事概要系统词","version":1,"revision":2,"status":"published","body":"管理员定制的故事概要提示词"}]`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	prompt.SetLuoshuiOverrideCatalog(prompt.NewCatalog(path))
	t.Cleanup(func() { prompt.SetLuoshuiOverrideCatalog(nil) })

	now := time.Date(2026, 11, 19, 7, 30, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	ctx := context.Background()
	const source = "林川在南方小城长大，丧父后与母亲相依为命。他进入照相馆后发现父亲日记里藏着旧案线索，追查雨夜车祸，最终真相大白、母子和解。"
	if _, err := store.CreateProject(ctx, screenplay.Project{ID: "synopsis-override", OwnerUsername: "alice", Title: "概要覆盖", SourceText: source}); err != nil {
		t.Fatal(err)
	}
	provider := &batchTranslateProvider{result: `{"title":"雨夜的底片","synopsis":"林川在南方小城长大，丧父后与母亲相依为命。他进入照相馆后发现父亲日记里藏着旧案线索，于是追查二十年前的雨夜车祸。母亲的沉默与老板的回避都与事故有关。最终真相大白、母子和解。","audience":"男频","genre":"悬疑推理","tags":["悬疑","都市","成长","复仇"]}`}
	resolver := func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, resolver, func() time.Time { return now })

	r := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/synopsis-override/synopsis/generate", map[string]any{"model": "text-main"})
	if r.Code != http.StatusOK {
		t.Fatalf("synopsis=%d %s", r.Code, r.Body.String())
	}
	if !strings.Contains(provider.request.SystemPrompt, "管理员定制的故事概要提示词") {
		t.Fatalf("概要路由未使用覆盖提示词: %q", provider.request.SystemPrompt)
	}
}
