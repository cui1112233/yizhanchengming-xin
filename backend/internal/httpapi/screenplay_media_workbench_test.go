package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

func TestScreenplayMediaImageReferencesAndFailureHistory(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	p, _ := projects.CreateProject(ctx, screenplay.Project{ID: "media-p", OwnerUsername: "alice", Title: "片", AspectRatio: "9:16"})
	media := screenplay.NewMediaMemoryStore()
	objects := &memoryImageToolObjects{files: map[string][]byte{}}
	s := &ScreenplayMediaWorkbench{Projects: projects, Media: media, Cut: aicut.NewMemoryStore(), Objects: objects, Assets: agentworkspace.NewMemoryAssetStore()}
	s.ReferencesProvider = agentImageReferenceProvider(func(context.Context, agentworkspace.Asset, time.Duration) (string, error) {
		return "https://assets.example/ref.png", nil
	})
	ref, e := s.saveBytes(ctx, p, "image", "ref.png", "image/png", bytes.NewReader(testImageToolPNG(t)), "")
	if e != nil {
		t.Fatal(e)
	}
	calls := 0
	s.Images = func(_ context.Context, owner string, in AgentImageGenerationInput) ([]string, int, error) {
		calls++
		if owner != "alice" || in.ModelID != "mine" || len(in.ReferenceAssetIDs) != 1 || in.ReferenceAssetIDs[0] != ref.AssetID || len(in.ReferenceImages) != 1 || in.ReferenceImages[0] != "https://assets.example/ref.png" {
			t.Fatalf("request=%+v owner=%s", in, owner)
		}
		return []string{"data:image/png;base64," + base64.StdEncoding.EncodeToString(testImageToolPNG(t))}, 200, nil
	}
	st, _ := media.LoadMedia(ctx, "alice", p.ID)
	_, e = s.action(ctx, p, screenplayMediaAction{Action: "image", ModelID: "mine", Prompt: "人物", ReferenceItemIDs: []string{ref.ID}})
	if e != nil {
		t.Fatal(e)
	}
	if calls != 1 {
		t.Fatal(calls)
	}
	st, _ = media.LoadMedia(ctx, "alice", p.ID)
	if len(st.Items) != 3 {
		t.Fatal(st)
	}
	s.Images = func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
		return nil, 500, fmt.Errorf("provider offline")
	}
	_, e = s.action(ctx, p, screenplayMediaAction{Action: "image", ModelID: "mine", Prompt: "再试"})
	if e == nil {
		t.Fatal("expected failure")
	}
	st, _ = media.LoadMedia(ctx, "alice", p.ID)
	if st.Items[len(st.Items)-1].Status != "failed" {
		t.Fatal(st)
	}
}
func TestScreenplayMediaTimelinePackageAndScope(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	p, _ := projects.CreateProject(ctx, screenplay.Project{ID: "media-p", OwnerUsername: "alice", Title: "片", AspectRatio: "9:16"})
	objects := &memoryImageToolObjects{files: map[string][]byte{}}
	s := &ScreenplayMediaWorkbench{Projects: projects, Media: screenplay.NewMediaMemoryStore(), Cut: aicut.NewMemoryStore(), Objects: objects}
	item, e := s.saveBytes(ctx, p, "image", "image.png", "image/png", bytes.NewReader(testImageToolPNG(t)), "")
	if e != nil {
		t.Fatal(e)
	}
	st, _ := s.Media.LoadMedia(ctx, "alice", p.ID)
	tl, e := s.timeline(ctx, p, screenplayMediaAction{Clips: []screenplayMediaClip{{ItemID: item.ID, DurationMS: 2000, Scale: 1.2}}, Subtitles: []screenplayMediaSubtitle{{Text: "你好", DurationMS: 1000}}}, st)
	if e != nil {
		t.Fatal(e)
	}
	if tl.Tracks[0].Clips[0].CanvasTransform.ScaleX != 1.2 || tl.Tracks[2].Clips[0].Text != "你好" {
		t.Fatalf("timeline=%+v", tl)
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	if e = s.packageMedia(w, r, p); e != nil {
		t.Fatal(e)
	}
	z, e := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	if e != nil {
		t.Fatal(e)
	}
	if len(z.File) != 2 {
		t.Fatal(z.File)
	}
	found := false
	for _, f := range z.File {
		if strings.HasPrefix(f.Name, "media/") {
			body, _ := f.Open()
			raw, _ := io.ReadAll(body)
			body.Close()
			found = bytes.Equal(raw, testImageToolPNG(t))
		}
	}
	if !found {
		t.Fatal("package missing actual media")
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/?format=jianying", nil)
	if e = s.packageMedia(w, r, p); e != nil {
		t.Fatal(e)
	}
	native, e := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	if e != nil {
		t.Fatal(e)
	}
	files := map[string]bool{}
	for _, f := range native.File {
		files[f.Name] = true
	}
	if !files["draft_content.json"] || !files["draft_meta_info.json"] || !files["连接素材地址.ps1"] {
		t.Fatal(files)
	}
	mux := http.NewServeMux()
	now := time.Now()
	RegisterScreenplayMediaWorkbenchRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s)
	resp := signedJSONRequest(t, mux, now, "bob", "GET", "/api/comic-screenplays/projects/media-p/media", nil)
	if resp.Code == 200 {
		t.Fatal("foreign owner accepted")
	}
}
func TestScreenplayMediaRemoteReferencesRejectUnsupportedProtocol(t *testing.T) {
	_, status, err := generateIntegrationChannelImage(context.Background(), integrationChannelTarget{ModelID: "gpt-image-1", Kind: "image", BaseURL: "https://example.com/v1/images/generations", FullURL: true}, AgentImageGenerationInput{Prompt: "人物", ReferenceImages: []string{"https://assets.example/ref.png"}})
	if err == nil || status != 422 {
		t.Fatalf("status=%d error=%v", status, err)
	}
}

func TestScreenplayMediaJianyingReferentialIntegrityAndUnits(t *testing.T) {
	tl := aicut.Timeline{Tracks: []aicut.Track{{ID: "v", Kind: aicut.TrackVideo, Clips: []aicut.Clip{{ID: "shot", AssetID: "vid", Kind: aicut.ClipVideo, StartMS: 3000, DurationMS: 1500, SourceStartMS: 500, CanvasTransform: &aicut.CanvasTransform{ScaleX: 1.4, ScaleY: 1.4, Opacity: 1}}}}, {ID: "a", Kind: aicut.TrackAudio, Clips: []aicut.Clip{{ID: "voice", AssetID: "audio", Kind: aicut.ClipAudio, DurationMS: 1200}}}, {ID: "t", Kind: aicut.TrackSubtitle, Clips: []aicut.Clip{{ID: "text", Kind: aicut.ClipSubtitle, Text: "中文你好🌟", StartMS: 100, DurationMS: 900}}}}}
	assets := []aicut.Asset{{ID: "vid", Kind: "video", Name: "片段.mp4", DurationMS: 6000}, {ID: "audio", Kind: "audio", Name: "配音.mp3", DurationMS: 2000}}
	paths := map[string]string{"vid": "media/vid.mp4", "audio": "media/audio.mp3"}
	draft, meta, e := screenplayJianyingDraft(aicut.Project{ID: "p", Title: "作品", AspectRatio: "9:16"}, tl, assets, paths, map[string][2]int{"vid": {720, 1280}})
	if e != nil {
		t.Fatal(e)
	}
	if draft["duration"] != int64(4500000) || meta["tm_duration"] != int64(4500000) {
		t.Fatalf("duration=%v", draft["duration"])
	}
	materials := draft["materials"].(map[string][]map[string]any)
	ids := map[string]bool{}
	for _, kind := range []string{"videos", "audios", "texts", "speeds"} {
		for _, m := range materials[kind] {
			ids[m["id"].(string)] = true
			if p, ok := m["path"]; ok {
				found := false
				for _, expected := range paths {
					if p == expected {
						found = true
					}
				}
				if !found {
					t.Fatal("unpackaged path", p)
				}
			}
		}
	}
	tracks := draft["tracks"].([]map[string]any)
	for _, track := range tracks {
		for _, seg := range track["segments"].([]map[string]any) {
			if !ids[seg["material_id"].(string)] {
				t.Fatal("missing material", seg)
			}
			for _, ref := range seg["extra_material_refs"].([]string) {
				if !ids[ref] {
					t.Fatal("missing speed ref", ref)
				}
			}
		}
	}
	video := tracks[0]["segments"].([]map[string]any)[0]
	if video["source_timerange"].(map[string]int64)["start"] != 500000 || video["target_timerange"].(map[string]int64)["duration"] != 1500000 {
		t.Fatal(video)
	}
	if video["clip"].(map[string]any)["scale"].(map[string]float64)["x"] != 1.4 {
		t.Fatal(video)
	}
	if !strings.Contains(materials["texts"][0]["content"].(string), "中文你好🌟") {
		t.Fatal(materials["texts"])
	}
	delete(paths, "vid")
	if _, _, e = screenplayJianyingDraft(aicut.Project{}, tl, assets, paths); e == nil {
		t.Fatal("missing media accepted")
	}
}

func TestScreenplayMediaTTSUsesRoleVoiceAndIdempotency(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	p, _ := projects.CreateProject(ctx, screenplay.Project{ID: "tts-p", OwnerUsername: "alice", Title: "片"})
	s := &ScreenplayMediaWorkbench{Projects: projects, Media: screenplay.NewMediaMemoryStore(), Cut: aicut.NewMemoryStore(), Objects: agentworkspace.NewLocalAssetObjectStore(t.TempDir())}
	_, e := s.action(ctx, p, screenplayMediaAction{Action: "voice", TargetID: "narrator", VoiceID: "voice-own"})
	if e != nil {
		t.Fatal(e)
	}
	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var in ttsRequest
		if e := json.NewDecoder(r.Body).Decode(&in); e != nil {
			t.Fatal(e)
		}
		if in.Voice != "voice-own" || in.Input != "旁白内容" || in.Speed != 1.25 {
			t.Fatalf("input=%+v", in)
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Write([]byte("ID3-test-media"))
	}))
	defer upstream.Close()
	old := defaultTTSEndpoint
	defaultTTSEndpoint = upstream.URL
	defer func() { defaultTTSEndpoint = old }()
	in := screenplayMediaAction{Action: "tts", TargetID: "narrator", Text: "旁白内容", Speed: 1.25, IdempotencyKey: "once"}
	result, e := s.action(ctx, p, in)
	if e != nil {
		t.Fatal(e)
	}
	second, e := s.action(ctx, p, in)
	if e != nil {
		t.Fatal(e)
	}
	if result.(screenplay.MediaItem).ID != second.(screenplay.MediaItem).ID || calls != 1 {
		t.Fatal("duplicate paid request", calls)
	}
	st, _ := s.Media.LoadMedia(ctx, "alice", p.ID)
	if len(st.Items) != 2 {
		t.Fatal(st)
	}
}

func TestScreenplayMediaFinalizerRebuildUpdatesTimelineBeforeExport(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	p, _ := projects.CreateProject(ctx, screenplay.Project{ID: "final-p", OwnerUsername: "alice", Title: "片"})
	board, e := projects.SaveStoryboard(ctx, screenplay.Storyboard{ID: "board", ProjectID: p.ID, OwnerUsername: "alice", EpisodeNo: 1, Engine: screenplay.EngineXiaoyunqueV3, Shots: []screenplay.StoryboardShot{{ID: "shot", Seq: 1, Duration: 2, VideoDesc: "人物走路"}}})
	if e != nil {
		t.Fatal(e)
	}
	shot := board.Shots[0]
	repo := screenplay.NewMemoryProductionRepository()
	if e = repo.SaveShotTask(ctx, screenplay.ShotTask{Owner: "alice", ProjectID: p.ID, StoryboardID: board.ID, ShotID: shot.ID, Seq: 1, Episode: 1, TaskID: "new-task", Status: "succeeded", Artifact: "https://example.com/new.mp4"}); e != nil {
		t.Fatal(e)
	}
	cut := aicut.NewMemoryStore()
	media := screenplay.NewMediaMemoryStore()
	s := &ScreenplayMediaWorkbench{Projects: projects, Media: media, Cut: cut, Production: repo, Objects: agentworkspace.NewLocalAssetObjectStore(t.TempDir()), Assets: agentworkspace.NewMemoryAssetStore()}
	if e = s.ensureCut(ctx, p); e != nil {
		t.Fatal(e)
	}
	for _, id := range []string{"old", "new"} {
		a := aicut.Asset{ID: id, ProjectID: p.ID, Kind: "video", Name: id + ".mp4", ContentType: "video/mp4", ObjectKey: id, ByteSize: 1, DurationMS: 2000, CreatedAt: time.Now()}
		if e = cut.CreateAsset(ctx, "alice", a); e != nil {
			t.Fatal(e)
		}
		provider := id + "-task"
		s.append(ctx, p, screenplay.MediaItem{ID: id, AssetID: id, ShotID: shot.ID, TargetID: shot.ID, ProviderID: provider, Kind: "video", Status: "succeeded"})
	}
	st, _ := media.LoadMedia(ctx, "alice", p.ID)
	tl, e := s.timeline(ctx, p, screenplayMediaAction{Clips: []screenplayMediaClip{{ItemID: "old", DurationMS: 1000, Scale: 1.2}}}, st)
	if e != nil {
		t.Fatal(e)
	}
	s.Exports = NewAICutExportService(cut, s.Objects, nil, time.Now)
	_, e = s.Finalize(ctx, p)
	if e == nil {
		t.Fatal("missing export queue should fail")
	}
	after, _ := cut.Timeline(ctx, "alice", p.ID)
	if after.Revision <= tl.Revision || after.Tracks[0].Clips[0].AssetID != "new" || after.Tracks[0].Clips[0].CanvasTransform.ScaleX != 1.2 {
		t.Fatalf("stale timeline %+v", after)
	}
}

func TestScreenplayMediaPartialImportFailurePersistsWithoutResubmit(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	p, _ := projects.CreateProject(ctx, screenplay.Project{ID: "partial", OwnerUsername: "alice", Title: "片"})
	repo := screenplay.NewMemoryProductionRepository()
	repo.SaveShotTask(ctx, screenplay.ShotTask{Owner: "alice", ProjectID: p.ID, StoryboardID: "b", ShotID: "s", Seq: 1, TaskID: "known-task", Status: "succeeded", Artifact: "invalid-provider-result"})
	s := &ScreenplayMediaWorkbench{Projects: projects, Media: screenplay.NewMediaMemoryStore(), Cut: aicut.NewMemoryStore(), Production: repo, Objects: agentworkspace.NewLocalAssetObjectStore(t.TempDir()), Assets: agentworkspace.NewMemoryAssetStore()}
	s.syncProduced(ctx, p)
	s.syncProduced(ctx, p)
	st, e := s.Media.LoadMedia(ctx, "alice", p.ID)
	if e != nil {
		t.Fatal(e)
	}
	if len(st.Items) != 1 || st.Items[0].Status != "failed" || st.Items[0].ProviderID != "known-task" || st.Items[0].Error == "" {
		t.Fatal(st)
	}
}

func TestScreenplayMediaSeedreamAdapterSendsRealReferenceURLs(t *testing.T) {
	recorder := &integrationChannelRecorder{}
	server := httptest.NewServer(recorder)
	defer server.Close()
	images, status, e := generateIntegrationChannelImage(context.Background(), integrationChannelTarget{Kind: "image", ModelID: "doubao-seedream-4-5", BaseURL: server.URL + "/v1/images/generations", FullURL: true, Credential: "test-key", Params: map[string]any{"image": []string{"wrong-config-reference"}}}, AgentImageGenerationInput{Prompt: "人物", Count: 1, ReferenceImages: []string{"https://assets.example/actual-selected.png"}})
	if e != nil || status != 200 || len(images) != 1 {
		t.Fatalf("status=%d images=%v error=%v", status, images, e)
	}
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	refs, ok := recorder.body["image"].([]any)
	if !ok || len(refs) != 1 || refs[0] != "https://assets.example/actual-selected.png" {
		t.Fatal(recorder.body)
	}
}

func TestScreenplayMediaGridSplitAndPackagePrevalidation(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	p, _ := projects.CreateProject(ctx, screenplay.Project{ID: "grid", OwnerUsername: "alice", Title: "片"})
	a, e := projects.CreateAsset(ctx, screenplay.Asset{ID: "role", ProjectID: p.ID, OwnerUsername: "alice", Type: screenplay.AssetCharacter, Name: "角色"})
	if e != nil {
		t.Fatal(e)
	}
	s := &ScreenplayMediaWorkbench{Projects: projects, Media: screenplay.NewMediaMemoryStore(), Cut: aicut.NewMemoryStore(), Objects: agentworkspace.NewLocalAssetObjectStore(t.TempDir())}
	item, e := s.saveBytes(ctx, p, "image", "集合.png", "image/png", bytes.NewReader(testImageToolPNG(t)), "")
	if e != nil {
		t.Fatal(e)
	}
	result, e := s.action(ctx, p, screenplayMediaAction{Action: "split_grid", ItemID: item.ID, Rows: 1, Columns: 2, TargetIDs: []string{a.ID, ""}})
	if e != nil {
		t.Fatal(e)
	}
	created := result.(map[string]any)["createdItems"].([]screenplay.MediaItem)
	if len(created) != 2 {
		t.Fatal(created)
	}
	st, _ := s.Media.LoadMedia(ctx, "alice", p.ID)
	if st.Selected[a.ID] != created[0].ID {
		t.Fatal(st)
	}
	_, e = s.timeline(ctx, p, screenplayMediaAction{Clips: []screenplayMediaClip{{ItemID: created[0].ID, DurationMS: 2000}}}, st)
	if e != nil {
		t.Fatal(e)
	}
	w := httptest.NewRecorder()
	if e = s.packageMedia(w, httptest.NewRequest("GET", "/", nil), p); e != nil {
		t.Fatal(e)
	}
	archive, e := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	if e != nil {
		t.Fatal(e)
	}
	manifest, assets, e := parseScreenplayMediaPackage(ctx, archive)
	if e != nil || len(assets) != 3 || len(manifest.Timeline.Tracks[0].Clips) != 1 {
		t.Fatalf("assets=%d error=%v", len(assets), e)
	}
	var bad bytes.Buffer
	z := zip.NewWriter(&bad)
	f, _ := z.Create("../outside")
	f.Write([]byte("bad"))
	z.Close()
	archive, _ = zip.NewReader(bytes.NewReader(bad.Bytes()), int64(bad.Len()))
	if _, _, e = parseScreenplayMediaPackage(ctx, archive); e == nil {
		t.Fatal("zip traversal accepted")
	}
}

func TestScreenplayMediaPackageImportRestoresCurrentOwnerTimeline(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	source, _ := projects.CreateProject(ctx, screenplay.Project{ID: "source", OwnerUsername: "alice", Title: "原工程"})
	target, _ := projects.CreateProject(ctx, screenplay.Project{ID: "target", OwnerUsername: "bob", Title: "我的工程"})
	s := &ScreenplayMediaWorkbench{Projects: projects, Media: screenplay.NewMediaMemoryStore(), Cut: aicut.NewMemoryStore(), Objects: agentworkspace.NewLocalAssetObjectStore(t.TempDir())}
	item, e := s.saveBytes(ctx, source, "image", "画面.png", "image/png", bytes.NewReader(testImageToolPNG(t)), "")
	if e != nil {
		t.Fatal(e)
	}
	st, _ := s.Media.LoadMedia(ctx, "alice", source.ID)
	if _, e = s.timeline(ctx, source, screenplayMediaAction{Clips: []screenplayMediaClip{{ItemID: item.ID, DurationMS: 2000}}, Subtitles: []screenplayMediaSubtitle{{Text: "原中文字幕", DurationMS: 1000}}}, st); e != nil {
		t.Fatal(e)
	}
	out := httptest.NewRecorder()
	if e = s.packageMedia(out, httptest.NewRequest("GET", "/", nil), source); e != nil {
		t.Fatal(e)
	}
	var body bytes.Buffer
	multi := multipart.NewWriter(&body)
	file, _ := multi.CreateFormFile("file", "media.zip")
	file.Write(out.Body.Bytes())
	multi.WriteField("expectedRevision", "0")
	multi.Close()
	now := time.Now()
	mux := http.NewServeMux()
	RegisterScreenplayMediaWorkbenchRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s)
	req := httptest.NewRequest("POST", "/api/comic-screenplays/projects/target/media/import", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", multi.FormDataContentType())
	SignBridgeRequest(req, "bob", false, now, "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("import=%d %s", rec.Code, rec.Body.String())
	}
	tl, e := s.Cut.Timeline(ctx, "bob", target.ID)
	if e != nil || tl.Tracks[2].Clips[0].Text != "原中文字幕" {
		t.Fatalf("tl=%+v e=%v", tl, e)
	}
	newID := tl.Tracks[0].Clips[0].AssetID
	if newID == item.AssetID {
		t.Fatal("source identity reused")
	}
	if _, e = s.Cut.Asset(ctx, "bob", target.ID, newID); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Cut.Asset(ctx, "alice", target.ID, newID); e == nil {
		t.Fatal("owner scope leaked")
	}
	original, _ := s.Cut.Timeline(ctx, "alice", source.ID)
	if original.Tracks[0].Clips[0].AssetID != item.AssetID {
		t.Fatal("source overwritten")
	}
}

func TestScreenplayMediaSeparateGenerateAndEditPermissions(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	projects.CreateProject(ctx, screenplay.Project{ID: "shared-media", OwnerUsername: "alice", Title: "共享"})
	authors := &screenplay.AuthoringService{Store: projects, Repository: screenplay.NewMemoryAuthoringRepository()}
	st, e := authors.Load(ctx, "alice", "shared-media")
	if e != nil {
		t.Fatal(e)
	}
	st.Members = []screenplay.MemberGrant{{Username: "generator", Read: true, Generate: true}, {Username: "reader", Read: true}, {Username: "editor", Read: true, Edit: true}}
	if _, e = authors.Save(ctx, "alice", "shared-media", st, st.Revision); e != nil {
		t.Fatal(e)
	}
	calls := 0
	s := &ScreenplayMediaWorkbench{Projects: projects, Media: screenplay.NewMediaMemoryStore(), Cut: aicut.NewMemoryStore(), Objects: agentworkspace.NewLocalAssetObjectStore(t.TempDir()), Assets: agentworkspace.NewMemoryAssetStore(), Authoring: authors, Images: func(_ context.Context, owner string, _ AgentImageGenerationInput) ([]string, int, error) {
		calls++
		if owner != "alice" {
			t.Fatal(owner)
		}
		return nil, 500, fmt.Errorf("mock provider failure")
	}}
	now := time.Now()
	mux := http.NewServeMux()
	RegisterScreenplayMediaWorkbenchRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s)
	payload := map[string]any{"action": "image", "prompt": "人物", "modelId": "mine", "idempotencyKey": "permission-test"}
	for _, actor := range []string{"reader", "editor"} {
		res := signedJSONRequest(t, mux, now, actor, "POST", "/api/comic-screenplays/projects/shared-media/media/actions", payload)
		if res.Code != 403 {
			t.Fatalf("actor=%s code=%d %s", actor, res.Code, res.Body.String())
		}
	}
	res := signedJSONRequest(t, mux, now, "generator", "POST", "/api/comic-screenplays/projects/shared-media/media/actions", payload)
	if res.Code != 422 || calls != 1 {
		t.Fatalf("code=%d calls=%d %s", res.Code, calls, res.Body.String())
	}
	res = signedJSONRequest(t, mux, now, "generator", "POST", "/api/comic-screenplays/projects/shared-media/media/actions", map[string]any{"action": "voice", "targetId": "narrator", "voiceId": "voice"})
	if res.Code != 403 {
		t.Fatalf("generate-only user edited voice: %d", res.Code)
	}
}

func TestScreenplayMediaRenderedMP4PersistsAfterLocalExpiry(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	p, _ := projects.CreateProject(ctx, screenplay.Project{ID: "durable-final", OwnerUsername: "alice", Title: "成片"})
	objects := &memoryImageToolObjects{}
	cut := aicut.NewMemoryStore()
	s := &ScreenplayMediaWorkbench{Projects: projects, Media: screenplay.NewMediaMemoryStore(), Cut: cut, Objects: objects}
	if e := s.ensureCut(ctx, p); e != nil {
		t.Fatal(e)
	}
	project, _ := cut.Project(ctx, "alice", p.ID)
	tl, _ := cut.Timeline(ctx, "alice", p.ID)
	file := filepath.Join(t.TempDir(), "final.mp4")
	raw := []byte("mock-rendered-mp4-output")
	if e := os.WriteFile(file, raw, 0600); e != nil {
		t.Fatal(e)
	}
	now := time.Now()
	job := aicut.ExportJob{ID: "render-final", Owner: "alice", ProjectID: p.ID, Format: "mp4", State: aicut.ExportSucceeded, Progress: 100, OutputPath: file, ProjectSnapshot: project, TimelineSnapshot: tl, CreatedAt: now, UpdatedAt: now}
	if e := cut.CreateExportJob(ctx, "alice", job); e != nil {
		t.Fatal(e)
	}
	s.Exports = NewAICutExportService(cut, objects, nil, time.Now)
	item, e := s.persistExport(ctx, p, job, false)
	if e != nil || item.AssetID == "" {
		t.Fatalf("persist=%+v %v", item, e)
	}
	if e = os.Remove(file); e != nil {
		t.Fatal(e)
	}
	second, e := s.persistExport(ctx, p, job, false)
	if e != nil || second.AssetID != item.AssetID {
		t.Fatalf("expired local result lost %+v %v", second, e)
	}
	mux := http.NewServeMux()
	RegisterScreenplayMediaWorkbenchRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s)
	resp := signedJSONRequest(t, mux, now, "alice", "GET", "/api/comic-screenplays/projects/"+p.ID+"/media/exports/"+job.ID, nil)
	if resp.Code != 200 || !bytes.Equal(resp.Body.Bytes(), raw) {
		t.Fatalf("download=%d %q", resp.Code, resp.Body.Bytes())
	}
	foreign := signedJSONRequest(t, mux, now, "bob", "GET", "/api/comic-screenplays/projects/"+p.ID+"/media/exports/"+job.ID, nil)
	if foreign.Code == 200 {
		t.Fatal("foreign final download")
	}
}

func TestScreenplayMediaExportPersistenceFailureNeedsExplicitRetry(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	p, _ := projects.CreateProject(ctx, screenplay.Project{ID: "export-retry", OwnerUsername: "alice", Title: "片"})
	s := &ScreenplayMediaWorkbench{Projects: projects, Media: screenplay.NewMediaMemoryStore(), Cut: aicut.NewMemoryStore(), Objects: &memoryImageToolObjects{}}
	job := aicut.ExportJob{ID: "job", Owner: "alice", ProjectID: p.ID, State: aicut.ExportSucceeded, OutputPath: filepath.Join(t.TempDir(), "late.mp4"), CreatedAt: time.Now()}
	if _, e := s.persistExport(ctx, p, job, false); e == nil {
		t.Fatal("missing local output falsely complete")
	}
	st, _ := s.Media.LoadMedia(ctx, "alice", p.ID)
	if len(st.Items) != 1 || st.Items[0].Status != "failed" || st.Items[0].ProviderID != "job" {
		t.Fatal(st)
	}
	os.WriteFile(job.OutputPath, []byte("render-output"), 0600)
	if _, e := s.persistExport(ctx, p, job, false); e == nil {
		t.Fatal("read silently retried failed persistence")
	}
	item, e := s.persistExport(ctx, p, job, true)
	if e != nil || item.Status != "succeeded" || item.AssetID == "" {
		t.Fatalf("retry=%+v err=%v", item, e)
	}
}
