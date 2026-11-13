package screenplay

import (
	"context"
	"fmt"
	"testing"
)

type productionTestGenerator struct{}

func (productionTestGenerator) Generate(context.Context, TextRequest) (TextResponse, error) {
	return TextResponse{Content: `{"pass":true,"block_next_step":false,"summary":"已审核真实内容","issues":[]}`}, nil
}

type productionTestFinalizer struct{}

func (productionTestFinalizer) Finalize(context.Context, Project) ([]string, error) {
	return []string{"https://media.example/final.mp4"}, nil
}
func TestProductionMoreThan64UnitsResumeAfterRuntimeReplacement(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	p, _ := s.CreateProject(ctx, Project{ID: "many", OwnerUsername: "a", PermissionMode: PermissionBypass})
	b := NewBatch("b", p.ID, "a", "s", PermissionBypass)
	for _, d := range StepDefinitions {
		b.Step(d.Key).Status = StepDone
	}
	v := b.Step("7_video")
	v.Status = StepPending
	v.TotalEpisodes = 90
	b.Step("8_finalize").Status = StepPending
	b.Status = BatchRunning
	_ = s.CreateBatch(ctx, b)
	for n := 1; n <= 90; n++ {
		_, e := s.SaveStoryboard(ctx, Storyboard{ID: fmt.Sprint("board", n), ProjectID: p.ID, OwnerUsername: "a", EpisodeNo: n, Engine: EngineXiaoyunqueV3, ShotCount: 1, Shots: []StoryboardShot{{Seq: 1, VideoDesc: "真实镜头"}}})
		if e != nil {
			t.Fatal(e)
		}
	}
	repo := NewMemoryProductionRepository()
	_ = repo.SetProductionState(ctx, "a", p.ID, "running", "")
	media := &fakeBatchMediaRunner{}
	makeRuntime := func() *ProductionRuntime {
		return &ProductionRuntime{Store: s, Repository: repo, RedisDo: func(_ context.Context, args ...string) (any, error) {
			if args[0] == "SET" {
				return "OK", nil
			}
			return int64(1), nil
		}, Executor: func(string) BatchExecutor { return BatchExecutor{Store: s, Media: media} }}
	}
	r := makeRuntime()
	for n := 0; n < 45; n++ {
		j, _ := repo.GetProductionJob(ctx, "a", p.ID)
		if e := r.RunOne(ctx, j); e != nil {
			t.Fatal(e)
		}
	}
	r = makeRuntime()
	for n := 0; n < 46; n++ {
		j, _ := repo.GetProductionJob(ctx, "a", p.ID)
		e := r.RunOne(ctx, j)
		if e != nil && e != ErrClipsReady {
			t.Fatal(e)
		}
	}
	j, _ := repo.GetProductionJob(ctx, "a", p.ID)
	if j.Status != "clips_ready" || len(media.calls) != 90 || j.Units != 91 {
		t.Fatalf("job=%+v calls=%d", j, len(media.calls))
	}
}
func TestProductionRedisCompetitionDoesNotExecute(t *testing.T) {
	r := &ProductionRuntime{RedisDo: func(context.Context, ...string) (any, error) { return nil, nil }, Executor: func(string) BatchExecutor { t.Fatal("lost lease executed"); return BatchExecutor{} }}
	if e := r.RunOne(context.Background(), ProductionJob{Owner: "a", ProjectID: "p"}); e != nil {
		t.Fatal(e)
	}
}

type doctorResponseGenerator string

func (g doctorResponseGenerator) Generate(_ context.Context, r TextRequest) (TextResponse, error) {
	if r.SystemPrompt == "" || r.UserPrompt == "" {
		return TextResponse{}, fmt.Errorf("missing real input")
	}
	return TextResponse{Content: string(g)}, nil
}
func TestProductionDoctorBlocksAndRequiresExplicitVerdict(t *testing.T) {
	for _, body := range []string{`{"summary":"未给结论"}`, `{"pass":false,"block_next_step":false,"summary":"动机冲突","issues":[{"level":"error","target":"第一场","message":"前后矛盾","suggestion":"重写动机"}]}`} {
		e := BatchExecutor{Generator: doctorResponseGenerator(body)}
		out, err := e.doctor(context.Background(), Project{ID: "p"}, Unit{StepKey: "2_screenplay_doctor"}, "真实剧本")
		if body == `{"summary":"未给结论"}` {
			if err == nil {
				t.Fatal("missing verdict accepted")
			}
		} else if err != nil || out.review.Pass || !out.review.BlockNextStep || len(out.review.Issues) != 1 {
			t.Fatalf("out=%+v err=%v", out, err)
		}
	}
}
func TestProductionStaleArtifactsCannotSilentlyReuseOldContent(t *testing.T) {
	for _, key := range []string{"3_assets", "5_storyboard"} {
		t.Run(key, func(t *testing.T) {
			ctx := context.Background()
			s := NewMemoryStore()
			_, _ = s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "a"})
			_, _ = s.MergeAssets(ctx, "a", "p", []Asset{{ID: "a", ProjectID: "p", OwnerUsername: "a", Type: AssetCharacter, Name: "旧人物"}})
			_, _ = s.SaveStoryboard(ctx, Storyboard{ID: "b", ProjectID: "p", OwnerUsername: "a", EpisodeNo: 1, Engine: EngineXiaoyunqueV3, ShotCount: 1, Shots: []StoryboardShot{{Seq: 1, VideoDesc: "旧镜头"}}})
			b := NewBatch("batch", "p", "a", "s", PermissionBypass)
			for _, st := range b.Steps {
				st.Status = StepDone
			}
			b.Step(key).Status = StepStale
			if key == "5_storyboard" {
				b.Step(key).TotalEpisodes = 1
			}
			b.Status = BatchRunning
			_ = s.CreateBatch(ctx, b)
			result, e := (BatchExecutor{Store: s}).RunNext(ctx, "a", "p")
			if e == nil {
				t.Fatal("stale old content silently reused")
			}
			if result.Batch.Step(key).Status != StepStale {
				t.Fatal("retry lost regeneration intent")
			}
		})
	}
}
