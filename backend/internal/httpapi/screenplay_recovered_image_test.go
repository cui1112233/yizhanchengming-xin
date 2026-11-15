package httpapi

import (
	"context"
	"errors"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
)

func TestScreenplay139MissingImagePromptNeverUsesVideoInstructions(t *testing.T) {
	ctx := context.Background()
	store := screenplay.NewMemoryStore()
	p, err := store.CreateProject(ctx, screenplay.Project{ID: "optional-image-contract", OwnerUsername: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.SaveStoryboard(ctx, screenplay.Storyboard{ProjectID: p.ID, OwnerUsername: "alice", EpisodeNo: 1, Engine: "prompt_139", Shots: []screenplay.StoryboardShot{{ID: "optional-image", Seq: 1, Duration: 15, VideoDesc: "视频运镜：甲说你好，镜头推进"}}})
	if err != nil {
		t.Fatal(err)
	}
	s := &ScreenplayMediaWorkbench{Projects: store}
	p.StyleDesc = "只有画风不能代替缺少的图片提示词"
	text, err := s.creativeImagePrompt(ctx, p, "optional-image", "")
	if !errors.Is(err, screenplay.ErrInvalidArgument) || !strings.Contains(err.Error(), "图片提示词") || text != "" {
		t.Fatal("missing image instructions became billable video fallback", text, err)
	}
	text, err = s.creativeImagePrompt(ctx, p, "optional-image", "用户手填静帧：教室走廊")
	if err != nil || !strings.Contains(text, "用户手填静帧：教室走廊") || strings.Contains(text, "视频运镜") {
		t.Fatal(text, err)
	}
}

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
