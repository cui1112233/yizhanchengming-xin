package httpapi

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
)

func TestScreenplayCreativeImageSendsSelectedDesignAndAsset(t *testing.T) {
	ctx := context.Background()
	store := screenplay.NewMemoryStore()
	p, _ := store.CreateProject(ctx, screenplay.Project{ID: "creative-image", OwnerUsername: "alice", StyleDesc: "旧画风"})
	p.CreativeConfig = &screenplay.CreativeConfig{CharacterDesignPromptID: prompt.CreativeCharacter213, SceneDesignPromptID: prompt.CreativeScene289, ImageCoreInstruction: "保留衣服纹样", Style: &screenplay.CreativeStyleSnapshot{Prompt: "国风水墨"}}
	_, err := store.CreateAsset(ctx, screenplay.Asset{ID: "hero", ProjectID: p.ID, OwnerUsername: "alice", Type: screenplay.AssetCharacter, Name: "林舟", Description: "穿青色长袍", Prompt: "黑色长发"})
	if err != nil {
		t.Fatal(err)
	}
	s := &ScreenplayMediaWorkbench{Projects: store, Media: screenplay.NewMediaMemoryStore(), Cut: aicut.NewMemoryStore(), Objects: &memoryImageToolObjects{files: map[string][]byte{}}, Assets: agentworkspace.NewMemoryAssetStore()}
	calls := 0
	s.Images = func(_ context.Context, _ string, in AgentImageGenerationInput) ([]string, int, error) {
		calls++
		for _, want := range []string{"林舟", "青色长袍", "黑色长发", "国风水墨", "保留衣服纹样", "侧面清楚"} {
			if !strings.Contains(in.Prompt, want) {
				t.Fatalf("missing %q in assembled input", want)
			}
		}
		if strings.Contains(in.Prompt, "{stylePrompt}") || strings.Contains(in.Prompt, "旧画风") {
			t.Fatal("unresolved or stale style")
		}
		return []string{"data:image/png;base64," + base64.StdEncoding.EncodeToString(testImageToolPNG(t))}, 200, nil
	}
	_, err = s.generate(ctx, p, screenplayMediaAction{Action: "image", TargetID: "hero", ModelID: "test", Prompt: "侧面清楚", IdempotencyKey: "one"}, screenplay.MediaState{})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal(calls)
	}
	state, _ := s.Media.LoadMedia(ctx, "alice", p.ID)
	_, err = s.generate(ctx, p, screenplayMediaAction{Action: "image", TargetID: "hero", ModelID: "test", Prompt: "侧面清楚", IdempotencyKey: "one"}, state)
	if err != nil || calls != 1 {
		t.Fatalf("duplicate submitted: calls=%d error=%v", calls, err)
	}
	p.CreativeConfig.CharacterDesignPromptID = prompt.CreativeScene289
	_, err = s.generate(ctx, p, screenplayMediaAction{Action: "image", TargetID: "hero", ModelID: "test"}, state)
	if err == nil || calls != 1 {
		t.Fatalf("wrong-purpose template submitted: %v", err)
	}
}

func TestScreenplayCreativeSceneDoesNotNeedManualPrompt(t *testing.T) {
	ctx := context.Background()
	store := screenplay.NewMemoryStore()
	p, _ := store.CreateProject(ctx, screenplay.Project{ID: "scene", OwnerUsername: "alice"})
	p.CreativeConfig = &screenplay.CreativeConfig{SceneDesignPromptID: prompt.CreativeScene289}
	_, err := store.CreateAsset(ctx, screenplay.Asset{ID: "room", ProjectID: p.ID, OwnerUsername: "alice", Type: screenplay.AssetScene, Name: "书房", Description: "木窗与书架"})
	if err != nil {
		t.Fatal(err)
	}
	s := &ScreenplayMediaWorkbench{Projects: store}
	text, err := s.creativeImagePrompt(ctx, p, "room", "")
	if err != nil || !strings.Contains(text, "书房") || !strings.Contains(text, "木窗与书架") || strings.Contains(text, "{sceneName}") {
		t.Fatalf("scene pairing failed: %v", err)
	}
	text, err = s.creativeImagePrompt(ctx, p, "", "")
	if err == nil || text != "" {
		t.Fatal("empty image request accepted")
	}
}
