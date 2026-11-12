package httpapi

import (
	"context"
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
	runner := &screenplayVideoRunner{service: service, configs: configs, poll: time.Millisecond}
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
