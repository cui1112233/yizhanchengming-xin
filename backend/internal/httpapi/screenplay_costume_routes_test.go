package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/screenplay"
)

func TestScreenplayRoutesCostumeChange(t *testing.T) {
	now := time.Date(2026, 11, 18, 11, 0, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, screenplay.Project{ID: "costume-project", OwnerUsername: "alice", Title: "换装"}); err != nil {
		t.Fatal(err)
	}
	for _, asset := range []screenplay.Asset{
		{ProjectID: "costume-project", OwnerUsername: "alice", Type: screenplay.AssetCharacter, Name: "张三"},
		{ProjectID: "costume-project", OwnerUsername: "alice", Type: screenplay.AssetCharacter, Name: "张三-睡衣"},
		{ProjectID: "costume-project", OwnerUsername: "alice", Type: screenplay.AssetScene, Name: "卧室"},
	} {
		if _, err := store.CreateAsset(ctx, asset); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := store.SaveStoryboard(ctx, screenplay.Storyboard{
		ProjectID: "costume-project", OwnerUsername: "alice", EpisodeNo: 1, Engine: screenplay.EngineXiaoluoV7,
		Shots: []screenplay.StoryboardShot{
			{ID: "cs-1", Seq: 1, Scene: "卧室", VideoDesc: "第一镜", Characters: "张三、小Y"},
			{ID: "cs-2", Seq: 2, Scene: "卧室", VideoDesc: "第二镜", Characters: "张三"},
			{ID: "cs-3", Seq: 3, Scene: "学校", VideoDesc: "第三镜", Characters: "张三、李四"},
			{ID: "cs-4", Seq: 4, Scene: "卧室", VideoDesc: "第四镜", Characters: "张三-睡衣"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, nil, func() time.Time { return now })
	base := "/api/comic-screenplays/projects/costume-project/storyboards/xiaoluo_v7_15s/1/costume"

	options := signedJSONRequest(t, mux, now, "alice", http.MethodGet, base+"/options", nil)
	if options.Code != http.StatusOK {
		t.Fatalf("options=%d %s", options.Code, options.Body.String())
	}
	var optionsPayload struct {
		Options screenplay.CostumeChangeOptionsView `json:"options"`
	}
	if err := json.Unmarshal(options.Body.Bytes(), &optionsPayload); err != nil {
		t.Fatal(err)
	}
	if !containsText(options.Body.String(), `"张三-睡衣"`) || !containsText(options.Body.String(), `"卧室"`) {
		t.Fatalf("options 缺少资产库人物或场景：%s", options.Body.String())
	}
	if optionsPayload.Options.ShotCount != 4 {
		t.Fatalf("shot count = %d", optionsPayload.Options.ShotCount)
	}

	preview := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/preview", map[string]any{
		"shot_start": 1, "shot_end": 4, "scene": "卧室", "from": "张三", "to": "张三-睡衣",
	})
	if preview.Code != http.StatusOK {
		t.Fatalf("preview=%d %s", preview.Code, preview.Body.String())
	}
	if !containsText(preview.Body.String(), `"matched":2`) {
		t.Fatalf("preview matched = %s", preview.Body.String())
	}
	if containsText(preview.Body.String(), `"storyboard"`) {
		t.Fatalf("预览不应写库：%s", preview.Body.String())
	}
	untouched, err := store.GetStoryboard(ctx, "alice", "costume-project", 1, screenplay.EngineXiaoluoV7)
	if err != nil {
		t.Fatal(err)
	}
	if untouched.Shots[0].Characters != "张三、小Y" {
		t.Fatalf("预览改动了分镜：%q", untouched.Shots[0].Characters)
	}

	applied := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/apply", map[string]any{
		"shot_start": 1, "shot_end": 4, "scene": "卧室", "from": "张三", "to": "张三-睡衣",
	})
	if applied.Code != http.StatusOK {
		t.Fatalf("apply=%d %s", applied.Code, applied.Body.String())
	}
	if !containsText(applied.Body.String(), "已在分镜 1-4（卧室）将「张三」换成「张三-睡衣」，共 2 个分镜") {
		t.Fatalf("apply message = %s", applied.Body.String())
	}
	stored, err := store.GetStoryboard(ctx, "alice", "costume-project", 1, screenplay.EngineXiaoluoV7)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Shots[0].Characters != "张三-睡衣、小Y" || stored.Shots[1].Characters != "张三-睡衣" {
		t.Fatalf("stored shots = %+v", stored.Shots[:2])
	}
	if stored.Shots[2].Characters != "张三、李四" || stored.Shots[3].Characters != "张三-睡衣" {
		t.Fatalf("未命中分镜被改动：%+v", stored.Shots)
	}

	emptyPreview := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/preview", map[string]any{
		"shot_start": 1, "shot_end": 4, "scene": "学校", "from": "王五", "to": "张三-睡衣",
	})
	if emptyPreview.Code != http.StatusOK || !containsText(emptyPreview.Body.String(), "没有可换装的分镜") {
		t.Fatalf("empty preview=%d %s", emptyPreview.Code, emptyPreview.Body.String())
	}

	sameName := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/apply", map[string]any{
		"shot_start": 1, "shot_end": 4, "from": "张三", "to": "张三",
	})
	if sameName.Code != http.StatusBadRequest || !containsText(sameName.Body.String(), "原人物和替换人物不能相同") {
		t.Fatalf("same name=%d %s", sameName.Code, sameName.Body.String())
	}

	noMatch := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/apply", map[string]any{
		"shot_start": 1, "shot_end": 4, "scene": "学校", "from": "王五", "to": "张三-睡衣",
	})
	if noMatch.Code == http.StatusOK || !containsText(noMatch.Body.String(), "没有可换装的分镜") {
		t.Fatalf("no match apply=%d %s", noMatch.Code, noMatch.Body.String())
	}

	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodPost, base+"/apply", map[string]any{
		"shot_start": 1, "shot_end": 4, "from": "张三", "to": "张三-睡衣",
	})
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign apply=%d %s", foreign.Code, foreign.Body.String())
	}
}
