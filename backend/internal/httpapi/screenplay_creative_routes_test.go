package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

func TestScreenplayCreativeAPIIsolationPatchConflictAndClear(t *testing.T) {
	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.UTC)
	s := screenplay.NewMemoryStore()
	_, _ = s.CreateProject(context.Background(), screenplay.Project{ID: "p", OwnerUsername: "alice"})
	mux := http.NewServeMux()
	RegisterScreenplayCreativeRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s, nil)
	req := func(actor, method, path string, body any) (int, string) {
		w := signedJSONRequest(t, mux, now, actor, method, path, body)
		return w.Code, w.Body.String()
	}
	code, body := req("alice", "POST", "/api/screenplays/styles", map[string]any{"name": "私人风格", "prompt": "线条轻盈", "images": []string{}})
	if code != 201 {
		t.Fatalf("create %d %s", code, body)
	}
	var created struct {
		Style screenplay.UserStyle `json:"style"`
	}
	_ = json.Unmarshal([]byte(body), &created)
	id := created.Style.ID
	if code, body = req("bob", "GET", "/api/screenplays/creative-library", nil); code != 200 || strings.Contains(body, id) {
		t.Fatalf("library %d %s", code, body)
	}
	if code, _ = req("bob", "PUT", "/api/screenplays/styles/"+id, map[string]any{"name": "偷改", "prompt": "x", "expected_revision": 0}); code != 404 {
		t.Fatalf("foreign style status %d", code)
	}
	if code, body = req("alice", "PUT", "/api/screenplays/p/creative-config", map[string]any{"expected_revision": 0, "style_id": id, "script_prompt_id": prompt.CreativeScript144, "storyboard_prompt_id": prompt.CreativeStoryboard307, "character_design_prompt_id": prompt.CreativeCharacter213, "scene_design_prompt_id": prompt.CreativeScene289, "image_core_instruction": "构图清楚"}); code != 200 {
		t.Fatalf("save %d %s", code, body)
	}
	p, _ := s.GetProject(context.Background(), "alice", "p")
	if p.StyleDesc != "线条轻盈" || p.CreativeConfig.PromptSnapshots["script"].Body == "" {
		t.Fatal("missing final snapshots")
	}
	if strings.Contains(body, "prompt_snapshots") || strings.Contains(body, "script_content") {
		t.Fatalf("system body leaked %s", body)
	}
	if code, _ = req("alice", "PUT", "/api/screenplays/p/creative-config", map[string]any{"expected_revision": 0, "image_core_instruction": "stale"}); code != 409 {
		t.Fatalf("conflict %d", code)
	}
	if code, _ = req("alice", "PUT", "/api/screenplays/p/creative-config", map[string]any{"expected_revision": 1, "script_prompt_id": prompt.CreativeScene289}); code != 400 {
		t.Fatalf("wrong purpose %d", code)
	}
	if code, _ = req("alice", "PUT", "/api/screenplays/p/creative-config", map[string]any{"expected_revision": 1, "image_core_instruction": "改变要求"}); code != 200 {
		t.Fatalf("patch %d", code)
	}
	p, _ = s.GetProject(context.Background(), "alice", "p")
	if p.StyleDesc != "线条轻盈" || p.CreativeConfig.StoryboardPromptID != prompt.CreativeStoryboard307 {
		t.Fatal("omitted fields overwritten")
	}
	if code, body = req("alice", "PUT", "/api/screenplays/p/creative-config", map[string]any{"expected_revision": 2, "style_id": nil}); code != 200 {
		t.Fatalf("clear %d %s", code, body)
	}
	p, _ = s.GetProject(context.Background(), "alice", "p")
	if p.StyleDesc != "" || p.ArtStyle != "" || p.SelectedStyle != "" || p.CreativeConfig.Style != nil {
		t.Fatal("style not cleared")
	}
	if code, _ = req("bob", "GET", "/api/screenplays/p/creative-config", nil); code != 404 {
		t.Fatalf("foreign config %d", code)
	}
	for _, ref := range []string{"F:/private/a.png", "https://example.com/a.png", "/api/reference-assets/file/character/a/b?secret=x", "/api/reference-assets/file/character/../b"} {
		if code, _ = req("alice", "POST", "/api/screenplays/styles", map[string]any{"name": "x", "prompt": "x", "images": []string{ref}}); code != 400 {
			t.Fatalf("bad image ref accepted: %s", ref)
		}
	}
}
func TestScreenplayCreativeSharedReadOwnerWriteAndLock(t *testing.T) {
	now := time.Now()
	ctx := context.Background()
	s := screenplay.NewMemoryStore()
	_, _ = s.CreateProject(ctx, screenplay.Project{ID: "shared", OwnerUsername: "alice"})
	_, _ = s.SaveUserStyle(ctx, screenplay.UserStyle{ID: "secret", OwnerUsername: "alice", Name: "未分享画风", Prompt: "私人描述"}, nil)
	repo := screenplay.NewMemoryAuthoringRepository()
	svc := &screenplay.AuthoringService{Store: s, Repository: repo}
	st, _ := svc.Load(ctx, "alice", "shared")
	st.Members = []screenplay.MemberGrant{{Username: "bob", Read: true, Edit: true, Generate: true}}
	_, _ = svc.Save(ctx, "alice", "shared", st, st.Revision)
	mux := http.NewServeMux()
	RegisterScreenplayCreativeRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s, svc)
	read := signedJSONRequest(t, mux, now, "bob", "GET", "/api/screenplays/shared/creative-config", nil)
	if read.Code != 200 || strings.Contains(read.Body.String(), "私人描述") {
		t.Fatalf("shared read %d %s", read.Code, read.Body.String())
	}
	write := signedJSONRequest(t, mux, now, "bob", "PUT", "/api/screenplays/shared/creative-config", map[string]any{"expected_revision": 0, "style_id": "secret"})
	if write.Code != 403 {
		t.Fatalf("shared write %d %s", write.Code, write.Body.String())
	}
	library := signedJSONRequest(t, mux, now, "bob", "GET", "/api/screenplays/creative-library", nil)
	if strings.Contains(library.Body.String(), "未分享画风") {
		t.Fatal("owner library leaked")
	}
	_ = s.AcquireLock(ctx, "authoring:alice:shared", "running", time.Minute)
	write = signedJSONRequest(t, mux, now, "alice", "PUT", "/api/screenplays/shared/creative-config", map[string]any{"expected_revision": 0, "style_id": "secret"})
	if write.Code != 409 {
		t.Fatalf("lock %d %s", write.Code, write.Body.String())
	}
	_ = s.ReleaseLock(ctx, "authoring:alice:shared", "running")
	write = signedJSONRequest(t, mux, now, "alice", "PUT", "/api/screenplays/shared/creative-config", map[string]any{"expected_revision": 0, "style_id": "secret"})
	if write.Code != 200 {
		t.Fatalf("owner write %d %s", write.Code, write.Body.String())
	}
	read = signedJSONRequest(t, mux, now, "bob", "GET", "/api/screenplays/shared/creative-config", nil)
	if read.Code != 200 || !strings.Contains(read.Body.String(), "私人描述") {
		t.Fatalf("frozen shared selection %d %s", read.Code, read.Body.String())
	}
}
func TestCreativeStyleCompatibilityPreservesServerSnapshot(t *testing.T) {
	p := screenplay.Project{ArtStyle: "已冻结", SelectedStyle: "private", StyleDesc: "后端冻结文字", CreativeConfig: &screenplay.CreativeConfig{Style: &screenplay.CreativeStyleSnapshot{ID: "private", Name: "已冻结", Prompt: "后端冻结文字"}}}
	for _, selection := range []string{"private", ""} {
		got, e := resolveScreenplayStyle(nil, selection, "前端篡改名字", &p)
		if e != nil || got.StyleDesc != p.StyleDesc || got.ArtStyle != p.ArtStyle || got.SelectedStyle != p.SelectedStyle {
			t.Fatalf("preserve snapshot %+v %v", got, e)
		}
	}
	if _, e := resolveScreenplayStyle(nil, "batch-style", "批量画风", &p); e == nil {
		t.Fatal("legacy route replaced private style")
	}
}
