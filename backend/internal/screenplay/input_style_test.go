package screenplay

import (
	"context"
	"strings"
	"testing"
)

func TestSelectedStyleDescriptionReachesAssetAndStoryboardRequests(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	_, err := store.CreateProject(ctx, Project{ID: "style", OwnerUsername: "alice", ArtStyle: "显示名称", StyleDesc: "后端完整画风描述", SelectedStyle: "preset-id", SourceText: "原文", PermissionMode: PermissionBypass})
	if err != nil {
		t.Fatal(err)
	}
	gen := &fakeGenerator{responses: []string{charactersJSON, `{"scenes":[{"name":"客厅","prompt":"场景描述"}]}`, `{"props":[{"name":"杯子","prompt":"道具描述"}]}`, storyboardJSON, storyboardJSON}}
	for _, kind := range []AssetType{AssetCharacter, AssetScene, AssetProp} {
		_, err = (&AssetService{Store: store, Generator: gen}).Extract(ctx, AssetRequest{Owner: "alice", ProjectID: "style", Type: kind, Degree: DegreeMaster, Content: "第一章 正文"})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, engine := range []StoryboardEngine{EngineXiaoluoV7, EngineXiaoyunqueV3} {
		_, err = (&StoryboardService{Store: store, Generator: gen}).Generate(ctx, StoryboardRequest{IdempotencyKey: "test-inputstyletestgo-" + string(engine), Owner: "alice", ProjectID: "style", Engine: engine, Content: "# 第1集：测试\n正文", Style: "显示名称"})
		if err != nil {
			t.Fatal(err)
		}
	}
	for i, req := range gen.requests {
		if !strings.Contains(req.UserPrompt+req.SystemPrompt, "后端完整画风描述") {
			t.Errorf("request %d missing resolved style", i)
		}
	}
}
