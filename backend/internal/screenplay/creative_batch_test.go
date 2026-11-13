package screenplay

import (
	"context"
	"qiantie/backend/internal/prompt"
	"testing"
)

func TestCreativeBatchResumesSelectedStoryboardProtocol(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	_, err := store.CreateProject(ctx, Project{ID: "p", OwnerUsername: "o", CreativeConfig: &CreativeConfig{StoryboardPromptID: prompt.CreativeStoryboard307}})
	if err != nil {
		t.Fatal(err)
	}
	for _, engine := range []StoryboardEngine{EngineXiaoyunqueV3, EngineXiaoluoV9} {
		_, err = store.SaveStoryboard(ctx, Storyboard{ID: string(engine), ProjectID: "p", OwnerUsername: "o", Engine: engine, EpisodeNo: 1, Shots: []StoryboardShot{{ID: "s", Duration: 3, VideoDesc: "开门"}}})
		if err != nil {
			t.Fatal(err)
		}
	}
	got, ok := existingBatchStoryboard(ctx, store, "o", "p", 1)
	if !ok || got.Engine != EngineXiaoluoV9 {
		t.Fatal("production selected another protocol", got.Engine)
	}
}
