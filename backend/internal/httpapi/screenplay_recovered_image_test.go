package httpapi

import (
	"context"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
)

func TestScreenplayImageConsumesSavedImagePromptRatherThanVideoInstructions(t *testing.T) {
	ctx := context.Background()
	store := screenplay.NewMemoryStore()
	p, err := store.CreateProject(ctx, screenplay.Project{ID: "image-contract", OwnerUsername: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.SaveStoryboard(ctx, screenplay.Storyboard{ProjectID: p.ID, OwnerUsername: "alice", EpisodeNo: 1, Engine: screenplay.EngineXiaoluoV7, Shots: []screenplay.StoryboardShot{{ID: "image-shot", Seq: 1, Duration: 12, ImagePrompt: "静帧：甲站在窗前", VideoDesc: "视频：镜头推进十秒"}}})
	if err != nil {
		t.Fatal(err)
	}
	s := &ScreenplayMediaWorkbench{Projects: store}
	text, err := s.creativeImagePrompt(ctx, p, "image-shot", "")
	if err != nil || !strings.Contains(text, "静帧：甲站在窗前") || strings.Contains(text, "镜头推进十秒") {
		t.Fatal(text, err)
	}
}
