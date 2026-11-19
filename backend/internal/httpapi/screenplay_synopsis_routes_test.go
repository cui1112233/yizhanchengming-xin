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

func TestScreenplayRoutesSynopsisGenerate(t *testing.T) {
	now := time.Date(2026, 11, 19, 3, 0, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	ctx := context.Background()
	const source = "林川在一座南方小城长大。他自幼丧父，与母亲相依为命。高中毕业后进入照相馆工作，意外发现父亲留下的日记藏着一桩旧案。为查真相，他开始追查二十年前的雨夜车祸。随着线索浮现，母亲的沉默与照相馆老板的回避都与那场事故有关。最终他在暴雨中找回被篡改的底片，明白了父亲当年的选择，故事以真相大白与母子和解收尾。"
	if _, err := store.CreateProject(ctx, screenplay.Project{ID: "synopsis-project", OwnerUsername: "alice", Title: "故事摘要", SourceText: source}); err != nil {
		t.Fatal(err)
	}
	reply := `{"title":"雨夜的底片","synopsis":"林川在南方小城长大，丧父后与母亲相依为命。他高中毕业后进入照相馆，意外发现父亲日记里藏着旧案线索，于是追查二十年前的雨夜车祸。随着调查深入，母亲的沉默与老板的回避都与事故有关。最终他在暴雨中找回被篡改的底片，明白父亲的选择，真相大白、母子和解。","audience":"男频","genre":"悬疑推理","tags":["悬疑","都市","成长","复仇"]}`
	provider := &batchTranslateProvider{result: reply}
	resolver := func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, resolver, func() time.Time { return now })
	base := "/api/comic-screenplays/projects/synopsis-project"

	// 生成成功：provider 收到原文与洛水 prompt，结果写回创作配置。
	gen := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/synopsis/generate", map[string]any{"model": "text-main"})
	if gen.Code != http.StatusOK {
		t.Fatalf("generate=%d %s", gen.Code, gen.Body.String())
	}
	var payload struct {
		Result   screenplay.SynopsisResult `json:"synopsis_result"`
		Revision int64                     `json:"revision"`
	}
	if err := json.Unmarshal(gen.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Result.Title != "雨夜的底片" || payload.Result.Audience != "男频" || payload.Result.Genre != "悬疑推理" {
		t.Fatalf("result wrong: %+v", payload.Result)
	}
	if len(payload.Result.Tags) != 4 {
		t.Fatalf("tags wrong: %v", payload.Result.Tags)
	}
	if !strings.Contains(provider.request.SystemPrompt, "300-500") {
		t.Fatalf("provider did not receive luoshui prompt: %q", provider.request.SystemPrompt)
	}
	if !strings.Contains(provider.request.UserPrompt, "林川") {
		t.Fatalf("provider did not receive source: %q", provider.request.UserPrompt)
	}

	// workspace 顶层输出 synopsis_result，前端直接可用。
	ws := signedJSONRequest(t, mux, now, "alice", http.MethodGet, base+"/workspace", nil)
	if ws.Code != http.StatusOK {
		t.Fatalf("workspace=%d %s", ws.Code, ws.Body.String())
	}
	var wsPayload struct {
		SynopsisResult *screenplay.SynopsisResult `json:"synopsis_result"`
	}
	if err := json.Unmarshal(ws.Body.Bytes(), &wsPayload); err != nil {
		t.Fatal(err)
	}
	if wsPayload.SynopsisResult == nil || wsPayload.SynopsisResult.Title != "雨夜的底片" {
		t.Fatalf("workspace synopsis_result missing: %s", ws.Body.String())
	}

	// 空原文项目拒绝生成。
	if _, err := store.CreateProject(ctx, screenplay.Project{ID: "empty-project", OwnerUsername: "alice", Title: "空原文"}); err != nil {
		t.Fatal(err)
	}
	empty := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/empty-project/synopsis/generate", map[string]any{"model": "text-main"})
	if empty.Code == http.StatusOK {
		t.Fatalf("empty source should error: %s", empty.Body.String())
	}

	// 越权 404。
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodPost, base+"/synopsis/generate", map[string]any{"model": "text-main"})
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign generate=%d %s", foreign.Code, foreign.Body.String())
	}
}
