package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/batchfactoryv11"
)

type aicutCaptionProvider struct {
	result  string
	err     error
	request batchfactoryv11.TextCompletionRequest
}

func (p *aicutCaptionProvider) Complete(_ context.Context, input batchfactoryv11.TextCompletionRequest) (string, error) {
	p.request = input
	return p.result, p.err
}

func TestAICutCaptionCorrectionUsesConfiguredModelAndAtomicallySavesTextOnly(t *testing.T) {
	now := time.Date(2026, 11, 9, 18, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "校对", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	original := aicut.Timeline{Tracks: []aicut.Track{{ID: "subtitle-1", Kind: aicut.TrackSubtitle, Clips: []aicut.Clip{{ID: "sub-1", Kind: aicut.ClipSubtitle, Name: "字幕", StartMS: 1200, DurationMS: 1800, SourceDurationMS: 1800, Text: "落水之畔"}}}}}
	if _, err := store.SaveTimeline(t.Context(), "alice", "cut-1", original, 0, now); err != nil {
		t.Fatal(err)
	}
	provider := &aicutCaptionProvider{result: `{"corrections":[{"clipId":"sub-1","text":"洛水之畔","reason":"剧本原词"}]}`}
	var selectedModel string
	resolver := func(_ context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
		if owner != "alice" {
			t.Fatalf("owner=%q", owner)
		}
		selectedModel = modelID
		return provider, nil
	}
	mux := http.NewServeMux()
	RegisterAICutAIRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, resolver, func() time.Time { return now })
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/captions/correct", map[string]any{
		"script": "故事发生在洛水之畔。", "modelId": "text-main", "expectedRevision": 1,
	})
	if response.Code != http.StatusOK || !containsText(response.Body.String(), `"applied":1`) || selectedModel != "text-main" {
		t.Fatalf("response=%d %s model=%q", response.Code, response.Body.String(), selectedModel)
	}
	var payload struct {
		Timeline aicut.Timeline `json:"timeline"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	clip := payload.Timeline.Tracks[0].Clips[0]
	if clip.Text != "洛水之畔" || clip.StartMS != 1200 || clip.DurationMS != 1800 || payload.Timeline.Revision != 2 {
		t.Fatalf("timeline=%+v", payload.Timeline)
	}
	if provider.request.SystemPrompt == "" || provider.request.UserPrompt == "" {
		t.Fatalf("provider request missing prompt: %+v", provider.request)
	}
}

func TestAICutCaptionCorrectionHandlesEmptyModelFailureAndCancellation(t *testing.T) {
	now := time.Date(2026, 11, 9, 18, 30, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	_ = store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "校对", CreatedAt: now, UpdatedAt: now})
	_, _ = store.SaveTimeline(t.Context(), "alice", "cut-1", aicut.Timeline{Tracks: []aicut.Track{{ID: "subtitle-1", Kind: aicut.TrackSubtitle, Clips: []aicut.Clip{{ID: "sub-1", Kind: aicut.ClipSubtitle, DurationMS: 1000, SourceDurationMS: 1000, Text: "字幕"}}}}}, 0, now)

	tests := []struct {
		name        string
		body        map[string]any
		provider    batchfactoryv11.DirectorProvider
		resolverErr error
		wantStatus  int
		wantText    string
	}{
		{name: "empty script", body: map[string]any{"script": "", "expectedRevision": 1}, provider: &aicutCaptionProvider{}, wantStatus: http.StatusBadRequest, wantText: "校对剧本不能为空"},
		{name: "model missing", body: map[string]any{"script": "剧本", "expectedRevision": 1}, wantStatus: http.StatusUnprocessableEntity, wantText: "文本模型未配置"},
		{name: "provider failure", body: map[string]any{"script": "剧本", "expectedRevision": 1}, provider: &aicutCaptionProvider{err: errors.New("upstream failed")}, wantStatus: http.StatusBadGateway, wantText: "字幕纠错失败"},
		{name: "cancelled", body: map[string]any{"script": "剧本", "expectedRevision": 1}, provider: &aicutCaptionProvider{err: context.Canceled}, wantStatus: http.StatusRequestTimeout, wantText: "字幕纠错已取消"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			resolver := func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
				return tc.provider, tc.resolverErr
			}
			RegisterAICutAIRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, resolver, func() time.Time { return now })
			response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/captions/correct", tc.body)
			if response.Code != tc.wantStatus || !containsText(response.Body.String(), tc.wantText) {
				t.Fatalf("response=%d %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestAICutCaptionCorrectionRouteIsMountedByApplicationRouter(t *testing.T) {
	now := time.Date(2026, 11, 9, 19, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	_ = store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "校对", CreatedAt: now, UpdatedAt: now})
	_, _ = store.SaveTimeline(t.Context(), "alice", "cut-1", aicut.Timeline{Tracks: []aicut.Track{{ID: "subtitle-1", Kind: aicut.TrackSubtitle, Clips: []aicut.Clip{{ID: "sub-1", Kind: aicut.ClipSubtitle, DurationMS: 1000, SourceDurationMS: 1000, Text: "字幕"}}}}}, 0, now)
	provider := &aicutCaptionProvider{result: `{"corrections":[]}`}
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now }, AICut: store,
		TextProviderResolver: func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil },
	})
	response := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/captions/correct", map[string]any{"script": "剧本", "expectedRevision": 1})
	if response.Code != http.StatusOK || !containsText(response.Body.String(), `"applied":0`) {
		t.Fatalf("response=%d %s", response.Code, response.Body.String())
	}
}

func TestAICutFillerTrimNormalEmptyRepeatFailureAndCancel(t *testing.T) {
	now := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
	newStore := func(t *testing.T) *aicut.MemoryStore {
		store := aicut.NewMemoryStore()
		if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "去口误", CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatal(err)
		}
		_, err := store.SaveTimeline(t.Context(), "alice", "cut-1", aicut.Timeline{Tracks: []aicut.Track{
			{ID: "subtitle-1", Kind: aicut.TrackSubtitle, Clips: []aicut.Clip{
				{ID: "sub-1", Kind: aicut.ClipSubtitle, DurationMS: 500, SourceDurationMS: 500, Text: "嗯"},
				{ID: "sub-2", Kind: aicut.ClipSubtitle, StartMS: 500, DurationMS: 1500, SourceDurationMS: 1500, Text: "哈哈哈我寻思两个人分摊"},
				{ID: "sub-3", Kind: aicut.ClipSubtitle, StartMS: 2000, DurationMS: 1500, SourceDurationMS: 1500, Text: "那你现在有住的地方吗"},
			}},
			{ID: "video-1", Kind: aicut.TrackVideo, Clips: []aicut.Clip{{ID: "video-clip", Kind: aicut.ClipVideo, AssetID: "asset-1", DurationMS: 3500, SourceDurationMS: 3500}}},
		}}, 0, now)
		if err != nil {
			t.Fatal(err)
		}
		return store
	}

	t.Run("normal then repeat", func(t *testing.T) {
		store := newStore(t)
		provider := &aicutCaptionProvider{result: `{"fillers":[{"clipId":"sub-1","reason":"整行只有语气词"}]}`}
		mux := http.NewServeMux()
		RegisterAICutAIRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }, func() time.Time { return now })
		response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/filler-trim", map[string]any{"expectedRevision": 1})
		if response.Code != http.StatusOK || !containsText(response.Body.String(), `"fillerCount":1`) {
			t.Fatalf("response=%d %s", response.Code, response.Body.String())
		}
		current, _ := store.Timeline(t.Context(), "alice", "cut-1")
		if len(current.Tracks[0].Clips) != 2 || current.Tracks[0].Clips[0].ID != "sub-2" || len(current.Tracks[1].Clips) != 1 || current.Tracks[1].Clips[0].DurationMS != 3500 {
			t.Fatalf("direct filler delete changed the wrong timeline data: %+v", current)
		}
		provider.result = `{"fillers":[]}`
		repeated := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/filler-trim", map[string]any{"expectedRevision": 2})
		if repeated.Code != http.StatusOK || !containsText(repeated.Body.String(), `"removedMs":0`) {
			t.Fatalf("repeat=%d %s", repeated.Code, repeated.Body.String())
		}
	})

	t.Run("empty subtitles", func(t *testing.T) {
		store := aicut.NewMemoryStore()
		_ = store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-empty", Title: "空项目", CreatedAt: now, UpdatedAt: now})
		mux := http.NewServeMux()
		RegisterAICutAIRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
			return &aicutCaptionProvider{}, nil
		}, func() time.Time { return now })
		response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-empty/filler-trim", map[string]any{"expectedRevision": 0})
		if response.Code != http.StatusBadRequest || !containsText(response.Body.String(), "没有可处理的字幕") {
			t.Fatalf("response=%d %s", response.Code, response.Body.String())
		}
	})

	for _, tc := range []struct {
		name, result string
		err          error
		status       int
		text         string
	}{
		{name: "bad model output", result: `{"fillers":[{"clipId":"missing"}]}`, status: http.StatusBadGateway, text: "不存在"},
		{name: "provider failure", err: errors.New("upstream failed"), status: http.StatusBadGateway, text: "去口误停顿失败"},
		{name: "cancel", err: context.Canceled, status: http.StatusRequestTimeout, text: "去口误停顿已取消"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newStore(t)
			provider := &aicutCaptionProvider{result: tc.result, err: tc.err}
			mux := http.NewServeMux()
			RegisterAICutAIRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }, func() time.Time { return now })
			response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/filler-trim", map[string]any{"expectedRevision": 1})
			if response.Code != tc.status || !containsText(response.Body.String(), tc.text) {
				t.Fatalf("response=%d %s", response.Code, response.Body.String())
			}
			current, _ := store.Timeline(t.Context(), "alice", "cut-1")
			if current.Revision != 1 || len(current.Tracks[0].Clips) != 3 || len(current.Tracks[1].Clips) != 1 {
				t.Fatalf("failure changed timeline: %+v", current)
			}
		})
	}
}

func TestAICutCaptionCorrectionPreviewThenApply(t *testing.T) {
	now := time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "预览", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	original := aicut.Timeline{Tracks: []aicut.Track{{ID: "subtitle-1", Kind: aicut.TrackSubtitle, Clips: []aicut.Clip{{ID: "sub-1", Kind: aicut.ClipSubtitle, Name: "字幕", StartMS: 1200, DurationMS: 1800, SourceDurationMS: 1800, Text: "落水之畔"}}}}}
	if _, err := store.SaveTimeline(t.Context(), "alice", "cut-1", original, 0, now); err != nil {
		t.Fatal(err)
	}
	provider := &aicutCaptionProvider{result: `{"corrections":[{"clipId":"sub-1","text":"洛水之畔","reason":"剧本原词"}]}`}
	mux := http.NewServeMux()
	RegisterAICutAIRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }, func() time.Time { return now })

	previewed := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/captions/correct", map[string]any{
		"script": "故事发生在洛水之畔。", "modelId": "text-main", "expectedRevision": 1, "preview": true,
	})
	if previewed.Code != http.StatusOK || !containsText(previewed.Body.String(), `"message":"分析完成"`) {
		t.Fatalf("preview=%d %s", previewed.Code, previewed.Body.String())
	}
	var payload struct {
		Preview []struct {
			ClipID  string `json:"clipId"`
			StartMS int64  `json:"startMs"`
			EndMS   int64  `json:"endMs"`
			OldText string `json:"oldText"`
			NewText string `json:"newText"`
			Reason  string `json:"reason"`
		} `json:"preview"`
		Applied  int   `json:"applied"`
		Revision int64 `json:"revision"`
	}
	if err := json.Unmarshal(previewed.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Preview) != 1 || payload.Applied != 1 || payload.Revision != 1 {
		t.Fatalf("preview payload=%+v", payload)
	}
	item := payload.Preview[0]
	if item.ClipID != "sub-1" || item.OldText != "落水之畔" || item.NewText != "洛水之畔" || item.StartMS != 1200 || item.EndMS != 3000 || item.Reason != "剧本原词" {
		t.Fatalf("preview item=%+v", item)
	}
	unchanged, _ := store.Timeline(t.Context(), "alice", "cut-1")
	if unchanged.Revision != 1 || unchanged.Tracks[0].Clips[0].Text != "落水之畔" {
		t.Fatalf("preview changed timeline: %+v", unchanged)
	}

	provider.request = batchfactoryv11.TextCompletionRequest{}
	applied := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/captions/correct", map[string]any{
		"expectedRevision": 1,
		"corrections":      []map[string]any{{"clipId": item.ClipID, "text": item.NewText, "reason": item.Reason}},
	})
	if applied.Code != http.StatusOK || !containsText(applied.Body.String(), `"applied":1`) || !containsText(applied.Body.String(), "字幕纠错已写回") {
		t.Fatalf("apply=%d %s", applied.Code, applied.Body.String())
	}
	if provider.request.SystemPrompt != "" || provider.request.UserPrompt != "" {
		t.Fatalf("apply must not call the model: %+v", provider.request)
	}
	saved, _ := store.Timeline(t.Context(), "alice", "cut-1")
	if saved.Revision != 2 || saved.Tracks[0].Clips[0].Text != "洛水之畔" || saved.Tracks[0].Clips[0].DurationMS != 1800 {
		t.Fatalf("applied timeline=%+v", saved)
	}
}

func TestAICutFillerTrimPreviewThenApply(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "去水词", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveTimeline(t.Context(), "alice", "cut-1", aicut.Timeline{Tracks: []aicut.Track{
		{ID: "subtitle-1", Kind: aicut.TrackSubtitle, Clips: []aicut.Clip{
			{ID: "sub-1", Kind: aicut.ClipSubtitle, DurationMS: 500, SourceDurationMS: 500, Text: "嗯"},
			{ID: "sub-2", Kind: aicut.ClipSubtitle, StartMS: 500, DurationMS: 1500, SourceDurationMS: 1500, Text: "哈哈哈我寻思两个人分摊"},
			{ID: "sub-3", Kind: aicut.ClipSubtitle, StartMS: 2000, DurationMS: 1500, SourceDurationMS: 1500, Text: "那你现在有住的地方吗"},
		}},
		{ID: "video-1", Kind: aicut.TrackVideo, Clips: []aicut.Clip{{ID: "video-clip", Kind: aicut.ClipVideo, AssetID: "asset-1", DurationMS: 3500, SourceDurationMS: 3500}}},
	}}, 0, now); err != nil {
		t.Fatal(err)
	}
	provider := &aicutCaptionProvider{result: `{"fillers":[{"clipId":"sub-1","reason":"整行只有语气词"}]}`}
	mux := http.NewServeMux()
	RegisterAICutAIRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }, func() time.Time { return now })

	previewed := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/filler-trim", map[string]any{
		"modelId": "text-main", "expectedRevision": 1, "preview": true,
	})
	if previewed.Code != http.StatusOK || !containsText(previewed.Body.String(), `"message":"分析完成"`) || !containsText(previewed.Body.String(), `"fillerCount":1`) {
		t.Fatalf("preview=%d %s", previewed.Code, previewed.Body.String())
	}
	var payload struct {
		Preview []struct {
			ClipID  string `json:"clipId"`
			StartMS int64  `json:"startMs"`
			EndMS   int64  `json:"endMs"`
			Text    string `json:"text"`
			Reason  string `json:"reason"`
		} `json:"preview"`
		Summary  aicut.FillerTrimSummary `json:"summary"`
		Revision int64                   `json:"revision"`
	}
	if err := json.Unmarshal(previewed.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Preview) != 1 || payload.Revision != 1 || payload.Summary.FillerCount != 1 || payload.Summary.RemovedMS != 500 {
		t.Fatalf("preview payload=%+v", payload)
	}
	item := payload.Preview[0]
	if item.ClipID != "sub-1" || item.Text != "嗯" || item.StartMS != 0 || item.EndMS != 500 || item.Reason != "整行只有语气词" {
		t.Fatalf("preview item=%+v", item)
	}
	unchanged, _ := store.Timeline(t.Context(), "alice", "cut-1")
	if unchanged.Revision != 1 || len(unchanged.Tracks[0].Clips) != 3 || unchanged.Tracks[0].Clips[0].ID != "sub-1" {
		t.Fatalf("preview deleted subtitles: %+v", unchanged)
	}

	provider.request = batchfactoryv11.TextCompletionRequest{}
	applied := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/filler-trim", map[string]any{
		"expectedRevision": 1,
		"fillers":          []map[string]any{{"clipId": item.ClipID, "reason": item.Reason}},
	})
	if applied.Code != http.StatusOK || !containsText(applied.Body.String(), `"fillerCount":1`) || !containsText(applied.Body.String(), "水词字幕删除完成") {
		t.Fatalf("apply=%d %s", applied.Code, applied.Body.String())
	}
	if provider.request.SystemPrompt != "" || provider.request.UserPrompt != "" {
		t.Fatalf("apply must not call the model: %+v", provider.request)
	}
	saved, _ := store.Timeline(t.Context(), "alice", "cut-1")
	if saved.Revision != 2 || len(saved.Tracks[0].Clips) != 2 || saved.Tracks[0].Clips[0].ID != "sub-2" || len(saved.Tracks[1].Clips) != 1 || saved.Tracks[1].Clips[0].DurationMS != 3500 {
		t.Fatalf("applied timeline=%+v", saved)
	}
}

func TestAICutOralCleanUsesSeparateTimelineCutRoute(t *testing.T) {
	now := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-oral", Title: "口播", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	_, err := store.SaveTimeline(t.Context(), "alice", "cut-oral", aicut.Timeline{Tracks: []aicut.Track{
		{ID: "subtitle-1", Kind: aicut.TrackSubtitle, Clips: []aicut.Clip{
			{ID: "sub-1", Kind: aicut.ClipSubtitle, DurationMS: 500, SourceDurationMS: 500, Text: "嗯"},
			{ID: "sub-2", Kind: aicut.ClipSubtitle, StartMS: 2500, DurationMS: 1000, SourceDurationMS: 1000, Text: "那你现在有住的地方吗"},
		}},
		{ID: "video-1", Kind: aicut.TrackVideo, Clips: []aicut.Clip{{ID: "video-clip", Kind: aicut.ClipVideo, AssetID: "asset-1", DurationMS: 4000, SourceDurationMS: 4000}}},
	}}, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	provider := &aicutCaptionProvider{result: `{"fillers":[{"clipId":"sub-1"}]}`}
	mux := http.NewServeMux()
	RegisterAICutAIRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }, func() time.Time { return now })
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-oral/oral-clean", map[string]any{"expectedRevision": 1, "maxPauseMs": 1000})
	if response.Code != http.StatusOK || !containsText(response.Body.String(), `"fillerCount":1`) || !containsText(response.Body.String(), `"pauseCount":1`) {
		t.Fatalf("response=%d %s", response.Code, response.Body.String())
	}
	current, _ := store.Timeline(t.Context(), "alice", "cut-oral")
	if current.Revision != 2 || len(current.Tracks[0].Clips) != 1 || current.Tracks[0].Clips[0].StartMS != 1000 || len(current.Tracks[1].Clips) != 2 {
		t.Fatalf("oral clean timeline=%+v", current)
	}
}
