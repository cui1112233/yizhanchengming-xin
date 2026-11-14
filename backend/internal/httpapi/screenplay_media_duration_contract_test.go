package httpapi

import (
	"context"
	"qiantie/backend/internal/screenplay"
	"testing"
)

func TestScreenplayVideoRunnerValidatesAllDurationsBeforeAnySubmission(t *testing.T) {
	ctx := context.Background()
	configs := NewUserConfigStore(t.TempDir())
	_ = configs.write("alice", map[string]any{"scriptDefaults": map[string]any{"videoModelKey": "seedance-2-0-official"}})
	upstream := &fakeScreenplayVideoService{}
	runner := &screenplayVideoRunner{service: upstream, configs: configs, repo: screenplay.NewMemoryProductionRepository()}
	_, err := runner.ProduceEpisode(ctx, screenplay.Project{ID: "p", OwnerUsername: "alice"}, 1, screenplay.Storyboard{ID: "b", Shots: []screenplay.StoryboardShot{{Seq: 1, Duration: 12, VideoDesc: "第一镜头"}, {Seq: 2, VideoDesc: "模板未指定数值时长"}}})
	if err == nil || len(upstream.inputs) != 0 {
		t.Fatal("incomplete later shot caused partial paid submission", err, upstream.inputs)
	}
}
