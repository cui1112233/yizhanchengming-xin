package screenplay

import (
	"context"
	"fmt"
	"testing"
)

type fakeBatchMediaRunner struct {
	calls []int
}

type cancelledBatchMediaRunner struct{}

func (cancelledBatchMediaRunner) ProduceEpisode(ctx context.Context, _ Project, _ int, _ Storyboard) ([]string, error) {
	return nil, ctx.Err()
}

func (f *fakeBatchMediaRunner) ProduceEpisode(_ context.Context, _ Project, episode int, _ Storyboard) ([]string, error) {
	f.calls = append(f.calls, episode)
	return []string{fmt.Sprintf("video://episode/%d", episode)}, nil
}

func TestBatchExecutorRunsPersistedProjectToCompletion(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	project, err := store.CreateProject(ctx, Project{
		ID: "project-1", OwnerUsername: "alice", Title: "链路测试",
		ScriptMode: "b", AssetDegree: DegreeMaster, PermissionMode: PermissionBypass,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.SaveProjectSource(ctx, "alice", project.ID, "第一章\n雨夜相遇"); err != nil {
		t.Fatal(err)
	}
	document, err := store.SaveScreenplay(ctx, Document{
		ID: "document-1", ProjectID: project.ID, OwnerUsername: "alice", Source: "b",
		Content: "# 第1集\n## 场次1 雨夜\n她推门而入。", EpisodeCount: 1,
	}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.MergeAssets(ctx, "alice", project.ID, []Asset{{
		ID: "asset-1", ProjectID: project.ID, OwnerUsername: "alice", Type: AssetCharacter,
		Degree: DegreeMaster, Name: "女主", Description: "雨夜归来的少女",
	}, {
		ID: "asset-2", ProjectID: project.ID, OwnerUsername: "alice", Type: AssetScene,
		Degree: DegreeMaster, Name: "雨夜", Description: "雨中的旧宅",
	}, {
		ID: "asset-3", ProjectID: project.ID, OwnerUsername: "alice", Type: AssetProp,
		Degree: DegreeMaster, Name: "纸伞", Description: "女主携带的纸伞",
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err = store.SaveStoryboard(ctx, Storyboard{
		ID: "storyboard-1", ProjectID: project.ID, OwnerUsername: "alice", EpisodeNo: 1,
		Engine: EngineXiaoyunqueV3, Status: "done", ShotCount: 1,
		Shots: []StoryboardShot{{ID: "storyboard-shot-1", Seq: 1, Duration: 12, Scene: "雨夜", VideoDesc: "画风: 漫剧\n少女推门"}},
	}); err != nil {
		t.Fatal(err)
	}
	_ = document

	batch := NewBatch("batch-1", project.ID, "alice", "session-1", PermissionBypass)
	if err = store.CreateBatch(ctx, batch); err != nil {
		t.Fatal(err)
	}
	media := &fakeBatchMediaRunner{}
	executor := BatchExecutor{Store: store, Media: media}

	for attempts := 0; attempts < 16; attempts++ {
		result, runErr := executor.RunNext(ctx, "alice", project.ID)
		if runErr != nil {
			t.Fatalf("run next attempt %d: %v", attempts+1, runErr)
		}
		if result.Batch.Status == BatchCompleted {
			break
		}
	}

	completed, err := store.GetBatch(ctx, "alice", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != BatchCompleted {
		t.Fatalf("batch status = %s, want %s", completed.Status, BatchCompleted)
	}
	for _, def := range StepDefinitions {
		if step := completed.Step(def.Key); step == nil || step.Status != StepDone {
			t.Fatalf("step %s = %#v, want done", def.Key, step)
		}
	}
	if len(media.calls) != 1 || media.calls[0] != 1 {
		t.Fatalf("media calls = %v, want [1]", media.calls)
	}
}

func TestBatchExecutorFailsHonestlyWhenMediaRunnerIsMissing(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	project, err := store.CreateProject(ctx, Project{ID: "project-2", OwnerUsername: "alice", Title: "无视频通道", PermissionMode: PermissionBypass})
	if err != nil {
		t.Fatal(err)
	}
	batch := NewBatch("batch-2", project.ID, "alice", "session-2", PermissionBypass)
	for _, def := range StepDefinitions[:7] {
		batch.Steps[def.Key].Status = StepDone
	}
	batch.CurrentStep = "7_video"
	if err = store.CreateBatch(ctx, batch); err != nil {
		t.Fatal(err)
	}

	result, err := (BatchExecutor{Store: store}).RunNext(ctx, "alice", project.ID)
	if err == nil {
		t.Fatal("expected missing media runner error")
	}
	if result.Batch == nil || result.Batch.Step("7_video").Status != StepFailed {
		t.Fatalf("batch after failure = %#v", result.Batch)
	}
}

func TestBatchExecutorCancellationPausesInsteadOfReportingFalseFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := NewMemoryStore()
	project, err := store.CreateProject(context.Background(), Project{ID: "project-3", OwnerUsername: "alice", Title: "取消任务", PermissionMode: PermissionBypass})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.SaveStoryboard(context.Background(), Storyboard{ID: "storyboard-3", ProjectID: project.ID, OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoyunqueV3, Status: "done", ShotCount: 1, Shots: []StoryboardShot{{ID: "shot-3", Seq: 1, Duration: 10, VideoDesc: "画风: 漫剧\n推进"}}}); err != nil {
		t.Fatal(err)
	}
	batch := NewBatch("batch-3", project.ID, "alice", "session-3", PermissionBypass)
	for _, def := range StepDefinitions[:7] {
		batch.Steps[def.Key].Status = StepDone
	}
	batch.Steps["7_video"].TotalEpisodes = 1
	if err = store.CreateBatch(context.Background(), batch); err != nil {
		t.Fatal(err)
	}
	result, runErr := (BatchExecutor{Store: store, Media: cancelledBatchMediaRunner{}}).RunNext(ctx, "alice", project.ID)
	if runErr == nil || result.Batch == nil || result.Batch.Status != BatchPaused {
		t.Fatalf("result=%+v err=%v", result, runErr)
	}
	if result.Batch.Step("7_video").Status != StepPending {
		t.Fatalf("video step=%+v", result.Batch.Step("7_video"))
	}
}
