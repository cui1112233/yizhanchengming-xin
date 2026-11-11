package httpapi

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qiantie/backend/internal/aicut"
)

func TestAICutRoutesCreateAndPersistTimeline(t *testing.T) {
	now := time.Date(2026, 11, 9, 15, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	mux := http.NewServeMux()
	RegisterAICutRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func() time.Time { return now })

	created := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects", map[string]any{"title": "竖屏口播", "aspectRatio": "9:16"})
	if created.Code != http.StatusCreated {
		t.Fatalf("create=%d %s", created.Code, created.Body.String())
	}
	var createPayload struct {
		Project aicut.Project `json:"project"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &createPayload); err != nil {
		t.Fatal(err)
	}
	if createPayload.Project.ID == "" || createPayload.Project.Title != "竖屏口播" {
		t.Fatalf("project=%+v", createPayload.Project)
	}

	path := "/api/ai-cut/projects/" + createPayload.Project.ID + "/timeline"
	saved := signedJSONRequest(t, mux, now, "alice", http.MethodPut, path, map[string]any{
		"expectedRevision": 0,
		"timeline": map[string]any{"revision": 0, "tracks": []any{map[string]any{
			"id": "video-main", "kind": "video", "name": "主画面", "order": 0,
			"clips": []any{map[string]any{"id": "opening", "kind": "video", "assetId": "asset-1", "name": "开场", "startMs": 0, "durationMs": 3000, "sourceDurationMs": 5000, "linkGroupId": "link-opening", "compoundId": "compound-opening", "compoundName": "开场复合片段"}},
		}}},
	})
	if saved.Code != http.StatusOK || !containsText(saved.Body.String(), `"revision":1`) || !containsText(saved.Body.String(), `"linkGroupId":"link-opening"`) || !containsText(saved.Body.String(), `"compoundId":"compound-opening"`) || !containsText(saved.Body.String(), `"compoundName":"开场复合片段"`) {
		t.Fatalf("save=%d %s", saved.Code, saved.Body.String())
	}

	stale := signedJSONRequest(t, mux, now, "alice", http.MethodPut, path, map[string]any{"expectedRevision": 0, "timeline": map[string]any{"tracks": []any{}}})
	if stale.Code != http.StatusConflict || !containsText(stale.Body.String(), "时间线已更新") {
		t.Fatalf("stale=%d %s", stale.Code, stale.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodGet, path, nil)
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign=%d %s", foreign.Code, foreign.Body.String())
	}
}

func TestAICutRoutesPersistUndoRedoAcrossRequests(t *testing.T) {
	now := time.Date(2026, 11, 10, 3, 30, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-history", Title: "历史", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	first, _ := store.SaveTimeline(t.Context(), "alice", "cut-history", aicut.Timeline{Script: "第一版", Tracks: []aicut.Track{}}, 0, now)
	second, _ := store.SaveTimeline(t.Context(), "alice", "cut-history", aicut.Timeline{Script: "第二版", Tracks: []aicut.Track{}}, first.Revision, now)
	mux := http.NewServeMux()
	RegisterAICutRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func() time.Time { return now })
	path := "/api/ai-cut/projects/cut-history/timeline/history"
	state := signedJSONRequest(t, mux, now, "alice", http.MethodGet, path, nil)
	if state.Code != http.StatusOK || !containsText(state.Body.String(), `"canUndo":true`) || !containsText(state.Body.String(), `"canRedo":false`) {
		t.Fatalf("state=%d %s", state.Code, state.Body.String())
	}
	undo := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/undo", map[string]any{"expectedRevision": second.Revision})
	if undo.Code != http.StatusOK || !containsText(undo.Body.String(), `"script":"第一版"`) || !containsText(undo.Body.String(), `"canRedo":true`) {
		t.Fatalf("undo=%d %s", undo.Code, undo.Body.String())
	}
	redo := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/redo", map[string]any{"expectedRevision": second.Revision + 1})
	if redo.Code != http.StatusOK || !containsText(redo.Body.String(), `"script":"第二版"`) {
		t.Fatalf("redo=%d %s", redo.Code, redo.Body.String())
	}
}

func TestAICutRoutesRejectEmptyRename(t *testing.T) {
	now := time.Date(2026, 11, 9, 15, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	project := aicut.Project{ID: "cut-1", Title: "旧名称", AspectRatio: "16:9", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(t.Context(), "alice", project); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAICutRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func() time.Time { return now })
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPatch, "/api/ai-cut/projects/cut-1", map[string]any{"title": "  "})
	if response.Code != http.StatusBadRequest || !containsText(response.Body.String(), "项目名称不能为空") {
		t.Fatalf("rename=%d %s", response.Code, response.Body.String())
	}
}

func TestAICutRoutesUpdateProjectCanvasAppearance(t *testing.T) {
	now := time.Date(2026, 11, 11, 16, 5, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-alpha", Title: "透明工程", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAICutRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func() time.Time { return now })
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPatch, "/api/ai-cut/projects/cut-alpha", map[string]any{
		"backgroundMode": "transparent", "backgroundColor": "#224466",
	})
	if response.Code != http.StatusOK || !containsText(response.Body.String(), `"backgroundMode":"transparent"`) || !containsText(response.Body.String(), `"backgroundColor":"#224466"`) {
		t.Fatalf("update=%d %s", response.Code, response.Body.String())
	}
	invalid := signedJSONRequest(t, mux, now, "alice", http.MethodPatch, "/api/ai-cut/projects/cut-alpha", map[string]any{
		"backgroundMode": "transparent", "backgroundColor": "red",
	})
	if invalid.Code != http.StatusBadRequest || !containsText(invalid.Body.String(), "画布背景") {
		t.Fatalf("invalid=%d %s", invalid.Code, invalid.Body.String())
	}
}

func TestAICutImportSRTNormalDuplicateEmptyAndStaleRevision(t *testing.T) {
	now := time.Date(2026, 11, 9, 17, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "字幕测试", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAICutRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func() time.Time { return now })
	path := "/api/ai-cut/projects/cut-1/subtitles/import"

	first := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, map[string]any{
		"filename": "dialogue.srt", "expectedRevision": 0,
		"srt": "1\n00:00:01,000 --> 00:00:02,500\n第一句\n\n2\n00:00:03,000 --> 00:00:04,000\n第二句",
	})
	if first.Code != http.StatusOK || !containsText(first.Body.String(), `"imported":2`) || !containsText(first.Body.String(), `"revision":1`) {
		t.Fatalf("first=%d %s", first.Code, first.Body.String())
	}

	duplicate := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, map[string]any{
		"filename": "dialogue.srt", "expectedRevision": 1,
		"srt": "1\n00:00:01,000 --> 00:00:02,500\n第一句",
	})
	if duplicate.Code != http.StatusOK || !containsText(duplicate.Body.String(), `"revision":2`) {
		t.Fatalf("duplicate=%d %s", duplicate.Code, duplicate.Body.String())
	}
	var payload struct {
		Timeline aicut.Timeline `json:"timeline"`
	}
	if err := json.Unmarshal(duplicate.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Timeline.Tracks) != 2 {
		t.Fatalf("duplicate import tracks=%d, want 2", len(payload.Timeline.Tracks))
	}

	empty := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, map[string]any{"filename": "empty.srt", "expectedRevision": 2, "srt": "  "})
	if empty.Code != http.StatusBadRequest || !containsText(empty.Body.String(), "没有读取到有效字幕") {
		t.Fatalf("empty=%d %s", empty.Code, empty.Body.String())
	}

	stale := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, map[string]any{
		"filename": "stale.srt", "expectedRevision": 1, "srt": "1\n00:00:00,000 --> 00:00:01,000\n旧版本",
	})
	if stale.Code != http.StatusConflict || !containsText(stale.Body.String(), "时间线已更新") {
		t.Fatalf("stale=%d %s", stale.Code, stale.Body.String())
	}
}

func TestAICutRoutesAreMountedByApplicationRouter(t *testing.T) {
	now := time.Date(2026, 11, 9, 15, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AICut: store})
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/ai-cut/projects", nil)
	if response.Code != http.StatusOK || !containsText(response.Body.String(), `"projects":[]`) {
		t.Fatalf("projects=%d %s", response.Code, response.Body.String())
	}
}

func TestAICutASRRoutesSubmitListAndCancel(t *testing.T) {
	now := time.Date(2026, 11, 9, 19, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "语音识别", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateAsset(t.Context(), "alice", aicut.Asset{ID: "audio-1", ProjectID: "cut-1", Kind: "audio", Name: "voice.mp3", ContentType: "audio/mpeg", ObjectKey: "voice.mp3", ByteSize: 10, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	service := aicut.NewASRService(aicut.ASRServiceOptions{Store: store, Now: func() time.Time { return now }})
	mux := http.NewServeMux()
	RegisterAICutASRRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, service)
	created := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/asr-jobs", map[string]any{"assetId": "audio-1", "modelId": "asr-main", "language": "zh", "expectedRevision": 0})
	if created.Code != http.StatusAccepted || !containsText(created.Body.String(), `"state":"queued"`) {
		t.Fatalf("create=%d %s", created.Code, created.Body.String())
	}
	var payload struct {
		Job aicut.ASRJob `json:"job"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	listed := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/ai-cut/projects/cut-1/asr-jobs", nil)
	if listed.Code != http.StatusOK || !containsText(listed.Body.String(), payload.Job.ID) {
		t.Fatalf("list=%d %s", listed.Code, listed.Body.String())
	}
	canceled := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/asr-jobs/"+payload.Job.ID+"/cancel", map[string]any{})
	if canceled.Code != http.StatusOK || !containsText(canceled.Body.String(), `"state":"canceled"`) {
		t.Fatalf("cancel=%d %s", canceled.Code, canceled.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodGet, "/api/ai-cut/projects/cut-1/asr-jobs", nil)
	if foreign.Code != http.StatusOK || !containsText(foreign.Body.String(), `"jobs":[]`) {
		t.Fatalf("foreign=%d %s", foreign.Code, foreign.Body.String())
	}
}

func TestAICutAssetUploadPersistsObjectAndMetadata(t *testing.T) {
	now := time.Date(2026, 11, 9, 16, 30, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "测试", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	objects := &agentAssetObjectTestStore{files: map[string][]byte{}}
	mux := http.NewServeMux()
	RegisterAICutAssetRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, func() time.Time { return now })

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "source.mp4")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("\x00\x00\x00\x18ftypmp42video"))
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/ai-cut/projects/cut-1/assets", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	SignBridgeRequest(req, "alice", true, now, "secret")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	if response.Code != http.StatusCreated || !containsText(response.Body.String(), `"kind":"video"`) || containsText(response.Body.String(), "objectKey") {
		t.Fatalf("upload=%d %s", response.Code, response.Body.String())
	}
	if len(objects.files) != 1 {
		t.Fatalf("objects=%d", len(objects.files))
	}
}
