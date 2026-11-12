package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/batchfactoryv11"
)

type editPlanTestProvider struct {
	result  string
	calls   int
	request batchfactoryv11.TextCompletionRequest
}

func (p *editPlanTestProvider) Complete(_ context.Context, request batchfactoryv11.TextCompletionRequest) (string, error) {
	p.calls++
	p.request = request
	return p.result, nil
}

func newEditPlanRouteFixture(t *testing.T) (*aicut.MemoryStore, time.Time) {
	t.Helper()
	now := time.Date(2026, 11, 11, 11, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "自动剪辑", AspectRatio: "9:16", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	_, err := store.SaveTimeline(t.Context(), "alice", "cut-1", aicut.Timeline{Script: "自动字幕", Tracks: []aicut.Track{
		{ID: "video-1", Kind: aicut.TrackVideo, Name: "主画面", Clips: []aicut.Clip{{ID: "clip-1", AssetID: "asset-1", Kind: aicut.ClipVideo, Name: "镜头", DurationMS: 2000, SourceDurationMS: 3000, PlaybackRate: 1, Volume: 1}}},
		{ID: "subtitle-1", Kind: aicut.TrackSubtitle, Name: "字幕", Order: 1, Clips: []aicut.Clip{}},
	}}, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateAsset(t.Context(), "alice", aicut.Asset{ID: "asset-1", ProjectID: "cut-1", Name: "真实素材.mp4", Kind: "video", ContentType: "video/mp4", ObjectKey: "test/video.mp4", ByteSize: 100, DurationMS: 3000, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	return store, now
}

func TestAICutEditPlanLifecycleIsReviewFirstAndIdempotent(t *testing.T) {
	store, now := newEditPlanRouteFixture(t)
	provider := &editPlanTestProvider{result: `{"summary":{"sourceDurationMs":2000,"estimatedDurationMs":1800},"operations":[{"id":"caption-1","type":"add_caption","trackId":"subtitle-1","startMs":0,"durationMs":1800,"text":"自动字幕","reason":"补字幕","enabled":true}],"warnings":[]}`}
	mux := http.NewServeMux()
	RegisterAICutEditPlanRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, store, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }, func() time.Time { return now })
	body := map[string]any{"clientRequestId": "request-1", "modelId": "text-main", "expectedRevision": 1, "assetIds": []string{"asset-1"}, "goal": "自动剪成短片", "targetDurationMs": 1800, "aspectRatio": "9:16", "includeCaptions": true}
	created := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/edit-plans", body)
	if created.Code != http.StatusCreated || provider.calls != 1 || !containsText(created.Body.String(), `"state":"ready"`) {
		t.Fatalf("created=%d %s calls=%d", created.Code, created.Body.String(), provider.calls)
	}
	var payload struct {
		Plan aicut.EditPlan `json:"plan"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	duplicate := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/edit-plans", body)
	if duplicate.Code != http.StatusOK || provider.calls != 1 || !containsText(duplicate.Body.String(), payload.Plan.ID) {
		t.Fatalf("duplicate=%d %s calls=%d", duplicate.Code, duplicate.Body.String(), provider.calls)
	}
	current, _ := store.Timeline(t.Context(), "alice", "cut-1")
	if current.Revision != 1 || len(current.Tracks[1].Clips) != 0 {
		t.Fatalf("planning modified timeline: %#v", current)
	}
	apply := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/edit-plans/"+payload.Plan.ID+"/apply", map[string]any{"enabledOperationIds": []string{"caption-1"}})
	if apply.Code != http.StatusOK || !containsText(apply.Body.String(), `"revision":2`) || !containsText(apply.Body.String(), "自动字幕") {
		t.Fatalf("apply=%d %s", apply.Code, apply.Body.String())
	}
	again := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/edit-plans/"+payload.Plan.ID+"/apply", map[string]any{"enabledOperationIds": []string{"caption-1"}})
	if again.Code != http.StatusOK || !containsText(again.Body.String(), `"revision":2`) {
		t.Fatalf("repeat apply=%d %s", again.Code, again.Body.String())
	}
}

func TestAICutEditPlanRejectsEmptyAssetsAndMissingModel(t *testing.T) {
	store, now := newEditPlanRouteFixture(t)
	mux := http.NewServeMux()
	RegisterAICutEditPlanRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, store, nil, func() time.Time { return now })
	empty := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/edit-plans", map[string]any{"clientRequestId": "empty", "modelId": "text-main", "expectedRevision": 1, "goal": "自动剪辑", "targetDurationMs": 1000, "aspectRatio": "9:16"})
	if empty.Code != http.StatusBadRequest || !containsText(empty.Body.String(), "素材") {
		t.Fatalf("empty=%d %s", empty.Code, empty.Body.String())
	}
	missing := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/edit-plans", map[string]any{"clientRequestId": "missing", "modelId": "", "expectedRevision": 1, "assetIds": []string{"asset-1"}, "goal": "自动剪辑", "targetDurationMs": 1000, "aspectRatio": "9:16"})
	if missing.Code != http.StatusUnprocessableEntity || !containsText(missing.Body.String(), "文本模型未配置") {
		t.Fatalf("missing=%d %s", missing.Code, missing.Body.String())
	}
}

func TestAICutEditPlanRouteIsMountedByApplicationRouter(t *testing.T) {
	store, now := newEditPlanRouteFixture(t)
	provider := &editPlanTestProvider{result: `{"summary":{"sourceDurationMs":2000,"estimatedDurationMs":1800},"operations":[],"warnings":[]}`}
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now }, AICut: store,
		TextProviderResolver: func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil },
	})
	response := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/edit-plans", map[string]any{"clientRequestId": "router-1", "modelId": "text-main", "expectedRevision": 1, "assetIds": []string{"asset-1"}, "goal": "自动剪辑", "targetDurationMs": 1800, "aspectRatio": "9:16"})
	if response.Code != http.StatusCreated || !containsText(response.Body.String(), `"state":"ready"`) {
		t.Fatalf("response=%d %s", response.Code, response.Body.String())
	}
}

func TestEditPlanEmptyTimelineRealAssetPreviewApplyAndUndo(t *testing.T) {
	store, now := newEditPlanRouteFixture(t)
	_, err := store.SaveTimeline(t.Context(), "alice", "cut-1", aicut.Timeline{Tracks: []aicut.Track{}}, 1, now)
	if err != nil {
		t.Fatal(err)
	}
	provider := &editPlanTestProvider{result: `{"summary":{"sourceDurationMs":999999,"estimatedDurationMs":999999},"operations":[{"id":"place","type":"place_asset","trackId":"new-video","assetId":"asset-1","sourceStartMs":500,"durationMs":1500,"reason":"依据真实素材时长顺序拼接","enabled":true}],"warnings":[]}`}
	mux := http.NewServeMux()
	RegisterAICutEditPlanRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, store, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }, func() time.Time { return now })
	body := map[string]any{"clientRequestId": "empty-place", "modelId": "text-main", "expectedRevision": 2, "assetIds": []string{"asset-1"}, "goal": "拼接素材", "targetDurationMs": 1500, "aspectRatio": "9:16"}
	response := signedJSONRequest(t, mux, now, "alice", "POST", "/api/ai-cut/projects/cut-1/edit-plans", body)
	if response.Code != 201 {
		t.Fatalf("create=%d %s", response.Code, response.Body)
	}
	var payload struct {
		Plan aicut.EditPlan `json:"plan"`
	}
	_ = json.Unmarshal(response.Body.Bytes(), &payload)
	if !strings.Contains(provider.request.UserPrompt, "真实素材.mp4") || !strings.Contains(provider.request.UserPrompt, `"durationMs":3000`) || !strings.Contains(provider.request.UserPrompt, `"kind":"video"`) {
		t.Fatalf("provider missing actual metadata: %s", provider.request.UserPrompt)
	}
	if payload.Plan.Summary.EstimatedDurationMS != 1500 || len(payload.Plan.Request.SelectedAssets) != 1 || payload.Plan.Request.SelectedAssets[0].Name != "真实素材.mp4" || len(payload.Plan.Warnings) == 0 {
		t.Fatalf("plan=%+v", payload.Plan)
	}
	root := "/api/ai-cut/projects/cut-1/edit-plans/" + payload.Plan.ID
	preview := signedJSONRequest(t, mux, now, "alice", "POST", root+"/preview", map[string]any{"enabledOperationIds": []string{}})
	if preview.Code != 200 || !containsText(preview.Body.String(), `"estimatedDurationMs":0`) {
		t.Fatalf("preview=%d %s", preview.Code, preview.Body)
	}
	current, _ := store.Timeline(t.Context(), "alice", "cut-1")
	if current.Revision != 2 || len(current.Tracks) != 0 {
		t.Fatal("preview mutated timeline")
	}
	apply := signedJSONRequest(t, mux, now, "alice", "POST", root+"/apply", map[string]any{"enabledOperationIds": []string{"place"}})
	if apply.Code != 200 {
		t.Fatalf("apply=%d %s", apply.Code, apply.Body)
	}
	current, _ = store.Timeline(t.Context(), "alice", "cut-1")
	if current.Revision != 3 || current.Tracks[0].Clips[0].SourceDurationMS != 3000 {
		t.Fatalf("timeline=%+v", current)
	}
	undo, _, err := store.UndoTimeline(t.Context(), "alice", "cut-1", 3, now)
	if err != nil || len(undo.Tracks) != 0 {
		t.Fatalf("undo=%+v %v", undo, err)
	}
}
func TestEditPlanRejectsUnknownSelectedAssetBeforeProvider(t *testing.T) {
	store, now := newEditPlanRouteFixture(t)
	provider := &editPlanTestProvider{}
	mux := http.NewServeMux()
	RegisterAICutEditPlanRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, store, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }, func() time.Time { return now })
	response := signedJSONRequest(t, mux, now, "alice", "POST", "/api/ai-cut/projects/cut-1/edit-plans", map[string]any{"clientRequestId": "unknown-asset", "modelId": "text-main", "expectedRevision": 1, "assetIds": []string{"not-owned"}, "goal": "拼接", "aspectRatio": "9:16"})
	if response.Code != 404 || provider.calls != 0 {
		t.Fatalf("response=%d calls=%d", response.Code, provider.calls)
	}
}
