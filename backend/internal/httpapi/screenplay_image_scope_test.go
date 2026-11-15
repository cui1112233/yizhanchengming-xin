package httpapi

import (
	"context"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
)

func TestScreenplayCharacterLayoutOnlyAppliesToCharacterAsset(t *testing.T) {
	ctx := context.Background()
	store := screenplay.NewMemoryStore()
	p, err := store.CreateProject(ctx, screenplay.Project{ID: "image-scope", OwnerUsername: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	core := "角色专用：面部特写与全身三视图，纯白背景"
	p.CreativeConfig = &screenplay.CreativeConfig{ImageCoreInstruction: core, Style: &screenplay.CreativeStyleSnapshot{Prompt: "统一水墨画风"}}
	for _, kind := range []screenplay.AssetType{screenplay.AssetCharacter, screenplay.AssetScene, screenplay.AssetProp} {
		_, err = store.CreateAsset(ctx, screenplay.Asset{ID: string(kind), OwnerUsername: "alice", ProjectID: p.ID, Type: kind, Name: string(kind), Description: "该目标的独立描述"})
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err = store.SaveStoryboard(ctx, screenplay.Storyboard{OwnerUsername: "alice", ProjectID: p.ID, Engine: screenplay.EngineXiaoluoV9, EpisodeNo: 1, Shots: []screenplay.StoryboardShot{{ID: "shot", Seq: 1, Duration: 3, ImagePrompt: "书房中的完整单幅镜头"}}})
	if err != nil {
		t.Fatal(err)
	}
	s := &ScreenplayMediaWorkbench{Projects: store}
	for _, target := range []string{"character", "scene", "prop", "shot", ""} {
		t.Run(target, func(t *testing.T) {
			text, err := s.creativeImagePrompt(ctx, p, target, "保留本次构图要求")
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(text, core) != (target == "character") {
				t.Fatalf("character layout leaked to target %q: %s", target, text)
			}
			if !strings.Contains(text, "统一水墨画风") || !strings.Contains(text, "保留本次构图要求") {
				t.Fatal("lost shared style or explicit request")
			}
			if target == "shot" && !strings.Contains(text, "书房中的完整单幅镜头") {
				t.Fatal("lost shot prompt")
			}
		})
	}
}
