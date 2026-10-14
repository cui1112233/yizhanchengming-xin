package batchfactoryv11

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type queuedDirectorProvider struct {
	values []string
	calls  []TextCompletionRequest
}

func (p *queuedDirectorProvider) Complete(_ context.Context, input TextCompletionRequest) (string, error) {
	p.calls = append(p.calls, input)
	if len(p.values) == 0 {
		return "", ErrUnavailable
	}
	value := p.values[0]
	p.values = p.values[1:]
	return value, nil
}

func rawSetting(t *testing.T, value any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestDirectorRetriesOnceWhenGatewayReturnsNonJSONText(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	provider := &queuedDirectorProvider{values: []string{"先解释一下", validDirectorJSON()}}
	service := &DirectorService{Store: store, Provider: provider}
	if _, err := service.RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatalf("director should recover from one formatting-only response: %v", err)
	}
	if len(provider.calls) != 2 || !provider.calls[1].DisableJSONResponse {
		t.Fatalf("retry request should disable response_format: calls=%+v", provider.calls)
	}
}

func seedDirectorBook(t *testing.T, mode string, fixed bool) (*MemoryStore, Batch, Book) {
	t.Helper()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "B", Books: []CreateBookInput{{Title: "K", SourceText: "她被当众羞辱后沉默离开。", Videos: []CreateVideoInput{{Label: "old"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	patch := SettingsPatch{"productionMode": rawSetting(t, mode), "maxVideoDuration": rawSetting(t, 15), "fixedSingleVideo": rawSetting(t, fixed), "aspectRatio": rawSetting(t, "9:16")}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{Patch: patch, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	return store, batch, batch.Books[0]
}

func TestDirectorDefaultsStoryboardDurationLimitToTenSeconds(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "B", Books: []CreateBookInput{{Title: "K", SourceText: "她转身离开。"}}})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := snapshotForBook(batch, batch.Books[0])
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.MaxVideoDuration != 10 {
		t.Fatalf("default storyboard duration limit=%d, want 10", snapshot.MaxVideoDuration)
	}
}

func TestBuildAssetExtractionContractIncludesSelectedCharacterAndSceneOutputRules(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "林晚走进客厅。"}
	snapshot := DirectorSnapshot{Effective: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
		"assets": map[string]any{
			"enabled":   true,
			"scope":     "all",
			"character": map[string]any{"body": "CHARACTER OUTPUT RULE"},
			"scene":     map[string]any{"body": "SCENE OUTPUT RULE"},
		},
	})}}
	contract, err := BuildAssetExtractionContract(book, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"CHARACTER OUTPUT RULE", "SCENE OUTPUT RULE"} {
		if !strings.Contains(contract.SystemPrompt, want) {
			t.Fatalf("asset extraction contract missing %q: %s", want, contract.SystemPrompt)
		}
	}
}

func TestBuildAssetExtractionContractCarriesSavedStyleSystemIntoSelectedAssetRule(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "林晚走进客厅。"}
	snapshot := DirectorSnapshot{Effective: SettingsPatch{
		"h3StyleAnalysis": rawSetting(t, `{"schema_version":"h3-style-system/v1","prompt":"现代都市短剧；高级电影感；暖黄室内光。","fields":{"final_genre":"都市情感","trailer_style":"高级电影感","story_era":"现代都市"},"preset":{"id":"script-constraint-prefix-smart-unified","version":1}}`),
		"aiPromptConfig": rawSetting(t, map[string]any{
			"assets": map[string]any{
				"enabled":    true,
				"scope":      "all",
				"extraction": map[string]any{"body": "CUSTOM ASSET RULE"},
			},
		}),
	}}
	contract, err := BuildAssetExtractionContract(book, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contract.SystemPrompt, "CUSTOM ASSET RULE") {
		t.Fatalf("asset extraction system missing selected rule: %s", contract.SystemPrompt)
	}
	if !strings.Contains(contract.UserPrompt, "现代都市短剧；高级电影感；暖黄室内光。") {
		t.Fatalf("asset extraction input missing saved style result: %s", contract.UserPrompt)
	}
}

func TestBuildAssetExtractionContractRequiresTheFourSDAssetAuthorities(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"assets": map[string]any{"enabled": true, "extraction": map[string]any{
				"presetId": "batch-assets-sd-character-scene", "presetKey": "sd-character-scene", "body": "SD ASSET RULE",
			}},
		})}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	contract, err := BuildAssetExtractionContract(book, DirectorSnapshot{Effective: batch.SettingsState.Patch})
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"正式人物名单", "人物事实", "人物外形", "人物关系", "一次调用", "合并"} {
		if !strings.Contains(contract.SystemPrompt, required) {
			t.Fatalf("SD asset contract missing %q:\n%s", required, contract.SystemPrompt)
		}
	}
}

func TestDirectorReadbackUsesShotTimelineUnlessTheUserEditedThatVideo(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	provider := &queuedDirectorProvider{values: []string{validDirectorJSON()}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	video := latest.Books[0].Videos[0]
	if !strings.Contains(video.VideoPrompt, "00:00-00:03 | 中景｜缓慢推轨 | 林晚进入客厅") || strings.Contains(video.VideoPrompt, "林晚进入客厅并握紧玻璃杯。") {
		t.Fatalf("readback prompt must use the shot timeline: %q", video.VideoPrompt)
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{
		Patch: SettingsPatch{"videoPrompt": rawSetting(t, "用户编辑后的当前分镜")}, ExpectedRevision: video.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	latest, err = store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := latest.Books[0].Videos[0].VideoPrompt; got != "用户编辑后的当前分镜" {
		t.Fatalf("manual video override=%q", got)
	}
}

func TestBuildDirectorContractComposesPublicScriptStoryboardRules(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "林晚走进办公室，陆沉抬头看向她。"}
	config := map[string]any{
		"scriptComposition": map[string]any{
			"segmented":         map[string]any{"body": "PUBLIC SEGMENTED {duration}"},
			"shotlist":          map[string]any{"body": "PUBLIC SHOTLIST {duration}"},
			"general":           map[string]any{"body": "PUBLIC GENERAL {duration}"},
			"characterFocus":    map[string]any{"body": "PUBLIC FOCUS {focusCharacters} / {focusCount}"},
			"audioMatch":        map[string]any{"body": "PUBLIC AUDIO {audioDurationSec} / {unitMaxSec}"},
			"cardProtocol":      map[string]any{"body": "PUBLIC CARD {duration}"},
			"constraintWrapper": map[string]any{"body": "PUBLIC CONSTRAINT {duration}"},
		},
		"constraints": map[string]any{
			"enabled":    true,
			"selections": []any{map[string]any{"constraintCategory": "quality", "body": "QUALITY RULE"}},
		},
	}
	snapshot := DirectorSnapshot{
		Mode: "original", MaxVideoDuration: 10, AudioTargetSeconds: 63, AudioDurationSeconds: 62.5, AspectRatio: "9:16",
		Effective: SettingsPatch{
			"aiPromptConfig":        rawSetting(t, config),
			"starredCharacterNames": rawSetting(t, []string{"林晚", "陆沉"}),
		},
	}
	contract, err := BuildDirectorContract(book, HookRevision{}, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"PUBLIC SEGMENTED 10s", "PUBLIC SHOTLIST 10s", "PUBLIC GENERAL 10s",
		"PUBLIC FOCUS 林晚、陆沉 / 2", "PUBLIC AUDIO 62.5 / 10", "PUBLIC CARD 10s", "PUBLIC CONSTRAINT 10s", "QUALITY RULE",
	} {
		if !strings.Contains(contract.SystemPrompt, expected) {
			t.Fatalf("public script composition is missing %q:\n%s", expected, contract.SystemPrompt)
		}
	}
	if !strings.Contains(contract.SystemPrompt, "V11 JSON 存储适配") || !strings.Contains(contract.SystemPrompt, "不得输出任何 Markdown 标题") {
		t.Fatalf("public card protocol must be adapted to the V11 JSON boundary:\n%s", contract.SystemPrompt)
	}
}

func TestBuildDirectorContractDefersVisualPromptUntilVisualExtraction(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "正文"}
	config := map[string]any{
		"video":  map[string]any{"enabled": true, "scope": "all", "body": "VIDEO RULE"},
		"visual": map[string]any{"enabled": true, "scope": "all", "body": "VISUAL META"},
	}
	contract, err := BuildDirectorContract(book, HookRevision{}, DirectorSnapshot{
		Mode: "original", MaxVideoDuration: 10, AspectRatio: "9:16",
		Effective: SettingsPatch{"aiPromptConfig": rawSetting(t, config)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(contract.SystemPrompt, "VISUAL META") || strings.Contains(contract.SystemPrompt, "visual_prompt") {
		t.Fatalf("director must only generate storyboard video prompts, got:\n%s", contract.SystemPrompt)
	}
	if contract.Normalization.RequireVisualPrompt {
		t.Fatal("director must not require visual_prompt before visual extraction")
	}
}

func TestBuildDirectorContractKeepsSelectedVideoPromptWhenLegacySwitchWasOff(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "正文"}
	config := map[string]any{
		// Video prompts are now mandatory. Existing books may still carry the
		// retired enabled=false flag, which must not silently remove a selected
		// published video rule from the director contract.
		"video": map[string]any{"enabled": false, "scope": "all", "body": "MANDATORY VIDEO RULE"},
	}
	contract, err := BuildDirectorContract(book, HookRevision{}, DirectorSnapshot{
		Mode: "original", MaxVideoDuration: 10, AspectRatio: "9:16",
		Effective: SettingsPatch{"aiPromptConfig": rawSetting(t, config)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contract.SystemPrompt, "MANDATORY VIDEO RULE") {
		t.Fatalf("selected video rule was incorrectly disabled by the retired switch:\n%s", contract.SystemPrompt)
	}
}

func TestBuildDirectorContractUsesOnlySelectedNonH3VideoPresetRules(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "林晚走进客厅。"}
	config := map[string]any{
		"video": map[string]any{"presetId": "batch-video-custom", "presetKey": "h3-video-normal", "body": "CUSTOM VIDEO RULE"},
		"scriptComposition": map[string]any{
			"general":      map[string]any{"body": "H3 GENERAL RULE"},
			"cardProtocol": map[string]any{"body": "H3 CARD RULE"},
		},
		"constraints": map[string]any{"selections": []any{map[string]any{"body": "H3 CONSTRAINT RULE"}}},
	}
	contract, err := BuildDirectorContract(book, HookRevision{}, DirectorSnapshot{Mode: "original", MaxVideoDuration: 10, AspectRatio: "9:16", Effective: SettingsPatch{"aiPromptConfig": rawSetting(t, config)}})
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"H3 GENERAL RULE", "H3 CARD RULE", "H3 CONSTRAINT RULE", "scene_memory", "visual_context"} {
		if strings.Contains(contract.SystemPrompt, forbidden) {
			t.Fatalf("SD contract contains H3 rule %q:\n%s", forbidden, contract.SystemPrompt)
		}
	}
	if !strings.Contains(contract.SystemPrompt, "CUSTOM VIDEO RULE") {
		t.Fatal("selected non-H3 video rule is missing")
	}
}

func TestBuildDirectorContractRequiresH3CanonicalShotMetadata(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "林晚走进客厅。"}
	contract, err := BuildDirectorContract(book, HookRevision{}, DirectorSnapshot{
		Mode: "original", MaxVideoDuration: 10, AspectRatio: "9:16",
		Effective: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"video": map[string]any{
				"presetId":  "batch-video-h3-director",
				"presetKey": "h3-video-normal",
				"body":      "H3 VIDEO RULE",
			},
		})},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"scene_memory", "visual_context", "lighting", "rhythm", "audio", "H3 VIDEO RULE"} {
		if !strings.Contains(contract.SystemPrompt, expected) {
			t.Fatalf("H3 director contract missing %q:\n%s", expected, contract.SystemPrompt)
		}
	}
}

func TestBuildDirectorContractScalesOutputBudgetToPlanningScope(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "林晚走进客厅，陆沉转身看向她。"}
	tests := []struct {
		name     string
		snapshot DirectorSnapshot
		want     int
	}{
		{
			name:     "fixed single video avoids the multi-video output budget",
			snapshot: DirectorSnapshot{Mode: "original", MaxVideoDuration: 10, FixedSingleVideo: true, AspectRatio: "9:16"},
			want:     14000,
		},
		{
			name:     "ordinary storyboard keeps a complete but bounded output budget",
			snapshot: DirectorSnapshot{Mode: "original", MaxVideoDuration: 10, AspectRatio: "9:16"},
			want:     16000,
		},
		{
			name:     "audio-following planning retains the long-form output budget",
			snapshot: DirectorSnapshot{Mode: "original", MaxVideoDuration: 10, AudioTargetSeconds: 61, AspectRatio: "9:16"},
			want:     18000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contract, err := BuildDirectorContract(book, HookRevision{}, tt.snapshot)
			if err != nil {
				t.Fatal(err)
			}
			if contract.MaxTokens != tt.want {
				t.Fatalf("director max tokens=%d, want %d", contract.MaxTokens, tt.want)
			}
		})
	}
}

func TestDirectorDoesNotInjectDerivedOpeningPresetsIntoNormalStoryboard(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "正文"}
	config := map[string]any{
		"originalDirector": map[string]any{"enabled": true, "body": "LEGACY ORIGINAL DIRECTOR"},
		"viralDirector":    map[string]any{"enabled": true, "body": "LEGACY VIRAL DIRECTOR"},
	}
	contract, err := BuildDirectorContract(book, HookRevision{}, DirectorSnapshot{Mode: "original", MaxVideoDuration: 10, AspectRatio: "9:16", Effective: SettingsPatch{"aiPromptConfig": rawSetting(t, config)}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(contract.SystemPrompt, "LEGACY ORIGINAL DIRECTOR") || strings.Contains(contract.SystemPrompt, "LEGACY VIRAL DIRECTOR") {
		t.Fatalf("derived-opening presets must be used only by the working-front candidate flow: %s", contract.SystemPrompt)
	}
}

func TestDirectorRejectsAudioPlanningWithoutMeasuredDuration(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{"audioPlanningEnabled": rawSetting(t, true)}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = (&DirectorService{Store: store, Provider: &queuedDirectorProvider{values: []string{validDirectorJSON()}}}).RunDirector(context.Background(), "alice", batch.ID, book.ID)
	if err == nil || !strings.Contains(err.Error(), "音频规划") {
		t.Fatalf("missing measured audio must block director generation, err=%v", err)
	}
}

func TestDirectorIgnoresLegacyMaximumUntilStoryboardDurationIsSelected(t *testing.T) {
	_, batch, book := seedDirectorBook(t, "original", false)
	batch.SettingsState.Patch = SettingsPatch{"maxVideoDuration": rawSetting(t, 15)}
	snapshot, err := snapshotForBook(batch, book)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.MaxVideoDuration != 10 {
		t.Fatalf("legacy max duration must not opt into 15s storyboard splitting: %d", snapshot.MaxVideoDuration)
	}
}

func TestFixedSingleVideoUsesStoryboardDurationLimitInsteadOfLegacyExactDuration(t *testing.T) {
	_, batch, book := seedDirectorBook(t, "original", true)
	batch.SettingsState.Patch["storyboardDurationLimit"] = rawSetting(t, 10)
	batch.SettingsState.Patch["fixedVideoDuration"] = rawSetting(t, 5)
	snapshot, err := snapshotForBook(batch, book)
	if err != nil {
		t.Fatal(err)
	}
	contract, err := BuildDirectorContract(book, HookRevision{}, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contract.SystemPrompt, "1-10") || !strings.Contains(contract.SystemPrompt, "只生产 VIDEO01") || strings.Contains(contract.SystemPrompt, "严格等于") {
		t.Fatalf("fixed single video must use only the storyboard limit: %s", contract.SystemPrompt)
	}
}

func TestViralDirectorRequiresApprovedHookAndBehavioralEscalationContract(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "viral", false)
	provider := &queuedDirectorProvider{values: []string{"她猛地掀翻桌子，当众质问对方。", validDirectorJSON()}}
	service := &DirectorService{Store: store, Provider: provider}
	if _, err := service.RunDirector(context.Background(), "alice", batch.ID, book.ID); err == nil || !strings.Contains(err.Error(), "Hook") {
		t.Fatalf("expected Hook prerequisite, got %v", err)
	}
	hook, err := service.RunHook(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(provider.calls[0].SystemPrompt, "visible conflict") || !strings.Contains(provider.calls[0].SystemPrompt, "behavior escalation") {
		t.Fatal("Hook contract lost visible behavioral escalation")
	}
	if _, err := service.ApproveHook(context.Background(), "alice", batch.ID, book.ID, hook.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
}

func TestFixedSingleDirectorCreatesOneImmutableVideo(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", true)
	provider := &queuedDirectorProvider{values: []string{validDirectorJSON()}}
	service := &DirectorService{Store: store, Provider: provider}
	revision, err := service.RunDirector(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(revision.Videos) != 1 || revision.Videos[0].ID == "" {
		t.Fatalf("revision=%+v", revision)
	}
	if revision.Videos[0].DurationSeconds != 9 {
		t.Fatalf("duration=%v", revision.Videos[0].DurationSeconds)
	}
}

func TestNormalDirectorPayloadIncludesExistingAssets(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	assetProvider := &queuedDirectorProvider{values: []string{`{"characters":[{"name":"林溪","prompt":"短发女主，右耳一颗小痣"}],"scenes":[{"name":"客厅","prompt":"现代中式客厅"}],"props":[{"name":"手机","prompt":"黑色手机"}]}`}}
	if _, err := (&DirectorService{Store: store, Provider: assetProvider}).RunAssetExtraction(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	directorProvider := &queuedDirectorProvider{values: []string{validDirectorJSON()}}
	if _, err := (&DirectorService{Store: store, Provider: directorProvider}).RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	if len(directorProvider.calls) != 1 {
		t.Fatalf("director calls=%d", len(directorProvider.calls))
	}
	userPrompt := directorProvider.calls[0].UserPrompt
	for _, expected := range []string{`"assets"`, "林溪", "短发女主，右耳一颗小痣", "现代中式客厅"} {
		if !strings.Contains(userPrompt, expected) {
			t.Fatalf("director user prompt missing existing asset %q:\n%s", expected, userPrompt)
		}
	}
	if !strings.Contains(directorProvider.calls[0].SystemPrompt, "逐字照抄清单") {
		t.Fatal("director system prompt must instruct copying existing assets verbatim")
	}
}

func TestAssetExtractionDoesNotReplaceExistingDirectorVideos(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	provider := &queuedDirectorProvider{values: []string{`{"characters":[{"name":"林溪","prompt":"短发女主"}],"scenes":[{"name":"客厅","prompt":"现代客厅"}],"props":[{"name":"手机","prompt":"黑色手机"}]}`}}
	service := &DirectorService{Store: store, Provider: provider}
	assets, err := service.RunAssetExtraction(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 3 {
		t.Fatalf("assets=%+v", assets)
	}
	read, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if read.Books[0].DirectorRevision != nil || len(read.Books[0].Videos) != len(book.Videos) {
		t.Fatalf("asset extraction must preserve director/video state: %+v", read.Books[0])
	}
}

func TestAssetExtractionDoesNotRequireAudioPlanningMeasurement(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{"audioPlanningEnabled": rawSetting(t, true)}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{`{"characters":[{"name":"林溪","prompt":"短发女主"}],"scenes":[],"props":[]}`}}
	assets, err := (&DirectorService{Store: store, Provider: provider}).RunAssetExtraction(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatalf("asset extraction must not wait for audio planning: %v", err)
	}
	if len(assets) != 1 || len(provider.calls) != 1 {
		t.Fatalf("assets=%+v calls=%d", assets, len(provider.calls))
	}
}

func TestAssetExtractionPersistsSmartUnifiedAnalysisWithSeparateAssetCall(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"constraints": map[string]any{
				"enabled": true,
				"scope":   "all",
				"selections": []map[string]any{{
					"presetId": "script-constraint-prefix-smart-unified", "presetName": "智能统一", "presetVersion": 1,
					"constraintCategory": "prefix", "body": "SMART UNIFIED RULE",
				}},
			},
		})}, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{
		`{"smart_unified_analysis":{"schema_version":"h3-style-system/v1","prompt":"现代都市短剧；暖黄室内光。","fields":{"final_genre":"都市情感","trailer_style":"电影感","story_era":"现代都市"},"preset":{"id":"script-constraint-prefix-smart-unified","name":"智能统一","version":1}}}`,
		`{"characters":[{"name":"林溪","prompt":"短发女主"}],"scenes":[{"name":"客厅","prompt":"现代客厅"}],"props":[]}`,
	}}
	assets, err := (&DirectorService{Store: store, Provider: provider}).RunAssetExtraction(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 2 || len(assets) != 2 {
		t.Fatalf("smart-unified and asset extraction must use two requests, calls=%d assets=%+v", len(provider.calls), assets)
	}
	if !strings.Contains(provider.calls[0].SystemPrompt, "smart_unified_analysis") || strings.Contains(provider.calls[1].SystemPrompt, "智能统一与资产同次返回") {
		t.Fatalf("smart-unified and asset prompts were not separated: %+v", provider.calls)
	}
	read, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := rawString(read.Books[0].SettingsState.Patch, "h3StyleAnalysis", ""); !strings.Contains(got, "现代都市短剧；暖黄室内光。") {
		t.Fatalf("combined smart-unified analysis was not persisted: %q", got)
	}
}

func TestAssetExtractionWithSDCombinedPresetUsesOneModelCall(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"assets": map[string]any{
				"enabled": true,
				"extraction": map[string]any{
					"presetId":  "batch-assets-sd-character-scene",
					"presetKey": "sd-character-scene",
					"body":      "SD ONE CALL ASSET RULE",
				},
			},
		})}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{`{"characters":[{"name":"林溪","prompt":"完整人物外形与视频锚点"}],"scenes":[{"name":"客厅","prompt":"完整场景锚点"}],"props":[{"name":"手机","prompt":"完整道具锚点"}]}`}}
	assets, err := (&DirectorService{Store: store, Provider: provider}).RunAssetExtraction(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 1 || len(assets) != 3 {
		t.Fatalf("SD combined asset flow calls=%d assets=%+v, want one request with all asset kinds", len(provider.calls), assets)
	}
	if !strings.Contains(provider.calls[0].SystemPrompt, "SD ONE CALL ASSET RULE") {
		t.Fatalf("SD combined preset body was not used: %q", provider.calls[0].SystemPrompt)
	}
}

func TestAssetExtractionWithH3FullPresetCompilesAllCharactersOnce(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"assets": map[string]any{
				"enabled":    true,
				"extraction": map[string]any{"presetId": "batch-assets-h3", "presetKey": "h3-assets-full", "body": "# H3 人物场景道具提取\n\n## 调用一：H3 人物场景道具事实提取\nH3 FACT RULE\n\n## 调用二：H3 全人物外形编译\nH3 APPEARANCE RULE"},
			},
		})}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{
		`{"characters":[{"name":"林晚","prompt":"H3 人物设定：林晚，二十五岁左右的都市女性，身形纤细挺拔，椭圆脸，眉眼清秀，鼻梁细直，唇色自然。黑色长发柔顺披至肩胛，发尾微弯。穿米白色真丝衬衫、浅灰高腰西装裤和米色细跟鞋，衣料光泽克制，配戴小颗珍珠耳钉。气质温和但克制，站姿保持肩背舒展，目光沉静。黑长发、珍珠耳钉与米白衬衫为跨镜头固定识别点。"},{"name":"陆沉","prompt":"H3 人物设定：陆沉，二十八岁左右的都市男性，身材高挑修长，肩线平直，轮廓分明的长方脸，眉骨清晰，鼻梁高挺，眼神沉稳疏离。短黑发整齐向后梳理，两鬓利落。穿深炭灰羊毛西装、白色棉质衬衫和黑色皮鞋，领口与袖口始终整洁，配戴银色窄表盘手表。动作简洁，不做夸张表情，微抿的唇线与冷静注视构成主要气质。深灰西装、银色手表和后梳短发为跨镜头固定识别点。"}],"scenes":[{"name":"客厅","prompt":"现代客厅，傍晚暖光，沙发与落地窗构成清晰空间关系。"}],"props":[]}`,
	}}
	assets, err := (&DirectorService{Store: store, Provider: provider}).RunAssetExtraction(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 1 {
		t.Fatalf("H3 character flow calls=%d, want one complete asset request", len(provider.calls))
	}
	if !strings.Contains(provider.calls[0].SystemPrompt, "H3 FACT RULE") {
		t.Fatalf("first H3 call must receive H3's first-phase rule, got %q", provider.calls[0].SystemPrompt)
	}
	if !strings.Contains(provider.calls[0].SystemPrompt, "H3 APPEARANCE RULE") {
		t.Fatalf("complete H3 asset request must contain its detailed appearance rule")
	}
	byName := map[string]string{}
	for _, asset := range assets {
		byName[asset.Name] = asset.Prompt
	}
	if got := byName["林晚"]; !strings.Contains(got, "黑长发、珍珠耳钉与米白衬衫为跨镜头固定识别点") {
		t.Fatalf("林晚 prompt=%q", got)
	}
	if got := byName["陆沉"]; !strings.Contains(got, "深灰西装、银色手表和后梳短发为跨镜头固定识别点") {
		t.Fatalf("陆沉 prompt=%q", got)
	}
}

func TestH3FullAssetPresetCarriesReferenceAnalysisAndAppearanceAuthorities(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "..", "prompts", "批量工厂-H3人物场景道具提示词.md"))
	if err != nil {
		t.Fatal(err)
	}
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"assets": map[string]any{
				"enabled":    true,
				"extraction": map[string]any{"presetId": "batch-assets-h3", "presetKey": "h3-assets-full", "body": string(body)},
			},
		})}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{
		`{"characters":[{"name":"林晚","prompt":"亚洲二十五岁成年女性，现代都市女主，椭圆脸与清晰眉眼，乌黑长发，皮肤白皙自然，身形纤细挺拔，穿剪裁利落的浅色西装与真丝内搭，材质和主色稳定，珍珠耳钉是跨镜头固定识别点；她的身份、性别、年龄阶段、时代与关系均由原文强证据确认。"}],"scenes":[],"props":[]}`,
	}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunAssetExtraction(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 1 {
		t.Fatalf("H3 calls=%d, want one complete H3 asset request", len(provider.calls))
	}
	for _, want := range []string{"V59 AUTO正式人物名单判断", "V59 第二步：只按强制名单生成人物卡", "V52通用小说人物关系图与分镜选角Skill", "V55人物卡外形隔离系统规则", "强证据", "普通姓名不能只凭字形猜性别"} {
		if !strings.Contains(provider.calls[0].SystemPrompt, want) {
			t.Fatalf("complete H3 asset request missing reference authority %q:\n%s", want, provider.calls[0].SystemPrompt)
		}
	}
	for _, want := range []string{"slot_id", "有效性别", "年龄阶段", "不得擅自改名", "脸型骨相", "鞋履", "同批人物"} {
		if !strings.Contains(provider.calls[0].SystemPrompt, want) {
			t.Fatalf("complete H3 asset request missing appearance authority %q:\n%s", want, provider.calls[0].SystemPrompt)
		}
	}
}

func TestAssetExtractionWithH3ScenePresetCompilesEachSceneIndividually(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"assets": map[string]any{
				"enabled": true,
				"scene":   map[string]any{"presetId": "batch-scene-h3", "presetKey": "h3-scene-normal", "body": "H3 SCENE RULE"},
			},
		})}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{
		`{"characters":[{"name":"林晚","prompt":"短发女主"}],"scenes":[{"name":"客厅","prompt":"提取的客厅"},{"name":"医院走廊","prompt":"提取的走廊"}],"props":[]}`,
		`{"prompt":"H3 场景设定：现代公寓客厅，傍晚暖光，人物活动空间清晰"}`,
		`{"prompt":"H3 场景设定：医院走廊，夜间冷白顶灯，空间纵深清晰"}`,
	}}
	assets, err := (&DirectorService{Store: store, Provider: provider}).RunAssetExtraction(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 3 {
		t.Fatalf("H3 scene flow calls=%d, want extraction plus one call per scene", len(provider.calls))
	}
	byName := map[string]string{}
	for _, asset := range assets {
		byName[asset.Name] = asset.Prompt
	}
	if got := byName["客厅"]; got != "H3 场景设定：现代公寓客厅，傍晚暖光，人物活动空间清晰" {
		t.Fatalf("客厅 prompt=%q", got)
	}
	if got := byName["医院走廊"]; got != "H3 场景设定：医院走廊，夜间冷白顶灯，空间纵深清晰" {
		t.Fatalf("医院走廊 prompt=%q", got)
	}
}

func TestReDirectorPreservesOldVideoOverrideAsOrphaned(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	old := book.Videos[0]
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: old.ID}, SettingsUpdate{Patch: SettingsPatch{"quality": rawSetting(t, "cinematic")}, ExpectedRevision: old.Revision}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{validDirectorJSON()}}
	revision, err := (&DirectorService{Store: store, Provider: provider}).RunDirector(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(revision.OrphanedOverrides) != 1 || revision.OrphanedOverrides[0].VideoID != old.ID || revision.OrphanedOverrides[0].State != "orphaned" {
		t.Fatalf("orphaned=%+v", revision.OrphanedOverrides)
	}
	if len(revision.Videos) != 1 || revision.Videos[0].ID == old.ID {
		t.Fatalf("video identity was reused: %+v", revision.Videos)
	}
}

func TestReDirectorPreservesManualAssetExclusionAndRefreshesAutomaticSuggestions(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	provider := &queuedDirectorProvider{values: []string{validDirectorJSON(), validDirectorJSON()}}
	service := &DirectorService{Store: store, Provider: provider}
	if _, err := service.RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	first, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	book = first.Books[0]
	video := book.Videos[0]
	var characterID string
	for _, asset := range book.AssetRecords {
		if asset.Kind == "character" && asset.Name == "林晚" {
			characterID = asset.ID
			break
		}
	}
	if characterID == "" {
		t.Fatalf("assets=%+v", book.AssetRecords)
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{Patch: SettingsPatch{"assetSelection": rawSetting(t, map[string]any{"autoAssetIds": []string{characterID}, "addedAssetIds": []string{}, "excludedAssetIds": []string{characterID}})}, ExpectedRevision: video.Revision}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	selection, ok := decodeVideoAssetSelection(latest.Books[0].Videos[0].SettingsState.Patch)
	if !ok || !slices.Contains(selection.ExcludedAssetIDs, characterID) || selection.effectiveAssetIDSet()[characterID] {
		t.Fatalf("manual asset exclusion was not carried to the new storyboard: %+v", selection)
	}
}

func TestDirectorUsesSavedWorkingFrontWithoutChangingCapturedSource(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "batch", Books: []CreateBookInput{{Title: "book", SourceText: "captured source"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	if _, err := store.SaveDraft(context.Background(), "alice", Draft{Key: "working-front:" + book.ID, Kind: "working-front-content", Scope: batch.ID, Content: "edited front"}); err != nil {
		t.Fatal(err)
	}
	service := &DirectorService{Store: store, Provider: &queuedDirectorProvider{values: []string{validDirectorJSON()}}}
	if _, err := service.RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	read, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if read.Books[0].SourceText != "captured source" {
		t.Fatalf("captured source mutated: %q", read.Books[0].SourceText)
	}
	if got := read.Books[0].DirectorRevision.SourceDigest; got != sourceDigest("edited front") {
		t.Fatalf("director digest=%q", got)
	}
}

func TestDirectorUsesConfiguredProductionLinesWithoutChangingCapturedSource(t *testing.T) {
	store := NewMemoryStore()
	source := "一\n\n二\n三\n四\n五\n六"
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "batch", Books: []CreateBookInput{{Title: "book", SourceText: source, SourceMetadata: map[string]any{"contentRangeLines": 5}}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	service := &DirectorService{Store: store, Provider: &queuedDirectorProvider{values: []string{validDirectorJSON()}}}
	if _, err := service.RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	read, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if read.Books[0].SourceText != source {
		t.Fatalf("captured source mutated: %q", read.Books[0].SourceText)
	}
	if got := read.Books[0].DirectorRevision.SourceDigest; got != sourceDigest("一\n二\n三\n四\n五") {
		t.Fatalf("director digest=%q", got)
	}
}

func TestRewriteWorkingFrontPersistsCandidateWithoutChangingCapturedSource(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "batch", Books: []CreateBookInput{{Title: "book", SourceText: "captured source"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	if _, err := store.SaveDraft(context.Background(), "alice", Draft{Key: "working-front:" + book.ID, Kind: "working-front-content", Scope: batch.ID, Content: "editable front"}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{"viral candidate"}}
	service := &DirectorService{Store: store, Provider: provider}
	candidate, err := service.RewriteWorkingFront(context.Background(), "alice", batch.ID, book.ID, "", PresetSnapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if candidate != "viral candidate" {
		t.Fatalf("candidate=%q", candidate)
	}
	if len(provider.calls) != 1 || !strings.Contains(provider.calls[0].UserPrompt, "editable front") {
		t.Fatalf("rewrite must use working front: %+v", provider.calls)
	}
	draft, err := store.GetDraft(context.Background(), "alice", "working-front-candidate:"+book.ID, "working-front-viral-candidate", batch.ID)
	if err != nil || draft.Content != "viral candidate" {
		t.Fatalf("candidate draft=%+v err=%v", draft, err)
	}
	loaded, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil || loaded.Books[0].SourceText != "captured source" {
		t.Fatalf("captured source must remain immutable: book=%+v err=%v", loaded.Books[0], err)
	}
	working, err := store.GetDraft(context.Background(), "alice", "working-front:"+book.ID, "working-front-content", batch.ID)
	if err != nil || working.Content != "editable front" {
		t.Fatalf("working front mutated before replacement: %+v err=%v", working, err)
	}
}

func TestDirectorAppliesVideoRulesAndDefersVisualPromptToVisualExtraction(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	config := map[string]any{
		"assets": map[string]any{
			"enabled":    true,
			"scope":      "all",
			"extraction": map[string]any{"body": "资产统一为国风写实，人物服装必须连续。"},
		},
		"constraints": map[string]any{"enabled": true, "selections": []any{map[string]any{"body": "镜头不得跳轴，不要文字和水印。"}}},
		"video":       map[string]any{"enabled": true, "body": "视频动作必须连续，运镜克制。", "scope": "all"},
		"visual":      map[string]any{"enabled": true, "body": "画面采用冷色电影光，主体清晰。", "scope": "all"},
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, config)}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	provider := &queuedDirectorProvider{values: []string{validDirectorJSON()}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 1 {
		t.Fatalf("director calls=%d", len(provider.calls))
	}
	for _, expected := range []string{"资产统一为国风写实", "镜头不得跳轴", "视频动作必须连续"} {
		if !strings.Contains(provider.calls[0].SystemPrompt, expected) {
			t.Fatalf("director contract missing %q:\n%s", expected, provider.calls[0].SystemPrompt)
		}
	}
	for _, forbidden := range []string{"画面采用冷色电影光", "visual_prompt"} {
		if strings.Contains(provider.calls[0].SystemPrompt, forbidden) {
			t.Fatalf("director contract must defer visual extraction, got %q:\n%s", forbidden, provider.calls[0].SystemPrompt)
		}
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	video := latest.Books[0].Videos[0]
	if video.VisualPrompt != "" {
		t.Fatalf("visual prompt=%q", video.VisualPrompt)
	}
	final, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", latest.ID, latest.Books[0].ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(final.CompiledPrompt, "VISUAL") || strings.Contains(final.CompiledPrompt, "画面采用冷色电影光") {
		t.Fatalf("visual meta prompt leaked into final video prompt:\n%s", final.CompiledPrompt)
	}
}

func TestBuildDirectorContractUsesUnifiedAssetRuleAndIgnoresArchivedTypedRules(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "原文"}
	config := map[string]any{
		"assets": map[string]any{
			"enabled":    true,
			"extraction": map[string]any{"body": "SCRIPT EXTRACTION"},
			"character":  map[string]any{"body": "CHARACTER ONLY"},
			"scene":      map[string]any{"body": "SCENE ONLY"},
			"prop":       map[string]any{"body": "PROP ONLY"},
		},
		"constraints": map[string]any{"enabled": true, "selections": []any{map[string]any{"body": "CONSTRAINT ONLY"}}},
		"video":       map[string]any{"enabled": true, "body": "VIDEO ONLY"},
		"visual":      map[string]any{"enabled": true, "body": "VISUAL ONLY"},
	}
	contract, err := BuildDirectorContract(book, HookRevision{}, DirectorSnapshot{
		Mode: "original", MaxVideoDuration: 15, AspectRatio: "9:16",
		Effective: SettingsPatch{"aiPromptConfig": rawSetting(t, config)},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"SCRIPT EXTRACTION", "CONSTRAINT ONLY", "VIDEO ONLY"} {
		if !strings.Contains(contract.SystemPrompt, expected) {
			t.Fatalf("director contract missing %q:\n%s", expected, contract.SystemPrompt)
		}
	}
	for _, archived := range []string{"CHARACTER ONLY", "SCENE ONLY", "PROP ONLY"} {
		if strings.Contains(contract.SystemPrompt, archived) {
			t.Fatalf("archived typed asset rule leaked %q:\n%s", archived, contract.SystemPrompt)
		}
	}
	if strings.Contains(contract.SystemPrompt, "VISUAL ONLY") {
		t.Fatalf("visual meta must be deferred to visual extraction:\n%s", contract.SystemPrompt)
	}
}

func TestBuildDirectorContractDescribesOneCombinedAssetPrompt(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "原文"}
	config := map[string]any{
		"assets": map[string]any{
			"enabled":    true,
			"extraction": map[string]any{"body": "COMBINED ASSET RULE"},
		},
	}
	contract, err := BuildDirectorContract(book, HookRevision{}, DirectorSnapshot{
		Mode: "original", MaxVideoDuration: 15, AspectRatio: "9:16",
		Effective: SettingsPatch{"aiPromptConfig": rawSetting(t, config)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contract.SystemPrompt, "人物/场景/道具统一提取方案") || !strings.Contains(contract.SystemPrompt, "COMBINED ASSET RULE") {
		t.Fatalf("combined asset rule is missing from director contract:\n%s", contract.SystemPrompt)
	}
}

func TestBuildDirectorContractAddsBaseSetupRuleForEveryStoryboard(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "原文"}
	config := map[string]any{
		"constraints": map[string]any{
			"enabled":   true,
			"baseSetup": map[string]any{"enabled": true},
		},
	}
	contract, err := BuildDirectorContract(book, HookRevision{}, DirectorSnapshot{
		Mode: "original", MaxVideoDuration: 15, AspectRatio: "9:16",
		Effective: SettingsPatch{"aiPromptConfig": rawSetting(t, config)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contract.SystemPrompt, "基础设定：每个 storyboard 必须带入当前情节出现的人物与场景") {
		t.Fatalf("base setup constraint is missing from director contract:\n%s", contract.SystemPrompt)
	}
}

func TestV11ContractsUseFrozenHookAndDerivedOpeningPresetsOnlyInTheirOwnFlows(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "原文"}
	config := map[string]any{
		"hook":             map[string]any{"enabled": true, "body": "HOOK PRESET"},
		"originalDirector": map[string]any{"enabled": true, "body": "ORIGINAL DIRECTOR PRESET"},
		"constraints":      map[string]any{"enabled": true, "wrapper": map[string]any{"body": "CONSTRAINT WRAPPER"}},
	}
	snapshot := DirectorSnapshot{Mode: "original", MaxVideoDuration: 10, AspectRatio: "9:16", Effective: SettingsPatch{"aiPromptConfig": rawSetting(t, config)}}
	hook := BuildHookContract(book, snapshot)
	if !strings.Contains(hook.SystemPrompt, "HOOK PRESET") {
		t.Fatalf("hook preset missing from contract: %s", hook.SystemPrompt)
	}
	workingFront := BuildWorkingFrontRewriteContract(book, PresetSnapshot{Body: "HOOK PRESET"})
	if !strings.Contains(workingFront.SystemPrompt, "HOOK PRESET") {
		t.Fatalf("working-front rewrite preset missing from contract: %s", workingFront.SystemPrompt)
	}
	director, err := BuildDirectorContract(book, HookRevision{}, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(director.SystemPrompt, "ORIGINAL DIRECTOR PRESET") {
		t.Fatalf("derived-opening presets must not enter the ordinary storyboard contract: %s", director.SystemPrompt)
	}
	if strings.Contains(director.SystemPrompt, "CONSTRAINT WRAPPER") {
		t.Fatalf("retired constraint wrapper must not enter the director contract: %s", director.SystemPrompt)
	}
	if len(director.Normalization.AllowedPrefixKeys) != len(DirectorPrefixKeys) {
		t.Fatalf("unexpected prefix keys: %#v", director.Normalization.AllowedPrefixKeys)
	}
}

func TestBuildDirectorContractIgnoresRetiredVideoStylePrefix(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试书", SourceText: "原文"}
	config := map[string]any{
		"prefix": map[string]any{"enabled": true, "body": "RETIRED VIDEO STYLE PREFIX", "presetKey": "modern_conflict"},
	}
	contract, err := BuildDirectorContract(book, HookRevision{}, DirectorSnapshot{
		Mode: "original", MaxVideoDuration: 10, AspectRatio: "9:16",
		Effective: SettingsPatch{"aiPromptConfig": rawSetting(t, config)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(contract.SystemPrompt, "RETIRED VIDEO STYLE PREFIX") {
		t.Fatalf("retired video style prefix leaked into director contract: %s", contract.SystemPrompt)
	}
	if len(contract.Normalization.AllowedPrefixKeys) != len(DirectorPrefixKeys) {
		t.Fatalf("retired prefix selection constrained director keys: %#v", contract.Normalization.AllowedPrefixKeys)
	}
}

func TestDirectorPersistsSmartUnifiedAnalysisInsteadOfPrefixPresetBody(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	config := map[string]any{
		"constraints": map[string]any{
			"enabled": true,
			"selections": []any{map[string]any{
				"presetId":           "script-constraint-prefix-smart-unified",
				"constraintCategory": "prefix",
				"body":               "【智能统一画面前缀】这是元提示词正文，不能进入最终视频提示词。",
			}},
		},
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, config)}, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{validDirectorJSON()}}
	style := "影像媒介：真人数字电影短剧；整体氛围：克制悬疑。"
	revision, err := (&DirectorService{Store: store, Provider: provider}).RunDirectorWithSmartUnifiedStyle(context.Background(), "alice", batch.ID, book.ID, style)
	if err != nil {
		t.Fatal(err)
	}
	if revision.Output.SmartUnifiedStyle != style {
		t.Fatalf("smart style=%q", revision.Output.SmartUnifiedStyle)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, latest.Books[0].ID, latest.Books[0].Videos[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt.CompiledPrompt, style) || strings.Contains(prompt.CompiledPrompt, "这是元提示词正文") {
		t.Fatalf("smart unified prefix must use analysis result, got:\n%s", prompt.CompiledPrompt)
	}
}

func TestVisualPromptExtractionUpdatesOnlyExistingStoryboardVisualPrompts(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	config := map[string]any{
		"visual": map[string]any{"enabled": true, "scope": "all", "body": "VISUAL PRESET: 只输出首帧画面"},
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, config)}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	book = batch.Books[0]
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: book.Videos[0].ID}, SettingsUpdate{Patch: SettingsPatch{"videoPrompt": rawSetting(t, "人物站在窗边，停顿后转身。")}, ExpectedRevision: book.Videos[0].Revision}); err != nil {
		t.Fatal(err)
	}
	batch, err = store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	book = batch.Books[0]
	originalVideoPrompt := book.Videos[0].VideoPrompt
	provider := &queuedDirectorProvider{values: []string{"冷色室内，人物站在窗边，电影级构图。"}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunVisualPromptExtraction(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	video := latest.Books[0].Videos[0]
	if video.VisualPrompt != "冷色室内，人物站在窗边，电影级构图。" {
		t.Fatalf("visual prompt=%q", video.VisualPrompt)
	}
	if video.VideoPrompt != originalVideoPrompt {
		t.Fatalf("visual extraction must not rewrite video prompt: got %q want %q", video.VideoPrompt, originalVideoPrompt)
	}
	if len(provider.calls) != 1 || !strings.Contains(provider.calls[0].SystemPrompt, "VISUAL PRESET") {
		t.Fatalf("visual extraction call=%+v", provider.calls)
	}
}

func TestAssetRefreshReplacesOnlyGeneratedAssetsAndKeepsManualAssets(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "batch", Books: []CreateBookInput{{Title: "book", SourceText: "第一行"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	if _, err := store.CreateBookAsset(context.Background(), "alice", batch.ID, book.ID, CreateBookAssetInput{Kind: "prop", Name: "用户保留道具", Prompt: "人工设定"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PersistExtractedBookAssets(context.Background(), "alice", book, DirectorSnapshot{}, DirectorAssets{Scenes: []NamedPrompt{{Name: "旧自动场景", Prompt: "旧提示词"}}, Props: []NamedPrompt{{Name: "旧自动道具", Prompt: "旧提示词"}}}); err != nil {
		t.Fatal(err)
	}
	assets, err := store.PersistExtractedBookAssets(context.Background(), "alice", book, DirectorSnapshot{}, DirectorAssets{Scenes: []NamedPrompt{{Name: "当前五行场景", Prompt: "新提示词"}}})
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, asset := range assets {
		seen[asset.Name] = true
	}
	if seen["旧自动场景"] || seen["旧自动道具"] || !seen["当前五行场景"] || !seen["用户保留道具"] {
		t.Fatalf("asset refresh must replace generated assets but preserve manual assets: %#v", assets)
	}
}

func TestOriginalAudioSourceKeepsAudioPlanningMeasurement(t *testing.T) {
	_, batch, book := seedDirectorBook(t, "original", false)
	batch.SettingsState.Patch["audioSource"] = rawSetting(t, "original")
	batch.SettingsState.Patch["audioPlanningEnabled"] = rawSetting(t, true)
	batch.SettingsState.Patch["audioDurationSeconds"] = rawSetting(t, 42)
	snapshot, err := snapshotForBook(batch, book)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.AudioDurationSeconds != 42 || snapshot.AudioTargetSeconds != 42 {
		t.Fatalf("original audio must keep TTS duration planning: %#v", snapshot)
	}
}
