package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/screenplay"
	"testing"
)

func TestScreenplayAssetLibraryConsumesRealUploadedAndGeneratedMediaItems(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	project, err := projects.CreateProject(ctx, screenplay.Project{ID: "library-media-source", OwnerUsername: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	service := &ScreenplayMediaWorkbench{Projects: projects, Media: screenplay.NewMediaMemoryStore(), Cut: aicut.NewMemoryStore(), Objects: &memoryImageToolObjects{files: map[string][]byte{}}, Assets: agentworkspace.NewMemoryAssetStore()}
	uploaded, err := service.saveBytes(ctx, project, "image", "source.png", "image/png", bytes.NewReader(testImageToolPNG(t)), "")
	if err != nil {
		t.Fatal(err)
	}
	// Use an in-process image fixture so the real completion/storage path runs
	// without contacting any provider or spending model quota.
	service.Images = func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
		return []string{"data:image/png;base64," + base64.StdEncoding.EncodeToString(testImageToolPNG(t))}, 200, nil
	}
	if _, err = service.action(ctx, project, screenplayMediaAction{Action: "image", ModelID: "fixture-model", Prompt: "fixture"}); err != nil {
		t.Fatal(err)
	}
	state, err := service.Media.LoadMedia(ctx, "alice", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	repository := screenplay.NewMemoryAssetLibraryRepository()
	count := 0
	for _, item := range state.Items {
		if item.Kind != "image" || item.Status != "succeeded" || item.AssetID == "" {
			continue
		}
		count++
		saved, err := repository.SaveAssetLibraryEntry(ctx, screenplay.AssetLibraryEntry{ID: fmt.Sprintf("library-real-%d", count), Owner: "alice", Type: screenplay.AssetCharacter, Name: "人物", Image: item, SourceProjectID: project.ID}, 0)
		if err != nil {
			t.Fatal("real successful CM media rejected", item, err)
		}
		if saved.Image.Status != "succeeded" || saved.Image.AssetID != item.AssetID {
			t.Fatal(saved)
		}
		if _, err = service.Cut.Asset(ctx, "alice", project.ID, saved.Image.AssetID); err != nil {
			t.Fatal("library lost real media asset reference", err)
		}
	}
	if count != 2 || uploaded.Status != "succeeded" {
		t.Fatal("upload and generation paths were not both exercised", count, uploaded)
	}
}
