package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/screenplay"
)

type fakeLongVideoService struct {
	inputs    []scriptVideoLocalInput
	submitErr error
	polls     int
}

func (f *fakeLongVideoService) Submit(_ context.Context, _ string, in scriptVideoLocalInput) (scriptVideoSubmission, error) {
	f.inputs = append(f.inputs, in)
	if f.submitErr != nil {
		return scriptVideoSubmission{}, f.submitErr
	}
	return scriptVideoSubmission{TaskID: fmt.Sprintf("task-%d", len(f.inputs))}, nil
}

func (f *fakeLongVideoService) Poll(context.Context, string, string) (scriptVideoStatus, error) {
	f.polls++
	return scriptVideoStatus{Status: "failed", Error: "供应商离线"}, nil
}

func longVideoTestWorkbench(t *testing.T, video screenplayVideoService) (*ScreenplayMediaWorkbench, screenplay.Project) {
	t.Helper()
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	p, err := projects.CreateProject(ctx, screenplay.Project{ID: "long-video", OwnerUsername: "alice", Title: "片", AspectRatio: "9:16"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = projects.SaveStoryboard(ctx, screenplay.Storyboard{ID: "board", ProjectID: p.ID, OwnerUsername: "alice", EpisodeNo: 1, Engine: screenplay.EngineXiaoluoV9, Shots: []screenplay.StoryboardShot{
		{ID: "shot-1", Seq: 1, Duration: 15, VideoDesc: "雨夜推门", Characters: "小雨、阿岚", Scene: "旧宅", Props: "伞"},
		{ID: "shot-2", Seq: 2, Duration: 5, VideoDesc: "闪电照亮走廊", Characters: "阿岚", Scene: "走廊", Props: "灯"},
		{ID: "shot-3", Seq: 3, Duration: 9, VideoDesc: "阿岚回头", Scene: "走廊"},
		{ID: "shot-4", Seq: 4, Duration: 15, VideoDesc: "门外雨停", Scene: "门口"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	baseDir := t.TempDir()
	userDir := filepath.Join(baseDir, "alice")
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	config := map[string]any{"scriptDefaults": map[string]any{"videoModelKey": "seedance-2-0-official", "videoResolution": "720p"}}
	raw, _ := json.Marshal(config)
	if err := os.WriteFile(filepath.Join(userDir, "api-config.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	workbench := &ScreenplayMediaWorkbench{
		Projects: projects,
		Media:    screenplay.NewMediaMemoryStore(),
		Cut:      aicut.NewMemoryStore(),
		Objects:  &memoryImageToolObjects{files: map[string][]byte{}},
		Assets:   agentworkspace.NewMemoryAssetStore(),
		Configs:  NewUserConfigStore(baseDir),
		Video:    video,
	}
	return workbench, p
}

func TestLongVideoPlanGroupsByRecommendedDuration(t *testing.T) {
	ctx := context.Background()
	workbench, p := longVideoTestWorkbench(t, nil)
	view, err := workbench.planLongVideo(ctx, p, longVideoRequest{Episode: 1, TargetSeconds: 300})
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Sections) != 2 {
		t.Fatalf("sections=%d want 2", len(view.Sections))
	}
	if view.Sections[0].TotalSeconds != 29 || view.Sections[0].Label != "超长视频分镜*1" {
		t.Fatalf("first=%+v", view.Sections[0])
	}
	if view.Sections[0].Prompt != "雨夜推门\n闪电照亮走廊\n阿岚回头" {
		t.Fatalf("prompt=%q", view.Sections[0].Prompt)
	}
	if view.Sections[0].Characters != "小雨、阿岚" || view.Sections[0].Scene != "旧宅、走廊" {
		t.Fatalf("merged fields=%+v", view.Sections[0])
	}
	if view.Sections[0].ShotNumbers != "第1镜、第2镜、第3镜" {
		t.Fatalf("shot numbers=%q", view.Sections[0].ShotNumbers)
	}
	if view.TotalSeconds != 44 {
		t.Fatalf("total=%v", view.TotalSeconds)
	}
}

func TestLongVideoGenerateSubmitsMergedPromptWithSavedModel(t *testing.T) {
	ctx := context.Background()
	service := &fakeLongVideoService{submitErr: fmt.Errorf("quota exceeded")}
	workbench, p := longVideoTestWorkbench(t, service)
	if _, err := workbench.startLongVideo(ctx, p, longVideoRequest{Episode: 1, TargetSeconds: 300}); err != nil {
		t.Fatal(err)
	}
	// start 是异步的，等分组落到失败状态后再断言。
	deadline := time.Now().Add(5 * time.Second)
	var section screenplay.LongVideoSection
	for time.Now().Before(deadline) {
		batch, ok := workbench.longVideoBatch(ctx, p, 1)
		if ok && len(batch.Sections) == 2 && batch.Sections[0].Status == screenplay.LongVideoStatusFailed {
			section = batch.Sections[0]
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if section.Status != screenplay.LongVideoStatusFailed {
		t.Fatalf("section status=%q", section.Status)
	}
	if section.ProviderID != "" {
		t.Fatalf("failed submit must not record a provider task: %+v", section)
	}
	if len(service.inputs) != 2 {
		t.Fatalf("每组都要单独提交，submits=%d", len(service.inputs))
	}
	input := service.inputs[0]
	if input.ModelKey != "seedance-2-0-official" || input.Resolution != "720p" {
		t.Fatalf("input model=%+v", input)
	}
	duration, _ := input.Duration.(float64)
	if duration < 28.99 || duration > 29.01 {
		t.Fatalf("duration=%v want 29", input.Duration)
	}
	if input.Prompt != "雨夜推门\n闪电照亮走廊\n阿岚回头" {
		t.Fatalf("prompt=%q", input.Prompt)
	}
	if input.AspectRatio != "9:16" {
		t.Fatalf("aspect=%q", input.AspectRatio)
	}
}

func TestLongVideoCustomPromptIsKeptOnRestart(t *testing.T) {
	ctx := context.Background()
	workbench, p := longVideoTestWorkbench(t, &fakeLongVideoService{})
	first, err := workbench.prepareLongVideoBatch(ctx, p, longVideoRequest{Episode: 1, TargetSeconds: 300})
	if err != nil {
		t.Fatal(err)
	}
	edited := []screenplay.LongVideoSection{{Index: 1, Prompt: "本次改写后的第一组提示词"}}
	second, err := workbench.prepareLongVideoBatch(ctx, p, longVideoRequest{Episode: 1, TargetSeconds: 300, Sections: edited})
	if err != nil {
		t.Fatal(err)
	}
	if second.Sections[0].Prompt != "本次改写后的第一组提示词" {
		t.Fatalf("edited prompt lost: %+v", second.Sections[0])
	}
	if second.Sections[1].Prompt != first.Sections[1].Prompt {
		t.Fatalf("second section prompt changed: %q", second.Sections[1].Prompt)
	}
	if second.ID != first.ID {
		t.Fatal("same plan must keep the same batch id so results stay attached")
	}
}
