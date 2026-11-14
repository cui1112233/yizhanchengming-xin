package httpapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	"qiantie/backend/internal/screenplay"
)

type fakeScreenplayVideoService struct {
	inputs []scriptVideoLocalInput
}

func (f *fakeScreenplayVideoService) Submit(_ context.Context, _ string, input scriptVideoLocalInput) (scriptVideoSubmission, error) {
	f.inputs = append(f.inputs, input)
	return scriptVideoSubmission{TaskID: "task-1", Status: "processing"}, nil
}

func (f *fakeScreenplayVideoService) Poll(_ context.Context, _ string, _ string) (scriptVideoStatus, error) {
	return scriptVideoStatus{TaskID: "task-1", Status: "succeeded", VideoURL: "https://media.example/shot.mp4"}, nil
}

func TestScreenplayVideoRunnerUsesSavedAccountModelAndReturnsRealArtifact(t *testing.T) {
	dir := t.TempDir()
	configs := NewUserConfigStore(dir)
	if err := configs.write("alice", map[string]any{"scriptDefaults": map[string]any{"videoModelKey": "seedance-2-0-official", "videoResolution": "1080p"}}); err != nil {
		t.Fatal(err)
	}
	service := &fakeScreenplayVideoService{}
	runner := &screenplayVideoRunner{service: service, configs: configs, poll: time.Millisecond, repo: screenplay.NewMemoryProductionRepository()}
	artifacts, err := runner.ProduceEpisode(context.Background(), screenplay.Project{ID: "project-1", OwnerUsername: "alice", AspectRatio: "9:16"}, 1, screenplay.Storyboard{Shots: []screenplay.StoryboardShot{{Seq: 1, Duration: 12, VideoDesc: "画风: 漫剧\n雨夜推进"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(artifacts) != 1 || artifacts[0] != "https://media.example/shot.mp4" {
		t.Fatalf("artifacts=%v", artifacts)
	}
	if len(service.inputs) != 1 || service.inputs[0].ModelKey != "seedance-2-0-official" || service.inputs[0].Resolution != "1080p" || service.inputs[0].AspectRatio != "9:16" {
		t.Fatalf("input=%+v", service.inputs)
	}
}

type transientScreenplayVideoService struct {
	fakeScreenplayVideoService
	polls int
}

func (f *transientScreenplayVideoService) Poll(_ context.Context, _ string, id string) (scriptVideoStatus, error) {
	f.polls++
	if f.polls == 1 {
		return scriptVideoStatus{}, fmt.Errorf("temporary network outage")
	}
	return scriptVideoStatus{TaskID: id, Status: "succeeded", VideoURL: "https://media.example/resumed.mp4"}, nil
}
func TestScreenplayVideoRunnerReusesPersistedTaskAfterPollFailure(t *testing.T) {
	ctx := context.Background()
	configs := NewUserConfigStore(t.TempDir())
	_ = configs.write("alice", map[string]any{"scriptDefaults": map[string]any{"videoModelKey": "seedance-2-0-official"}})
	repo := screenplay.NewMemoryProductionRepository()
	service := &transientScreenplayVideoService{}
	p := screenplay.Project{ID: "p", OwnerUsername: "alice"}
	board := screenplay.Storyboard{ID: "b", Shots: []screenplay.StoryboardShot{{ID: "s", Seq: 1, Duration: 5, VideoDesc: "雨夜镜头"}}}
	r := &screenplayVideoRunner{service: service, configs: configs, repo: repo, References: func(context.Context, screenplay.Project, screenplay.StoryboardShot) ([]string, error) {
		return []string{"https://media.example/reference.png"}, nil
	}}
	if _, e := r.ProduceEpisode(ctx, p, 1, board); e == nil {
		t.Fatal("poll failure hidden")
	}
	record, e := repo.GetShotTask(ctx, "alice", "p", "b", 1)
	if e != nil || record.TaskID != "task-1" {
		t.Fatalf("task not durable: %+v %v", record, e)
	}
	r = &screenplayVideoRunner{service: service, configs: configs, repo: repo, References: r.References}
	got, e := r.ProduceEpisode(ctx, p, 1, board)
	if e != nil || len(got) != 1 || len(service.inputs) != 1 || len(service.inputs[0].ImageURLs) != 1 {
		t.Fatalf("got=%v err=%v submits=%v", got, e, service.inputs)
	}
	if _, e = r.ProduceEpisode(ctx, p, 1, board); e != nil || service.polls != 2 {
		t.Fatalf("completed clip polled again: %d %v", service.polls, e)
	}
}
func TestScreenplayVideoRunnerStopsUnknownSubmissionAndPausedJob(t *testing.T) {
	ctx := context.Background()
	configs := NewUserConfigStore(t.TempDir())
	_ = configs.write("alice", map[string]any{"scriptDefaults": map[string]any{"videoModelKey": "seedance-2-0-official"}})
	repo := screenplay.NewMemoryProductionRepository()
	_ = repo.SetProductionState(ctx, "alice", "p", "paused", "")
	service := &fakeScreenplayVideoService{}
	r := &screenplayVideoRunner{service: service, configs: configs, repo: repo}
	_, e := r.ProduceEpisode(ctx, screenplay.Project{ID: "p", OwnerUsername: "alice"}, 1, screenplay.Storyboard{ID: "b", Shots: []screenplay.StoryboardShot{{Seq: 1, VideoDesc: "镜头"}}})
	if e == nil || len(service.inputs) > 0 {
		t.Fatalf("paused job submitted: %v", e)
	}
}
func TestScreenplayVideoRunnerExplicitChangedRebuildPreservesHistory(t *testing.T) {
	ctx := context.Background()
	configs := NewUserConfigStore(t.TempDir())
	_ = configs.write("alice", map[string]any{"scriptDefaults": map[string]any{"videoModelKey": "seedance-2-0-official"}})
	repo := screenplay.NewMemoryProductionRepository()
	_ = repo.SetProductionState(ctx, "alice", "p", "running", "")
	service := &fakeScreenplayVideoService{}
	r := &screenplayVideoRunner{service: service, configs: configs, repo: repo}
	p := screenplay.Project{ID: "p", OwnerUsername: "alice"}
	b := screenplay.Storyboard{ID: "b", Shots: []screenplay.StoryboardShot{{ID: "s", Seq: 1, Duration: 5, VideoDesc: "旧镜头"}}}
	if _, e := r.ProduceEpisode(ctx, p, 1, b); e != nil {
		t.Fatal(e)
	}
	b.Shots[0].VideoDesc = "修改镜头"
	if _, e := r.ProduceEpisode(ctx, p, 1, b); e == nil {
		t.Fatal("changed request silently submitted")
	}
	if len(service.inputs) != 1 {
		t.Fatal("spent quota without rebuild")
	}
	_ = repo.SetProductionRebuild(ctx, "alice", "p", true)
	if _, e := r.ProduceEpisode(ctx, p, 1, b); e != nil {
		t.Fatal(e)
	}
	tasks, _ := repo.ListShotTasks(ctx, "alice", "p")
	if len(tasks) != 2 || len(service.inputs) != 2 || service.inputs[0].SourceTaskID == service.inputs[1].SourceTaskID {
		t.Fatalf("history=%+v inputs=%+v", tasks, service.inputs)
	}
	b.Shots[0].VideoDesc = "旧镜头"
	if _, e := r.ProduceEpisode(ctx, p, 1, b); e != nil || len(service.inputs) != 2 {
		t.Fatalf("old request version not reused: %v", e)
	}
}
func TestScreenplayVideoRunnerStableReferencesSurviveSigningRefresh(t *testing.T) {
	ctx := context.Background()
	configs := NewUserConfigStore(t.TempDir())
	_ = configs.write("alice", map[string]any{"scriptDefaults": map[string]any{"videoModelKey": "seedance-2-0-official"}})
	repo := screenplay.NewMemoryProductionRepository()
	service := &transientScreenplayVideoService{}
	count := 0
	r := &screenplayVideoRunner{service: service, configs: configs, repo: repo, References: func(context.Context, screenplay.Project, screenplay.StoryboardShot) ([]string, error) {
		count++
		return []string{fmt.Sprintf("https://media.example/a.png?signature=%d", count)}, nil
	}, ReferenceIDs: func(context.Context, screenplay.Project, screenplay.StoryboardShot) ([]string, error) {
		return []string{"tos:stable-object-key"}, nil
	}}
	p := screenplay.Project{ID: "p", OwnerUsername: "alice"}
	b := screenplay.Storyboard{ID: "b", Shots: []screenplay.StoryboardShot{{Seq: 1, Duration: 5, VideoDesc: "镜头"}}}
	if _, e := r.ProduceEpisode(ctx, p, 1, b); e == nil {
		t.Fatal("expected transient error")
	}
	if _, e := r.ProduceEpisode(ctx, p, 1, b); e != nil || len(service.inputs) != 1 {
		t.Fatalf("signed URL caused duplicate: %v %+v", e, service.inputs)
	}
}

type unknownSubmitScreenplayService struct{ calls int }

func (s *unknownSubmitScreenplayService) Submit(context.Context, string, scriptVideoLocalInput) (scriptVideoSubmission, error) {
	s.calls++
	return scriptVideoSubmission{}, fmt.Errorf("connection lost after sending")
}
func (s *unknownSubmitScreenplayService) Poll(context.Context, string, string) (scriptVideoStatus, error) {
	return scriptVideoStatus{}, fmt.Errorf("must not poll unknown")
}
func TestScreenplayVideoRunnerUnknownSubmissionDoesNotChargeAgain(t *testing.T) {
	ctx := context.Background()
	configs := NewUserConfigStore(t.TempDir())
	_ = configs.write("alice", map[string]any{"scriptDefaults": map[string]any{"videoModelKey": "seedance-2-0-official"}})
	s := &unknownSubmitScreenplayService{}
	r := &screenplayVideoRunner{service: s, configs: configs, repo: screenplay.NewMemoryProductionRepository()}
	p := screenplay.Project{ID: "p", OwnerUsername: "alice"}
	b := screenplay.Storyboard{ID: "b", Shots: []screenplay.StoryboardShot{{Seq: 1, Duration: 5, VideoDesc: "镜头"}}}
	for i := 0; i < 2; i++ {
		if _, e := r.ProduceEpisode(ctx, p, 1, b); e == nil {
			t.Fatal("unknown accepted")
		}
	}
	if s.calls != 1 {
		t.Fatalf("duplicate submits %d", s.calls)
	}
}
