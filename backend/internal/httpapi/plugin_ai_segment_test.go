package httpapi

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestParseAISegmentSRTCoversEverySubtitleBoundary(t *testing.T) {
	clips, err := parseAISegmentSRT("1\n00:00:00,000 --> 00:00:12,500\n开场\n\n2\n00:00:12,500 --> 00:00:47,000\n冲突出现\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(clips) != 2 || clips[0].StartUS != 0 || clips[0].EndUS != 12_500_000 || clips[1].Subtitle != "冲突出现" {
		t.Fatalf("clips=%+v", clips)
	}
}

func TestParseAISegmentDraftFindsMainTrackAndOverlappingText(t *testing.T) {
	draft := []byte(`{
		"materials":{"texts":[{"id":"text-1","content":"第一句"}]},
		"tracks":[
			{"type":"video","segments":[
				{"target_timerange":{"start":0,"duration":50000000}},
				{"target_timerange":{"start":50000000,"duration":55000000}}
			]},
			{"type":"text","id":"track-text","segments":[
				{"material_id":"text-1","target_timerange":{"start":45000000,"duration":15000000}}
			]}
		]
	}`)
	clips, tracks, err := parseAISegmentDraft(draft, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(clips) != 2 || len(tracks) != 1 || clips[0].Subtitle != "第一句" || clips[1].Subtitle != "第一句" {
		t.Fatalf("clips=%+v tracks=%+v", clips, tracks)
	}
}

func TestPlanAISegmentEpisodesNormalizesBadSuggestionsAndMergesShortTail(t *testing.T) {
	clips := make([]aiSegmentClip, 8)
	for i := range clips {
		clips[i] = aiSegmentClip{Index: i + 1, StartUS: int64(i) * 30_000_000, EndUS: int64(i+1) * 30_000_000}
	}
	constraints := defaultAISegmentConstraints()
	constraints.FrontCount = 1
	episodes, warnings := planAISegmentEpisodes(clips, []aiSegmentSuggestion{
		{EndClip: 4, Reason: "第一次冲突收束"},
		{EndClip: 4, Reason: "重复"},
		{EndClip: 99, Reason: "越界"},
		{EndClip: 6, Reason: "新冲突"},
	}, constraints)
	if len(episodes) != 3 {
		t.Fatalf("episodes=%+v warnings=%v", episodes, warnings)
	}
	if episodes[0].StartClip != 1 || episodes[0].EndClip != 4 || episodes[2].EndClip != 8 {
		t.Fatalf("episodes do not cover every clip: %+v", episodes)
	}
	if episodes[2].DurationSeconds != 60 {
		t.Fatalf("tail duration=%v, want 60", episodes[2].DurationSeconds)
	}
}

func TestPlanAISegmentEpisodesKeepsOversizedClipAndWarns(t *testing.T) {
	clips := []aiSegmentClip{{Index: 1, StartUS: 0, EndUS: 120_000_000}, {Index: 2, StartUS: 120_000_000, EndUS: 170_000_000}}
	episodes, warnings := planAISegmentEpisodes(clips, nil, aiSegmentConstraints{FrontCount: 0, BackMin: 45, BackTarget: 60, BackMax: 90})
	if len(episodes) != 2 || episodes[0].EndClip != 1 {
		t.Fatalf("episodes=%+v", episodes)
	}
	if len(warnings) == 0 || !strings.Contains(strings.Join(warnings, " "), "单个主轨片段") {
		t.Fatalf("warnings=%v", warnings)
	}
}

type aiSegmentJSONProvider struct {
	requests []batchfactoryv11.TextCompletionRequest
}

func (p *aiSegmentJSONProvider) Complete(_ context.Context, request batchfactoryv11.TextCompletionRequest) (string, error) {
	p.requests = append(p.requests, request)
	return `{"episodes":[{"end_clip":2,"reason":"转折"}]}`, nil
}

func TestAISegmentAnalyzeRouteUsesBackendPromptAndReturnsValidatedEpisodes(t *testing.T) {
	provider := &aiSegmentJSONProvider{}
	mux := http.NewServeMux()
	registerPluginToolRoutes(mux, BridgeAuth{Secret: "secret"}, provider)
	body := `{"mode":"ai","modelId":"text-model","script":"完整剧本","clips":[` +
		`{"index":1,"startUs":0,"endUs":50000000,"subtitle":"开场"},` +
		`{"index":2,"startUs":50000000,"endUs":110000000,"subtitle":"冲突"},` +
		`{"index":3,"startUs":110000000,"endUs":170000000,"subtitle":"转折"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/plugins/ai-segment/analyze", strings.NewReader(body))
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"episodes"`) || !strings.Contains(rec.Body.String(), `"endClip":3`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(provider.requests) != 1 || provider.requests[0].SystemPrompt == "" || strings.Contains(provider.requests[0].UserPrompt, "你是") {
		t.Fatalf("requests=%+v", provider.requests)
	}
}

func TestAISegmentAnalyzeRouteRejectsEmptyAndBrokenInputs(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want string
	}{
		{"empty script", `{"mode":"ai","modelId":"m","clips":[{"index":1,"startUs":0,"endUs":50000000}]}`, "请先提供剧本文稿"},
		{"empty model", `{"mode":"ai","script":"x","clips":[{"index":1,"startUs":0,"endUs":50000000}]}`, "请选择模型"},
		{"empty srt", `{"mode":"subtitle","srt":""}`, "没有读取到有效字幕"},
		{"broken draft", `{"mode":"main-track","draftContent":"not-json"}`, "剪映草稿无法解析"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			registerPluginToolRoutes(mux, BridgeAuth{Secret: "secret"}, &aiSegmentJSONProvider{})
			req := httptest.NewRequest(http.MethodPost, "/api/plugins/ai-segment/analyze", strings.NewReader(tc.body))
			SignBridgeRequest(req, "alice", false, time.Now(), "secret")
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), tc.want) {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestBuildAISegmentFFmpegArgsUsesExactEpisodeRange(t *testing.T) {
	args := strings.Join(buildAISegmentFFmpegArgs("input.mp4", "output.mp4", aiSegmentEpisode{StartSeconds: 12.5, EndSeconds: 47}, "1080p", 30), " ")
	for _, expected := range []string{"-ss 12.500", "-t 34.500", "scale=-2:1080,fps=30", "-c:v libx264", "output.mp4"} {
		if !strings.Contains(args, expected) {
			t.Fatalf("args=%s missing %s", args, expected)
		}
	}
}

func TestAISegmentExportRejectsMissingVideo(t *testing.T) {
	mux := http.NewServeMux()
	registerPluginToolRoutes(mux, BridgeAuth{Secret: "secret"}, &aiSegmentJSONProvider{})
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("episodes", `[{"startSeconds":0,"endSeconds":1}]`)
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/plugins/ai-segment/export", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "请选择视频文件") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
