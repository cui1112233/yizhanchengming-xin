package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type voiceDesignTestTransport func(*http.Request) (*http.Response, error)

func TestScreenplayVoiceDesignModelOptionsUseOwnerConfiguredServices(t *testing.T) {
	s, p := positioningFixture(t)
	ctx := context.Background()
	s.Catalog = novelfetchworkshop.NewMemoryStore()
	calls := 0
	s.VoiceDesigner = func(context.Context, string, string, screenplayVoiceDesignInput) (screenplayVoiceDesignResult, error) {
		calls++
		return screenplayVoiceDesignResult{}, errors.New("must not call supplier while listing")
	}
	if got := s.voiceDesignModelOptions(ctx, p.OwnerUsername); len(got) != 0 {
		t.Fatal(got)
	}
	_ = s.Catalog.PutConfig(ctx, p.OwnerUsername, map[string]any{"modelCatalog": []any{
		map[string]any{"kind": "audio", "enabled": true, "endpoint": "https://tts.example/v1/audio/speech", "credential": "mock-tts", "modelId": "ordinary-tts"},
		map[string]any{"kind": "audio", "enabled": true, "endpoint": "https://api.minimax.cn/v1/t2a_v2", "credential": "mock-secret", "modelId": "speech-2.6-hd"},
		map[string]any{"kind": "audio", "enabled": true, "endpoint": "https://dashscope.aliyuncs.com/api/v1", "credential": "mock-qwen", "modelId": "qwen-audio-3.1-tts-vd-2025"},
	}})
	got := s.voiceDesignModelOptions(ctx, p.OwnerUsername)
	if len(got) != 2 || got[0]["id"] != "minimax" || got[1]["id"] != "qwen-tts" {
		t.Fatal(got)
	}
	encoded, _ := json.Marshal(got)
	if strings.Contains(string(encoded), "secret") || strings.Contains(string(encoded), "https:") || strings.Contains(string(encoded), "ordinary-tts") {
		t.Fatal("private configuration or TTS leaked", string(encoded))
	}
	if other := s.voiceDesignModelOptions(ctx, "bob"); len(other) != 0 {
		t.Fatal("other owner sees services", other)
	}
	s.VoiceDesigner = nil
	if unavailable := s.voiceDesignModelOptions(ctx, p.OwnerUsername); len(unavailable) != 0 {
		t.Fatal("missing executor still selectable", unavailable)
	}
	if calls != 0 {
		t.Fatal("listing called supplier", calls)
	}
}

func (f voiceDesignTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestScreenplayVoiceDesignOfficialHTTPProtocols(t *testing.T) {
	for _, engine := range []string{"minimax", "qwen-tts"} {
		t.Run(engine, func(t *testing.T) {
			calls := 0
			audio := []byte("mock-preview-audio")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Header.Get("Authorization") != "Bearer mock-secret" || r.Method != "POST" {
					t.Error("missing authenticated JSON request")
				}
				var body map[string]any
				if e := json.NewDecoder(r.Body).Decode(&body); e != nil {
					t.Fatal(e)
				}
				w.Header().Set("Content-Type", "application/json")
				if engine == "minimax" {
					if r.URL.Path != "/v1/voice_design" || body["prompt"] != "低沉沉稳的中年男声" || body["preview_text"] != "这是用于试听的一段明确文本，请按角色的音色自然朗读。" {
						t.Errorf("body=%v path=%s", body, r.URL.Path)
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"voice_id": "voice-minimax", "trial_audio": hex.EncodeToString(audio), "base_resp": map[string]any{"status_code": 0}})
				} else {
					input, _ := body["input"].(map[string]any)
					params, _ := body["parameters"].(map[string]any)
					if r.URL.Path != "/api/v1/services/audio/tts/customization" || body["model"] != "voice-enrollment" || input["action"] != "create_voice" || input["target_model"] != "qwen-audio-3.1-tts-flash" || input["voice_prompt"] != "低沉沉稳的中年男声" || params["response_format"] != "wav" {
						t.Errorf("body=%v path=%s", body, r.URL.Path)
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"output": map[string]any{"voice_id": "voice-qwen", "target_model": "qwen-audio-3.1-tts-flash", "preview_audio": map[string]any{"data": base64.StdEncoding.EncodeToString(audio), "response_format": "wav"}}})
				}
			}))
			defer server.Close()
			local, _ := url.Parse(server.URL)
			endpoint := "https://api.minimax.cn/v1/voice_design"
			if engine == "qwen-tts" {
				endpoint = "https://dashscope.aliyuncs.com/api/v1/services/audio/tts/customization"
			}
			client := &http.Client{Timeout: 10 * time.Second, Transport: voiceDesignTestTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Scheme != "https" || (r.URL.Host != "api.minimax.cn" && r.URL.Host != "dashscope.aliyuncs.com") {
					t.Fatal("unexpected official target", r.URL)
				}
				clone := r.Clone(r.Context())
				u := *r.URL
				u.Scheme = local.Scheme
				u.Host = local.Host
				clone.URL = &u
				return http.DefaultTransport.RoundTrip(clone)
			})}
			result, e := requestScreenplayVoiceDesign(context.Background(), screenplayVoiceDesignTarget{Engine: engine, Endpoint: endpoint, Credential: "mock-secret", TargetModel: "qwen-audio-3.1-tts-flash"}, screenplayVoiceDesignInput{Prompt: "低沉沉稳的中年男声", PreviewText: "这是用于试听的一段明确文本，请按角色的音色自然朗读。", RequestKey: "mock-once"}, client)
			if e != nil || string(result.Audio) != string(audio) || result.VoiceID == "" || result.Engine != engine || calls != 1 {
				t.Fatal(result, e, calls)
			}
		})
	}
}

func TestScreenplayVoiceDesignBlocksNewKeyWhileSameTargetAndEngineRunning(t *testing.T) {
	s, p := positioningFixture(t)
	ctx := context.Background()
	_, _ = s.Projects.CreateAsset(ctx, screenplay.Asset{ID: "hero-running", OwnerUsername: p.OwnerUsername, ProjectID: p.ID, Name: "甲", Type: screenplay.AssetCharacter})
	s.Catalog = novelfetchworkshop.NewMemoryStore()
	_ = s.Catalog.PutConfig(ctx, p.OwnerUsername, map[string]any{"modelCatalog": []any{map[string]any{"kind": "audio", "enabled": true, "endpoint": "https://api.minimax.cn", "credential": "mock-only"}}})
	_, e := s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
		st.VoiceDesign = &screenplay.VoiceDesignState{Items: []screenplay.VoiceDesignDraft{{ID: 1, Name: "甲", AssetID: "hero-running"}}}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	var calls atomic.Int32
	s.VoiceDesigner = func(context.Context, string, string, screenplayVoiceDesignInput) (screenplayVoiceDesignResult, error) {
		count := calls.Add(1)
		if count == 1 {
			started <- struct{}{}
			<-release
		}
		return screenplayVoiceDesignResult{VoiceID: fmt.Sprintf("voice-%d", count), Engine: "minimax", ContentType: "audio/mpeg", Audio: []byte("ID3mock-preview")}, nil
	}
	st, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	first := screenplayMediaAction{TargetID: "voice-draft:1", Prompt: "音色描述", Text: "试听文本", ExpectedRevision: st.Revision, IdempotencyKey: "first"}
	done := make(chan error, 1)
	go func() { _, e := s.designVoice(ctx, p, first); done <- e }()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("first mocked request did not start")
	}
	latest, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	second := first
	second.TargetID = "hero-running"
	second.ExpectedRevision = latest.Revision
	second.IdempotencyKey = "changed-key"
	_, secondErr := s.designVoice(ctx, p, second)
	callCount := calls.Load()
	close(release)
	firstErr := <-done
	if !errors.Is(secondErr, screenplay.ErrInvalidState) || callCount != 1 || firstErr != nil {
		t.Fatal(secondErr, callCount, firstErr)
	}
	latest, _ = s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	attempts := 0
	for _, item := range latest.Items {
		if item.Kind == "voice_design" {
			attempts++
			if item.VoiceEngine != "minimax" {
				t.Fatal("running attempt did not retain engine", item)
			}
		}
	}
	if attempts != 1 {
		t.Fatal("duplicate attempt persisted", attempts)
	}
	second.ExpectedRevision = latest.Revision
	second.IdempotencyKey = "after-terminal"
	if _, e = s.designVoice(ctx, p, second); e != nil || calls.Load() != 2 {
		t.Fatal("terminal attempt incorrectly blocked a new request", e, calls.Load())
	}
}

func TestScreenplayVoiceDesignRunningGuardHandlesLegacyUnknownEngine(t *testing.T) {
	st := screenplay.MediaState{Items: []screenplay.MediaItem{{ID: "legacy", Kind: "voice_design", TargetID: "hero", Status: "running"}}}
	if e := screenplayVoiceAlreadyRunning(st, "hero", "minimax"); !errors.Is(e, screenplay.ErrInvalidState) {
		t.Fatal(e)
	}
	st.Items[0].VoiceEngine = "minimax"
	if e := screenplayVoiceAlreadyRunning(st, "hero", "qwen-tts"); e != nil {
		t.Fatal("independent engine blocked", e)
	}
	st.Items[0].Status = "failed"
	if e := screenplayVoiceAlreadyRunning(st, "hero", "minimax"); e != nil {
		t.Fatal("failed attempt blocked", e)
	}
}

func TestScreenplayVoiceDesignOwnerConfigurationAndPersistentIdempotency(t *testing.T) {
	s, p := positioningFixture(t)
	ctx := context.Background()
	_, _ = s.Projects.CreateAsset(ctx, screenplay.Asset{ID: "voice-hero", OwnerUsername: p.OwnerUsername, ProjectID: p.ID, Type: screenplay.AssetCharacter, Name: "甲"})
	s.Catalog = novelfetchworkshop.NewMemoryStore()
	_ = s.Catalog.PutConfig(ctx, p.OwnerUsername, map[string]any{"modelCatalog": []any{map[string]any{"kind": "audio", "enabled": true, "endpoint": "https://api.minimax.cn/v1/t2a_v2", "credential": "mock-secret", "modelId": "speech-2.6-hd"}}})
	if _, e := resolveScreenplayVoiceDesignTarget(ctx, s.Catalog, "bob", "minimax"); e == nil {
		t.Fatal("other owner's configured service accepted")
	}
	calls := 0
	audio := []byte("ID3mock-preview")
	s.VoiceDesigner = func(_ context.Context, owner, engine string, in screenplayVoiceDesignInput) (screenplayVoiceDesignResult, error) {
		calls++
		if owner != "alice" || engine != "minimax" || in.Prompt != "清亮少女音" {
			t.Fatal(owner, engine, in.Prompt)
		}
		return screenplayVoiceDesignResult{VoiceID: "designed-voice", Engine: engine, TargetModel: "speech-2.6-hd", ContentType: "audio/mpeg", Audio: audio}, nil
	}
	st, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	in := screenplayMediaAction{Action: "voice_design", TargetID: "voice-hero", Prompt: "清亮少女音", Text: "请用当前角色的声音自然地朗读这一段试听文本。", IdempotencyKey: "once", ExpectedRevision: st.Revision}
	result, e := s.designVoice(ctx, p, in)
	if e != nil {
		t.Fatal(e)
	}
	item := result.(screenplay.MediaItem)
	if item.VoiceID != "designed-voice" || item.VoiceEngine != "minimax" || item.VoiceDescription != in.Prompt || item.VoicePreviewText != in.Text || item.Kind != "audio" {
		t.Fatal(item)
	}
	asset, e := s.Cut.Asset(ctx, p.OwnerUsername, p.ID, item.AssetID)
	if e != nil {
		t.Fatal(e)
	}
	reader, e := s.Objects.Open(ctx, aicutObjectAsset(asset))
	if e != nil {
		t.Fatal(e)
	}
	saved, _ := io.ReadAll(reader)
	reader.Close()
	if string(saved) != string(audio) {
		t.Fatal("preview audio not persisted")
	}
	again, e := s.designVoice(ctx, p, in)
	if e != nil || again.(screenplay.MediaItem).VoiceID != item.VoiceID || calls != 1 {
		t.Fatal(again, e, calls)
	}
	st, _ = s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if st.Voices[in.TargetID] != "" {
		t.Fatal("designed voice automatically bound to ordinary TTS")
	}
	in.IdempotencyKey = "foreign"
	in.TargetID = "foreign-role"
	in.ExpectedRevision = st.Revision
	if _, e = s.designVoice(ctx, p, in); e == nil || calls != 1 {
		t.Fatal("foreign target called provider", e, calls)
	}
}

func TestScreenplayVoiceDesignRejectsUnofficialServiceAndCancelledLateAudio(t *testing.T) {
	ctx := context.Background()
	catalog := novelfetchworkshop.NewMemoryStore()
	_ = catalog.PutConfig(ctx, "alice", map[string]any{"modelCatalog": []any{map[string]any{"kind": "audio", "enabled": true, "endpoint": "https://attacker.example/v1/voice_design", "credential": "mock-only"}}})
	if _, e := resolveScreenplayVoiceDesignTarget(ctx, catalog, "alice", "minimax"); e == nil {
		t.Fatal("unofficial origin accepted")
	}
	s, p := positioningFixture(t)
	_, _ = s.Projects.CreateAsset(ctx, screenplay.Asset{ID: "hero", OwnerUsername: p.OwnerUsername, ProjectID: p.ID, Name: "甲", Type: screenplay.AssetCharacter})
	s.Catalog = novelfetchworkshop.NewMemoryStore()
	_ = s.Catalog.PutConfig(ctx, "alice", map[string]any{"modelCatalog": []any{map[string]any{"kind": "audio", "enabled": true, "endpoint": "https://api.minimax.cn", "credential": "mock-only"}}})
	cancelCtx, cancel := context.WithCancel(ctx)
	s.VoiceDesigner = func(context.Context, string, string, screenplayVoiceDesignInput) (screenplayVoiceDesignResult, error) {
		cancel()
		return screenplayVoiceDesignResult{VoiceID: "late", Engine: "minimax", Audio: []byte("ID3late"), ContentType: "audio/mpeg"}, nil
	}
	st, _ := s.Media.LoadMedia(ctx, "alice", p.ID)
	_, e := s.designVoice(cancelCtx, p, screenplayMediaAction{TargetID: "hero", Prompt: "音色", Text: "试听文本", IdempotencyKey: "cancel", ExpectedRevision: st.Revision})
	if !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	st, _ = s.Media.LoadMedia(ctx, "alice", p.ID)
	for _, m := range st.Items {
		if m.Status == "succeeded" {
			t.Fatal("late audio saved successful", m)
		}
	}
}

func TestScreenplayVoiceAnalysisOriginal127AndAtomicCancellation(t *testing.T) {
	if digest := fmt.Sprintf("%x", sha256.Sum256([]byte(prompt.ScreenplayVoiceDesignAnalysis127))); digest != "2020cc394de1b2d5917dbec74cb5c584af889bb56b3c0c83b713b39b39505959" {
		t.Fatal("original127 changed", digest)
	}
	s, p := positioningFixture(t)
	ctx := context.Background()
	_, _ = s.Projects.CreateAsset(ctx, screenplay.Asset{ID: "hero", OwnerUsername: p.OwnerUsername, ProjectID: p.ID, Name: "甲", Type: screenplay.AssetCharacter})
	calls := 0
	preview := strings.Repeat("这是角色的代表性试听台词。", 6)
	payload, _ := json.Marshal([]screenplay.VoiceDesignDraft{{ID: 1, Name: "甲", Role: "主角", Prompt: "清亮少年音", PreviewText: preview}})
	s.PositioningText = func(_ context.Context, owner string, in screenplay.TextRequest) (screenplay.TextResponse, error) {
		calls++
		if owner != "alice" || in.Model != "text-selected" || in.SystemPrompt != prompt.ScreenplayVoiceDesignAnalysis127 || in.UserPrompt != p.SourceText || in.JSONResponse {
			t.Fatal(owner, in.Model, in.SystemPrompt)
		}
		return screenplay.TextResponse{Content: string(payload)}, nil
	}
	st, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	in := screenplayMediaAction{Action: "voice_analyze", ModelID: "text-selected", IdempotencyKey: "original127", ExpectedRevision: st.Revision}
	result, e := s.analyzeVoices(ctx, p, in)
	if e != nil {
		t.Fatal(e)
	}
	state := result.(*screenplay.VoiceDesignState)
	if len(state.Items) != 1 || state.Items[0].AssetID != "hero" || state.Items[0].Prompt != "清亮少年音" || state.Source != "project_original" {
		t.Fatal(state)
	}
	if _, e = s.analyzeVoices(ctx, p, in); e != nil || calls != 1 {
		t.Fatal(e, calls)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	s.PositioningText = func(context.Context, string, screenplay.TextRequest) (screenplay.TextResponse, error) {
		cancel()
		return screenplay.TextResponse{Content: `[]`}, nil
	}
	st, _ = s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	in.IdempotencyKey = "cancel127"
	in.ExpectedRevision = st.Revision
	if _, e = s.analyzeVoices(cancelCtx, p, in); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	st, _ = s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if st.VoiceDesign.Status != "cancelled" || len(st.VoiceDesign.Items) != 1 || st.VoiceDesign.Items[0].Name != "甲" {
		t.Fatal(st.VoiceDesign)
	}
}

func TestScreenplayVoiceDraftEditsPersistWithoutModelAndIgnoreForeignBinding(t *testing.T) {
	s, p := positioningFixture(t)
	ctx := context.Background()
	_, _ = s.Projects.CreateAsset(ctx, screenplay.Asset{ID: "hero", OwnerUsername: p.OwnerUsername, ProjectID: p.ID, Name: "甲", Type: screenplay.AssetCharacter})
	s.VoiceDesigner = func(context.Context, string, string, screenplayVoiceDesignInput) (screenplayVoiceDesignResult, error) {
		t.Fatal("saving drafts called voice provider")
		return screenplayVoiceDesignResult{}, nil
	}
	s.PositioningText = func(context.Context, string, screenplay.TextRequest) (screenplay.TextResponse, error) {
		t.Fatal("saving drafts called text model")
		return screenplay.TextResponse{}, nil
	}
	st, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	in := screenplayMediaAction{ExpectedRevision: st.Revision, VoiceDesignDrafts: []screenplay.VoiceDesignDraft{{ID: 1, Name: "甲", Role: "主角", Prompt: "修改后的声音", PreviewText: "用户试听", AssetID: "foreign-owner"}, {ID: 2, Name: "旁白", Prompt: "", PreviewText: ""}}}
	_, e := s.saveVoiceDesignDrafts(ctx, p, in)
	if e != nil {
		t.Fatal(e)
	}
	loaded, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if loaded.VoiceDesign.Status != "draft" || loaded.VoiceDesign.Items[0].Prompt != "修改后的声音" || loaded.VoiceDesign.Items[0].AssetID != "hero" || loaded.VoiceDesign.Items[1].AssetID != "" {
		t.Fatal(loaded.VoiceDesign)
	}
	if _, e = s.saveVoiceDesignDrafts(ctx, p, in); !errors.Is(e, screenplay.ErrRevisionConflict) {
		t.Fatal(e)
	}
}

func TestScreenplayVoiceAnalysisUsesCurrentChapterOriginalRatherThanOldProjectSource(t *testing.T) {
	s, p := positioningFixture(t)
	ctx := context.Background()
	s.Authoring = &screenplay.AuthoringService{Store: s.Projects, Repository: screenplay.NewMemoryAuthoringRepository()}
	_, e := s.Authoring.Save(ctx, p.OwnerUsername, p.ID, screenplay.AuthoringState{Chapters: []screenplay.Chapter{{ID: "chapter1", Title: "第一章", SourceText: "用户刚编辑的章节原文", ScriptText: "生成剧本不应该送给127"}}}, 0)
	if e != nil {
		t.Fatal(e)
	}
	s.PositioningText = func(_ context.Context, _ string, in screenplay.TextRequest) (screenplay.TextResponse, error) {
		if !strings.Contains(in.UserPrompt, "用户刚编辑的章节原文") || strings.Contains(in.UserPrompt, p.SourceText) || strings.Contains(in.UserPrompt, "生成剧本不应该") {
			t.Fatal(in.UserPrompt)
		}
		return screenplay.TextResponse{Content: `[]`}, nil
	}
	st, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	result, e := s.analyzeVoices(ctx, p, screenplayMediaAction{ModelID: "text-model", ExpectedRevision: st.Revision, IdempotencyKey: "chapter-original"})
	if e != nil || result.(*screenplay.VoiceDesignState).Source != "authoring_chapters" {
		t.Fatal(result, e)
	}
}
