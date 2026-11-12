package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"testing"
	"time"
)

func TestAICutCreativeCanvasScriptTargetsOnlyNamedTimeline(t *testing.T) {
	now := time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC)
	cuts := aicut.NewMemoryStore()
	projects := agentworkspace.NewMemoryStore()
	if err := cuts.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-named", Title: "命名时间线", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	collection := aicut.NewMemoryTimelineCollectionStore(cuts)
	named, err := collection.Create(t.Context(), "alice", "cut-named", "分镜版本")
	if err != nil {
		t.Fatal(err)
	}
	canvas := aicutCreativeSeed()
	canvas.Nodes[0].Summary = "只追加到命名时间线"
	task := aicutCreativeTaskID("alice", "cut-named")
	if err := projects.CreateProject(t.Context(), "alice", agentworkspace.Project{ID: task, Title: "创作", Canvas: canvas, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	scoped := aicutRequestTimelineStore{aicutExportTimelineStore{Store: cuts, timelines: collection}}
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	mux := http.NewServeMux()
	RegisterAICutCreativeCanvasRoutes(mux, auth, scoped, projects, nil, nil, func() time.Time { return now })
	before, err := collection.Timeline(t.Context(), "alice", "cut-named", named.ID)
	if err != nil {
		t.Fatal(err)
	}
	response := signedJSONRequestWithTimeline(t, withAICutTimelineScope(mux, auth), now, "POST", "/api/ai-cut/projects/cut-named/creative-canvas/script", named.ID, map[string]any{"nodeIds": []string{"creative-story"}, "expectedRevision": before.Revision})
	if response.Code != http.StatusOK {
		t.Fatalf("append=%d %s", response.Code, response.Body.String())
	}
	actual, _ := collection.Timeline(t.Context(), "alice", "cut-named", named.ID)
	defaultTimeline, _ := cuts.Timeline(t.Context(), "alice", "cut-named")
	if actual.Script != "只追加到命名时间线" || defaultTimeline.Script != "" {
		t.Fatalf("named=%q default=%q", actual.Script, defaultTimeline.Script)
	}
}

func TestAICutCreativeCanvasBindingAndTransfer(t *testing.T) {
	now := time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC)
	cuts := aicut.NewMemoryStore()
	projects := agentworkspace.NewMemoryStore()
	assets := agentworkspace.NewMemoryAssetStore()
	objects := agentworkspace.NewLocalAssetObjectStore(t.TempDir())
	if err := cuts.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-one", Title: "测试", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAICutCreativeCanvasRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, cuts, projects, assets, objects, func() time.Time { return now })
	endpoint := "/api/ai-cut/projects/cut-one/creative-canvas"
	bind := func() (string, agentworkspace.Canvas) {
		t.Helper()
		r := signedJSONRequest(t, mux, now, "alice", "POST", endpoint, nil)
		if r.Code != 200 {
			t.Fatalf("bind=%d %s", r.Code, r.Body.String())
		}
		var p struct {
			TaskID string                `json:"taskId"`
			Canvas agentworkspace.Canvas `json:"canvas"`
		}
		if err := json.Unmarshal(r.Body.Bytes(), &p); err != nil {
			t.Fatal(err)
		}
		return p.TaskID, p.Canvas
	}
	task, canvas := bind()
	if task == "" || len(canvas.Nodes) != 5 || len(canvas.Edges) < 4 {
		t.Fatalf("invalid seed %+v", canvas)
	}
	if err := agentworkspace.ValidateCanvasForStorage(canvas); err != nil {
		t.Fatal(err)
	}
	canvas.Nodes[0].Summary = "用户编辑"
	canvas.Nodes = append(canvas.Nodes, agentworkspace.Node{ID: "generated-image", Type: "image", Title: "生成结果", Version: 1, Status: "confirmed", AssetRef: map[string]any{"id": "generated-asset", "kind": "image"}})
	if _, err := projects.SaveCanvas(t.Context(), "alice", task, canvas, canvas.Revision, now); err != nil {
		t.Fatal(err)
	}
	again, saved := bind()
	if task != again || saved.Nodes[0].Summary != "用户编辑" {
		t.Fatal("binding overwrote saved canvas")
	}
	if got := saved.Nodes[len(saved.Nodes)-1].AssetRef["url"]; got != agentAssetContentURL(task, "generated-asset") {
		t.Fatalf("rebound generated image preview URL=%v", got)
	}
	if r := signedJSONRequest(t, mux, now, "bob", "POST", endpoint, nil); r.Code != 404 {
		t.Fatalf("foreign bind=%d", r.Code)
	}
	data := opaquePNG(t)
	source := agentworkspace.Asset{ID: "source-one", ProjectID: task, Kind: "image", Name: "cover.png", ContentType: "image/png", ObjectKey: "agent-assets/test/source.png", ByteSize: int64(len(data)), CreatedAt: now}
	if err := assets.CreateAsset(t.Context(), "alice", source); err != nil {
		t.Fatal(err)
	}
	if err := objects.Put(t.Context(), source, bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	for _, ids := range [][]string{nil, {"source-one", "missing"}} {
		r := signedJSONRequest(t, mux, now, "alice", "POST", endpoint+"/assets", map[string]any{"assetIds": ids})
		if r.Code < 400 {
			t.Fatalf("invalid source accepted %s", r.Body.String())
		}
		list, _ := cuts.ListAssets(t.Context(), "alice", "cut-one")
		if len(list) != 0 {
			t.Fatal("partially imported invalid batch")
		}
	}
	transfer := func() aicut.Asset {
		t.Helper()
		r := signedJSONRequest(t, mux, now, "alice", "POST", endpoint+"/assets", map[string]any{"assetIds": []string{"source-one", "source-one"}})
		if r.Code != 200 {
			t.Fatalf("transfer=%d %s", r.Code, r.Body.String())
		}
		var p struct {
			Assets []aicut.Asset `json:"assets"`
		}
		_ = json.Unmarshal(r.Body.Bytes(), &p)
		if len(p.Assets) != 1 {
			t.Fatalf("assets=%+v", p.Assets)
		}
		return p.Assets[0]
	}
	first := transfer()
	second := transfer()
	if first.ID != second.ID {
		t.Fatal("unstable target")
	}
	stored, err := cuts.Asset(t.Context(), "alice", "cut-one", first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.ObjectKey == source.ObjectKey {
		t.Fatal("shared object reference")
	}
	if err := objects.Delete(t.Context(), source); err != nil {
		t.Fatal(err)
	}
	body, err := objects.Open(t.Context(), aicutObjectAsset(stored))
	if err != nil {
		t.Fatal(err)
	}
	copied, _ := io.ReadAll(body)
	body.Close()
	if !bytes.Equal(data, copied) {
		t.Fatal("copied media differs")
	}
	list, _ := cuts.ListAssets(t.Context(), "alice", "cut-one")
	if len(list) != 1 {
		t.Fatal("duplicate rows")
	}
	timeline, _ := cuts.Timeline(t.Context(), "alice", "cut-one")
	if timeline.Revision != 0 {
		t.Fatal("changed timeline")
	}
}

func TestAICutCreativeCanvasScriptAppendPreservesTimeline(t *testing.T) {
	now := time.Now().UTC()
	cuts := aicut.NewMemoryStore()
	projects := agentworkspace.NewMemoryStore()
	if err := cuts.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-script", Title: "剧本", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	task := aicutCreativeTaskID("alice", "cut-script")
	canvas := aicutCreativeSeed()
	canvas.Nodes[0].Summary = "新故事"
	canvas.Nodes[4].Config["sourceText"] = "新分镜"
	if err := projects.CreateProject(t.Context(), "alice", agentworkspace.Project{ID: task, Title: "创作", Canvas: canvas, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	timeline, err := cuts.Timeline(t.Context(), "alice", "cut-script")
	if err != nil {
		t.Fatal(err)
	}
	timeline.Script = "原剧本"
	timeline.Markers = []aicut.TimelineMarker{{ID: "mark", TimeMS: 200, Label: "保留"}}
	timeline, err = cuts.SaveTimeline(t.Context(), "alice", "cut-script", timeline, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAICutCreativeCanvasRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, cuts, projects, nil, nil, func() time.Time { return now })
	endpoint := "/api/ai-cut/projects/cut-script/creative-canvas/script"
	r := signedJSONRequest(t, mux, now, "alice", "POST", endpoint, map[string]any{"nodeIds": []string{"creative-story", "creative-storyboard"}, "expectedRevision": timeline.Revision})
	if r.Code != 200 {
		t.Fatalf("append=%d %s", r.Code, r.Body.String())
	}
	saved, err := cuts.Timeline(t.Context(), "alice", "cut-script")
	if err != nil {
		t.Fatal(err)
	}
	if saved.Script != "原剧本\n\n新故事\n\n新分镜" || len(saved.Markers) != 1 || saved.Markers[0].Label != "保留" || len(saved.Tracks) != len(timeline.Tracks) {
		t.Fatalf("saved=%+v", saved)
	}
	stale := signedJSONRequest(t, mux, now, "alice", "POST", endpoint, map[string]any{"nodeIds": []string{"creative-story"}, "expectedRevision": timeline.Revision})
	if stale.Code != 409 {
		t.Fatalf("stale=%d %s", stale.Code, stale.Body.String())
	}
	for _, ids := range [][]string{{"creative-role"}, {"foreign-node"}} {
		bad := signedJSONRequest(t, mux, now, "alice", "POST", endpoint, map[string]any{"nodeIds": ids, "expectedRevision": saved.Revision})
		if bad.Code != 400 {
			t.Fatalf("bad=%d %s", bad.Code, bad.Body.String())
		}
	}
	unchanged, _ := cuts.Timeline(t.Context(), "alice", "cut-script")
	if unchanged.Revision != saved.Revision {
		t.Fatal("invalid append changed timeline")
	}
}

func TestAICutCreativeCanvasRejectsForeignAndMissingObjectsBeforeCopy(t *testing.T) {
	now := time.Now().UTC()
	cuts := aicut.NewMemoryStore()
	projects := agentworkspace.NewMemoryStore()
	assets := agentworkspace.NewMemoryAssetStore()
	objects := agentworkspace.NewLocalAssetObjectStore(t.TempDir())
	if err := cuts.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-scope", Title: "素材", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	task := aicutCreativeTaskID("alice", "cut-scope")
	if err := projects.CreateProject(t.Context(), "alice", agentworkspace.Project{ID: task, Title: "创作", Canvas: aicutCreativeSeed(), CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	data := opaquePNG(t)
	for _, source := range []agentworkspace.Asset{{ID: "good", ProjectID: task, Kind: "image", Name: "good.png", ContentType: "image/png", ObjectKey: "source/good.png", ByteSize: int64(len(data)), CreatedAt: now}, {ID: "missing-body", ProjectID: task, Kind: "image", Name: "missing.png", ContentType: "image/png", ObjectKey: "source/missing.png", ByteSize: 12, CreatedAt: now}, {ID: "foreign", ProjectID: "other-project", Kind: "image", Name: "other.png", ContentType: "image/png", ObjectKey: "source/other.png", ByteSize: 12, CreatedAt: now}} {
		if err := assets.CreateAsset(t.Context(), "alice", source); err != nil {
			t.Fatal(err)
		}
		if source.ID == "good" {
			if err := objects.Put(t.Context(), source, bytes.NewReader(data)); err != nil {
				t.Fatal(err)
			}
		}
	}
	mux := http.NewServeMux()
	RegisterAICutCreativeCanvasRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, cuts, projects, assets, objects, func() time.Time { return now })
	endpoint := "/api/ai-cut/projects/cut-scope/creative-canvas/assets"
	tooMany := make([]string, 31)
	for i := range tooMany {
		tooMany[i] = "good"
	}
	for _, ids := range [][]string{{"good", "foreign"}, {"good", "missing-body"}, tooMany, {" "}} {
		bad := signedJSONRequest(t, mux, now, "alice", "POST", endpoint, map[string]any{"assetIds": ids})
		if bad.Code < 400 {
			t.Fatalf("accepted %v", ids)
		}
		stored, _ := cuts.ListAssets(t.Context(), "alice", "cut-scope")
		if len(stored) != 0 {
			t.Fatal("prevalidation left partial destination records")
		}
	}
}

func TestAICutCreativeCanvasConcurrentBindingAndTransfer(t *testing.T) {
	now := time.Now().UTC()
	cuts := aicut.NewMemoryStore()
	projects := agentworkspace.NewMemoryStore()
	assets := agentworkspace.NewMemoryAssetStore()
	objects := agentworkspace.NewLocalAssetObjectStore(t.TempDir())
	if err := cuts.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-parallel", Title: "并发", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAICutCreativeCanvasRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, cuts, projects, assets, objects, func() time.Time { return now })
	endpoint := "/api/ai-cut/projects/cut-parallel/creative-canvas"
	results := make(chan int, 8)
	for range 8 {
		go func() { r := signedJSONRequest(t, mux, now, "alice", "POST", endpoint, nil); results <- r.Code }()
	}
	for range 8 {
		if code := <-results; code != 200 {
			t.Fatalf("concurrent bind=%d", code)
		}
	}
	list, err := projects.ListProjects(t.Context(), "alice")
	if err != nil || len(list) != 1 {
		t.Fatalf("projects=%v err=%v", list, err)
	}
	data := opaquePNG(t)
	source := agentworkspace.Asset{ID: "parallel", ProjectID: list[0].ID, Kind: "image", Name: "image.png", ContentType: "image/png", ObjectKey: "source/parallel.png", ByteSize: int64(len(data)), CreatedAt: now}
	if err := assets.CreateAsset(t.Context(), "alice", source); err != nil {
		t.Fatal(err)
	}
	if err := objects.Put(t.Context(), source, bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	for range 8 {
		go func() {
			r := signedJSONRequest(t, mux, now, "alice", "POST", endpoint+"/assets", map[string]any{"assetIds": []string{"parallel"}})
			results <- r.Code
		}()
	}
	for range 8 {
		if code := <-results; code != 200 {
			t.Fatalf("concurrent transfer=%d", code)
		}
	}
	stored, err := cuts.ListAssets(t.Context(), "alice", "cut-parallel")
	if err != nil || len(stored) != 1 {
		t.Fatalf("assets=%v err=%v", stored, err)
	}
	body, err := objects.Open(t.Context(), aicutObjectAsset(stored[0]))
	if err != nil {
		t.Fatal(err)
	}
	defer body.Close()
	actual, _ := io.ReadAll(body)
	if !bytes.Equal(actual, data) {
		t.Fatal("concurrent copy damaged object")
	}
}
