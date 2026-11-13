package httpapi

import (
	"context"
	"net/http"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

type authoringWorkflowStore struct {
	*screenplay.MemoryStore
	screenplay.ProductionRepository
}

func TestScreenplayWorkflowSharedRightsArchiveLockAndInvalidate(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	s := &authoringWorkflowStore{screenplay.NewMemoryStore(), screenplay.NewMemoryProductionRepository()}
	s.CreateProject(ctx, screenplay.Project{ID: "p", OwnerUsername: "alice", Title: "项目", SourceText: "旧原文", PermissionMode: screenplay.PermissionBypass})
	s.CreateProject(ctx, screenplay.Project{ID: "archive", OwnerUsername: "alice", Title: "归档项目"})
	repo := screenplay.NewMemoryAuthoringRepository()
	svc := &screenplay.AuthoringService{Store: s, Repository: repo}
	st, _ := svc.Load(ctx, "alice", "p")
	st.Members = []screenplay.MemberGrant{{Username: "bob", Read: true, Edit: true}}
	st, _ = svc.Save(ctx, "alice", "p", st, 0)
	arch, _ := svc.Load(ctx, "alice", "archive")
	arch.Archived = true
	svc.Save(ctx, "alice", "archive", arch, 0)
	calls := 0
	mux := http.NewServeMux()
	RegisterScreenplayWorkflowRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s, func(_ context.Context, o, _ string) (batchfactoryv11.DirectorProvider, error) {
		calls++
		if o != "alice" {
			t.Fatal("shared model leaked identity", o)
		}
		return screenplayProviderFunc(func(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
			return "# 第1集：完成\n## 1-1：日 内 家\n△ 人物进入", nil
		}), nil
	}, func() time.Time { return now }, nil, svc)
	base := "/api/comic-screenplays/projects/p"
	for _, path := range []string{"/generate", "/assets/extract", "/storyboards/generate", "/batch", "/batch/run-next"} {
		res := signedJSONRequest(t, mux, now, "bob", "POST", base+path, map[string]any{})
		if res.Code != 403 || calls != 0 {
			t.Fatal(path, res.Code, calls)
		}
	}
	res := signedJSONRequest(t, mux, now, "bob", "GET", base+"/workspace", nil)
	if res.Code != 200 || !strings.Contains(res.Body.String(), "旧原文") {
		t.Fatal(res.Code, res.Body.String())
	}
	res = signedJSONRequest(t, mux, now, "alice", "GET", "/api/comic-screenplays/projects", nil)
	if res.Code != 200 || strings.Contains(res.Body.String(), "归档项目") {
		t.Fatal(res.Code, res.Body.String())
	}
	s.AcquireLock(ctx, "authoring:alice:p", "held", time.Minute)
	res = signedJSONRequest(t, mux, now, "bob", "PUT", base+"/source", map[string]any{"content": "冲突原文"})
	if res.Code != 409 {
		t.Fatal("lock", res.Code)
	}
	s.ReleaseLock(ctx, "authoring:alice:p", "held")
	b := screenplay.NewBatch("b", "p", "alice", "", screenplay.PermissionBypass)
	b.Status = screenplay.BatchCompleted
	for _, step := range b.Steps {
		step.Status = screenplay.StepDone
	}
	s.CreateBatch(ctx, b)
	s.SetProductionState(ctx, "alice", "p", "running", "")
	res = signedJSONRequest(t, mux, now, "bob", "PUT", base+"/source", map[string]any{"content": "新原文"})
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	p, _ := s.GetProject(ctx, "alice", "p")
	if p.SourceText != "新原文" {
		t.Fatal(p)
	}
	b, _ = s.GetBatch(ctx, "alice", "p")
	job, _ := s.GetProductionJob(ctx, "alice", "p")
	if b.Status != screenplay.BatchPaused || job.Status != "paused" || b.Step("1_screenplay").Status != screenplay.StepStale {
		t.Fatal(b, job)
	}
	res = signedJSONRequest(t, mux, now, "bob", "DELETE", base, nil)
	if res.Code != 403 {
		t.Fatal("delete", res.Code)
	}
	st, _ = svc.Load(ctx, "alice", "p")
	st.Members[0].Generate = true
	st, _ = svc.Save(ctx, "alice", "p", st, st.Revision)
	res = signedJSONRequest(t, mux, now, "bob", "POST", base+"/generate", map[string]any{"mode": "a1", "content": "原文", "confirmed": true})
	if res.Code != 200 || calls != 1 {
		t.Fatal("grant", res.Code, res.Body.String(), calls)
	}
	s.MutateBatch(ctx, "alice", "p", func(batch *screenplay.Batch) error {
		batch.Status = screenplay.BatchRunning
		batch.Step("0_ingest").Status = screenplay.StepPending
		return nil
	})
	res = signedJSONRequest(t, mux, now, "bob", "POST", base+"/batch/run-next", map[string]any{})
	if res.Code != 200 {
		t.Fatal("nested authoring lock blocks legacy unit", res.Code, res.Body.String())
	}
	s.SetProductionState(ctx, "alice", "p", "running", "")
	res = signedJSONRequest(t, mux, now, "bob", "POST", base+"/batch/cancel", map[string]any{})
	job, _ = s.GetProductionJob(ctx, "alice", "p")
	if res.Code != 200 || job.Status != "paused" {
		t.Fatal(res.Code, job)
	}
}
