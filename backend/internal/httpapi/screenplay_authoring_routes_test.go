package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

func TestAuthoringRoutesPaidPermissionAndChapterPublication(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	s := screenplay.NewMemoryStore()
	s.CreateProject(ctx, screenplay.Project{ID: "p", OwnerUsername: "alice", ModelID: "configured", SourceText: "真实原文"})
	repo := screenplay.NewMemoryAuthoringRepository()
	svc := screenplay.AuthoringService{Store: s, Repository: repo}
	st, _ := svc.Load(ctx, "alice", "p")
	st.Chapters = []screenplay.Chapter{{ID: "c", Title: "第一章", SourceText: "真实原文"}}
	st.Members = []screenplay.MemberGrant{{Username: "bob", Read: true, Edit: true}}
	svc.Save(ctx, "alice", "p", st, 0)
	calls := 0
	mux := http.NewServeMux()
	RegisterScreenplayAuthoringRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s, repo, func(_ context.Context, o, model string) (batchfactoryv11.DirectorProvider, error) {
		calls++
		if o != "alice" || model != "configured" {
			t.Fatalf("identity %s model %s", o, model)
		}
		return screenplayProviderFunc(func(_ context.Context, r batchfactoryv11.TextCompletionRequest) (string, error) {
			if !strings.Contains(r.UserPrompt, "真实原文") {
				t.Fatal("missing real context")
			}
			return "# 第1集：真实结果\n## 1-1：日 内 房间\n人物：小明\n小明：你好", nil
		}), nil
	})
	base := "/api/comic-screenplays/projects/p"
	for _, actor := range []string{"bob", "charlie"} {
		res := signedJSONRequest(t, mux, now, actor, "POST", base+"/chapters/c/script", map[string]any{"expected_revision": 1})
		if res.Code != 403 || calls != 0 {
			t.Fatalf("paidpermission=%d calls=%d", res.Code, calls)
		}
	}
	res := signedJSONRequest(t, mux, now, "alice", "POST", base+"/chapters/c/script", map[string]any{"expected_revision": 0})
	if res.Code != 409 || calls != 0 {
		t.Fatal(res.Code, calls)
	}
	res = signedJSONRequest(t, mux, now, "alice", "POST", base+"/chapters/c/script", map[string]any{"expected_revision": 1})
	if res.Code != 200 || calls != 1 {
		t.Fatalf("result %d %s calls %d", res.Code, res.Body.String(), calls)
	}
	doc, e := s.ActiveDocument(ctx, "alice", "p")
	if e != nil || !strings.Contains(doc.Content, "小明：你好") {
		t.Fatal(doc, e)
	}
	shots, _ := s.ListShots(ctx, "alice", "p", doc.ID)
	if len(shots) != 1 {
		t.Fatalf("chapter result not usable shots: %#v", shots)
	}
	saved, _ := repo.Load(ctx, "alice", "p")
	if saved.Chapters[0].ScriptText == "" {
		t.Fatal("chapter result missing")
	}
	read := signedJSONRequest(t, mux, now, "bob", "GET", base+"/chapters", nil)
	var body map[string]any
	json.Unmarshal(read.Body.Bytes(), &body)
	if read.Code != 200 || body["access"].(map[string]any)["generate"] != false {
		t.Fatal(read.Code, read.Body.String())
	}
}
func TestAuthoringAssistantPersistsContextAndConfirmation(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	s := screenplay.NewMemoryStore()
	s.CreateProject(ctx, screenplay.Project{ID: "p", OwnerUsername: "alice", SourceText: "内容"})
	repo := screenplay.NewMemoryAuthoringRepository()
	svc := screenplay.AuthoringService{Store: s, Repository: repo}
	st, _ := svc.Load(ctx, "alice", "p")
	st.Chapters = []screenplay.Chapter{{ID: "c", Title: "旧章", SourceText: "内容"}}
	svc.Save(ctx, "alice", "p", st, 0)
	calls := 0
	mux := http.NewServeMux()
	RegisterScreenplayAuthoringRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s, repo, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		calls++
		return screenplayProviderFunc(func(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
			return `{"reply":"建议改章名","summary":"改章名","actions":[{"kind":"chapter","chapter":{"id":"c","title":"新章","source_text":"内容","script_text":""}}]}`, nil
		}), nil
	})
	base := "/api/comic-screenplays/projects/p/assistant"
	res := signedJSONRequest(t, mux, now, "alice", "POST", base+"/chat", map[string]any{"expected_revision": 1, "message": "改章名"})
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	st, _ = repo.Load(ctx, "alice", "p")
	if len(st.Messages) != 2 || len(st.Proposals) != 1 || st.Chapters[0].Title != "旧章" {
		t.Fatal(st)
	}
	id := st.Proposals[0].ID
	res = signedJSONRequest(t, mux, now, "alice", "POST", base+"/apply", map[string]any{"expected_revision": 2, "proposal_id": id, "confirmed": false})
	if res.Code != 400 {
		t.Fatal(res.Code)
	}
	res = signedJSONRequest(t, mux, now, "alice", "POST", base+"/apply", map[string]any{"expected_revision": 2, "proposal_id": id, "confirmed": true})
	if res.Code != 200 || calls != 1 {
		t.Fatal(res.Code, res.Body.String(), calls)
	}
	st, _ = repo.Load(ctx, "alice", "p")
	if st.Chapters[0].Title != "新章" || st.Proposals[0].Status != "applied" {
		t.Fatal(st)
	}
	res = signedJSONRequest(t, mux, now, "alice", "POST", base+"/apply", map[string]any{"expected_revision": 3, "proposal_id": id, "confirmed": true})
	if res.Code != 409 {
		t.Fatal("replay", res.Code)
	}
}

type authoringTestAccounts struct{ auth.Store }

func (a authoringTestAccounts) FindAccount(_ context.Context, username string) (auth.Account, error) {
	if username == "bob" {
		return auth.Account{Username: "bob", Active: true}, nil
	}
	return auth.Account{}, auth.ErrNotFound
}
func TestAuthoringDisplayArchiveMembersAndUnknownAccount(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	s := screenplay.NewMemoryStore()
	s.CreateProject(ctx, screenplay.Project{ID: "p", OwnerUsername: "alice", Title: "测试作品"})
	repo := screenplay.NewMemoryAuthoringRepository()
	mux := http.NewServeMux()
	RegisterScreenplayAuthoringRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: authoringTestAccounts{}}, s, repo, nil)
	base := "/api/comic-screenplays/projects/p"
	res := signedJSONRequest(t, mux, now, "alice", "PUT", base+"/display", map[string]any{"expected_revision": 0, "display": map[string]any{"editor_version": "classic", "font_size": 18, "active_tab": "chapters", "show_descriptions": true}})
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	res = signedJSONRequest(t, mux, now, "alice", "PUT", base+"/team", map[string]any{"expected_revision": 1, "members": []map[string]any{{"username": "unknown", "read": true}}})
	if res.Code != 400 {
		t.Fatal(res.Code, res.Body.String())
	}
	res = signedJSONRequest(t, mux, now, "alice", "PUT", base+"/team", map[string]any{"expected_revision": 1, "members": []map[string]any{{"username": "bob", "read": true}}})
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	res = signedJSONRequest(t, mux, now, "bob", "PUT", base+"/display", map[string]any{"expected_revision": 2, "display": map[string]any{}})
	if res.Code != 403 {
		t.Fatal(res.Code)
	}
	res = signedJSONRequest(t, mux, now, "bob", "GET", "/api/comic-screenplays/shared", nil)
	if res.Code != 200 || !strings.Contains(res.Body.String(), "测试作品") {
		t.Fatal(res.Code, res.Body.String())
	}
	res = signedJSONRequest(t, mux, now, "alice", "POST", base+"/archive", map[string]any{"expected_revision": 2})
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	res = signedJSONRequest(t, mux, now, "alice", "GET", "/api/comic-screenplays/archived", nil)
	if res.Code != 200 || !strings.Contains(res.Body.String(), "测试作品") {
		t.Fatal(res.Code, res.Body.String())
	}
	res = signedJSONRequest(t, mux, now, "alice", "POST", base+"/restore", map[string]any{"expected_revision": 3})
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	st, _ := repo.Load(ctx, "alice", "p")
	if st.Archived || st.Display["editor_version"] != "classic" || st.Members[0].Generate {
		t.Fatal(st)
	}
}
func TestAuthoringArchivePausesRunningAndScheduledWithoutRestoreGeneration(t *testing.T) {
	for _, status := range []string{"running", "scheduled"} {
		t.Run(status, func(t *testing.T) {
			ctx := context.Background()
			now := time.Now()
			repo := screenplay.NewMemoryProductionRepository()
			s := &authoringWorkflowStore{screenplay.NewMemoryStore(), repo}
			s.CreateProject(ctx, screenplay.Project{ID: "p", OwnerUsername: "alice"})
			batch := screenplay.NewBatch("b", "p", "alice", "", screenplay.PermissionBypass)
			s.CreateBatch(ctx, batch)
			repo.SetProductionState(ctx, "alice", "p", status, "")
			if status == "scheduled" {
				repo.SetProductionSchedule(ctx, "alice", "p", now.Add(time.Hour))
			}
			authors := screenplay.NewMemoryAuthoringRepository()
			mux := http.NewServeMux()
			RegisterScreenplayAuthoringRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s, authors, nil)
			base := "/api/comic-screenplays/projects/p"
			res := signedJSONRequest(t, mux, now, "alice", "POST", base+"/archive", map[string]any{"expected_revision": 0})
			if res.Code != 200 {
				t.Fatal(res.Code, res.Body.String())
			}
			job, _ := repo.GetProductionJob(ctx, "alice", "p")
			batch, _ = s.GetBatch(ctx, "alice", "p")
			if job.Status != "paused" || job.RunAt != nil || batch.Status != screenplay.BatchPaused {
				t.Fatal("archive left runnable work", job, batch)
			}
			res = signedJSONRequest(t, mux, now, "alice", "POST", base+"/chapters/plan", map[string]any{"expected_revision": 1})
			if res.Code != 422 || !strings.Contains(res.Body.String(), "项目已归档") {
				t.Fatal("archived project allowed model generation", res.Code, res.Body.String())
			}
			res = signedJSONRequest(t, mux, now, "alice", "POST", base+"/restore", map[string]any{"expected_revision": 1})
			if res.Code != 200 {
				t.Fatal(res.Code, res.Body.String())
			}
			job, _ = repo.GetProductionJob(ctx, "alice", "p")
			if job.Status != "paused" {
				t.Fatal("restore resumed paid work", job)
			}
		})
	}
}
