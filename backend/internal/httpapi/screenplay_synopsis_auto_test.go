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

// 洛水对齐：资产解析入口自动附带「研读剧本 · 故事概要」。
// 洛水流水线在人物/场景/道具解析之前自动跑 synopsis Agent；V88 在 Extract 前自动补跑。
func TestScreenplayAssetExtractAutoSynopsis(t *testing.T) {
	now := time.Date(2026, 11, 19, 5, 0, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	ctx := context.Background()
	const source = "林川在一座南方小城长大。他自幼丧父，与母亲相依为命。高中毕业后进入照相馆工作，意外发现父亲留下的日记藏着一桩旧案。为查真相，他开始追查二十年前的雨夜车祸。最终真相大白，母子和解。"
	if _, err := store.CreateProject(ctx, screenplay.Project{ID: "auto-synopsis", OwnerUsername: "alice", Title: "自动概要", SourceText: source}); err != nil {
		t.Fatal(err)
	}
	const reply = `{"title":"雨夜的底片","synopsis":"林川在南方小城长大，丧父后与母亲相依为命。他进入照相馆后发现父亲日记里藏着旧案线索，于是追查二十年前的雨夜车祸。母亲的沉默与老板的回避都与事故有关。最终真相大白、母子和解。","audience":"男频","genre":"悬疑推理","tags":["悬疑","都市","成长","复仇"]}`
	const characters = `{"characters":[{"name":"林川","prompt":"年轻男子，照相馆学徒"}]}`
	synopsisCalls, assetCalls := 0, 0
	var firstSynopsisUserPrompt string
	provider := screenplayProviderFunc(func(_ context.Context, req batchfactoryv11.TextCompletionRequest) (string, error) {
		if strings.Contains(req.SystemPrompt, "故事摘要") {
			synopsisCalls++
			if synopsisCalls == 1 {
				firstSynopsisUserPrompt = req.UserPrompt
			}
			return reply, nil
		}
		assetCalls++
		return characters, nil
	})
	resolver := func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, resolver, func() time.Time { return now })
	base := "/api/comic-screenplays/projects/auto-synopsis"

	// 第一次资产解析：先自动生成概要（收到洛水 prompt 与原文），再解析资产；响应附带 synopsis_result。
	r := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/assets/extract", map[string]any{"type": "character", "degree": "standard", "model": "own-text", "content": "林川遇见老板。", "confirmed": true})
	if r.Code != http.StatusOK {
		t.Fatalf("extract=%d %s", r.Code, r.Body.String())
	}
	var payload struct {
		AssetResult     screenplay.AssetResult     `json:"-"`
		SynopsisResult  *screenplay.SynopsisResult `json:"synopsis_result"`
		SynopsisWarning string                     `json:"synopsis_warning"`
	}
	if err := json.Unmarshal(r.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.SynopsisResult == nil || payload.SynopsisResult.Title != "雨夜的底片" || payload.SynopsisResult.Genre != "悬疑推理" {
		t.Fatalf("auto synopsis wrong: %s", r.Body.String())
	}
	if payload.SynopsisWarning != "" {
		t.Fatalf("unexpected warning: %q", payload.SynopsisWarning)
	}
	if !strings.Contains(firstSynopsisUserPrompt, "林川") {
		t.Fatalf("synopsis provider did not receive source: %q", firstSynopsisUserPrompt)
	}
	if synopsisCalls != 1 || assetCalls != 1 {
		t.Fatalf("calls synopsis=%d asset=%d", synopsisCalls, assetCalls)
	}

	// 创作配置已落库，workspace 顶层输出。
	ws := signedJSONRequest(t, mux, now, "alice", http.MethodGet, base+"/workspace", nil)
	var wsPayload struct {
		SynopsisResult *screenplay.SynopsisResult `json:"synopsis_result"`
	}
	if err := json.Unmarshal(ws.Body.Bytes(), &wsPayload); err != nil {
		t.Fatal(err)
	}
	if wsPayload.SynopsisResult == nil || wsPayload.SynopsisResult.Title != "雨夜的底片" {
		t.Fatalf("workspace synopsis missing: %s", ws.Body.String())
	}

	// 第二次资产解析：已有概要结果，跳过自动生成（概要调用数不再增加），资产照常解析。
	r2 := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/assets/extract", map[string]any{"type": "character", "degree": "standard", "model": "own-text", "content": "林川遇见老板。", "confirmed": true})
	if r2.Code != http.StatusOK {
		t.Fatalf("extract2=%d %s", r2.Code, r2.Body.String())
	}
	if synopsisCalls != 1 || assetCalls != 2 {
		t.Fatalf("after second extract calls synopsis=%d asset=%d", synopsisCalls, assetCalls)
	}
}

// 概要生成失败不阻断资产解析：资产照常落库，响应带 synopsis_warning 且无 synopsis_result。
func TestScreenplayAssetExtractAutoSynopsisFailure(t *testing.T) {
	now := time.Date(2026, 11, 19, 5, 30, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, screenplay.Project{ID: "auto-synopsis-fail", OwnerUsername: "alice", Title: "概要失败", SourceText: "甲遇见乙。"}); err != nil {
		t.Fatal(err)
	}
	const characters = `{"characters":[{"name":"甲","prompt":"主角"}]}`
	provider := screenplayProviderFunc(func(_ context.Context, req batchfactoryv11.TextCompletionRequest) (string, error) {
		if strings.Contains(req.SystemPrompt, "故事摘要") {
			return "", context.DeadlineExceeded
		}
		return characters, nil
	})
	resolver := func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, resolver, func() time.Time { return now })

	r := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/auto-synopsis-fail/assets/extract", map[string]any{"type": "character", "degree": "standard", "model": "own-text", "content": "甲遇见乙。", "confirmed": true})
	if r.Code != http.StatusOK {
		t.Fatalf("extract=%d %s", r.Code, r.Body.String())
	}
	var payload struct {
		SynopsisResult  *screenplay.SynopsisResult `json:"synopsis_result"`
		SynopsisWarning string                     `json:"synopsis_warning"`
	}
	if err := json.Unmarshal(r.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.SynopsisResult != nil {
		t.Fatalf("unexpected synopsis: %+v", payload.SynopsisResult)
	}
	if payload.SynopsisWarning == "" {
		t.Fatalf("expected synopsis_warning: %s", r.Body.String())
	}
	rows, _ := store.ListAssets(ctx, "alice", "auto-synopsis-fail", screenplay.AssetCharacter)
	if len(rows) != 1 || rows[0].Name != "甲" {
		t.Fatalf("assets not extracted: %+v", rows)
	}
}

// 无原文/无模型/standalone 重解析：不触发自动概要（静默跳过）。
func TestScreenplayAssetExtractAutoSynopsisSkip(t *testing.T) {
	now := time.Date(2026, 11, 19, 6, 0, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, screenplay.Project{ID: "no-source", OwnerUsername: "alice", Title: "无原文", ModelID: "our-configured-model", PermissionMode: screenplay.PermissionBypass}); err != nil {
		t.Fatal(err)
	}
	const characters = `{"characters":[{"name":"甲","prompt":"主角"}]}`
	synopsisCalls := 0
	provider := screenplayProviderFunc(func(_ context.Context, req batchfactoryv11.TextCompletionRequest) (string, error) {
		if strings.Contains(req.SystemPrompt, "故事摘要") {
			synopsisCalls++
		}
		return characters, nil
	})
	resolver := func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, resolver, func() time.Time { return now })

	// 无原文项目：静默跳过，解析照常。
	r := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/no-source/assets/extract", map[string]any{"type": "character", "degree": "standard", "model": "own-text", "content": "甲遇见乙。", "confirmed": true})
	if r.Code != http.StatusOK {
		t.Fatalf("extract=%d %s", r.Code, r.Body.String())
	}
	// standalone 重解析：不触发概要。
	r2 := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/no-source/assets/extract", map[string]any{"mode": "standalone", "type": "character", "content": "甲遇见乙。", "confirmed": true})
	if r2.Code != http.StatusOK {
		t.Fatalf("reparse=%d %s", r2.Code, r2.Body.String())
	}
	if synopsisCalls != 0 {
		t.Fatalf("unexpected synopsis calls=%d", synopsisCalls)
	}
}
