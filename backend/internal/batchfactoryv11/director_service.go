package batchfactoryv11

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
)

type DirectorService struct {
	Store    Store
	Provider DirectorProvider
}

// retryableModelOutputError keeps the production contract strict while making
// an occasional malformed model reply recoverable. Configuration and input
// validation continue to use ErrInvalid so they stop immediately.
func retryableModelOutputError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: 文本模型输出不符合当前格式要求，可重试：%v", ErrUnavailable, err)
}

func sourceDigest(source string) string {
	sum := sha256.Sum256([]byte(source))
	return hex.EncodeToString(sum[:])
}

func rawString(patch SettingsPatch, key, fallback string) string {
	raw, ok := patch[key]
	if !ok {
		return fallback
	}
	var value string
	if json.Unmarshal(raw, &value) != nil || strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func rawBool(patch SettingsPatch, key string, fallback bool) bool {
	raw, ok := patch[key]
	if !ok {
		return fallback
	}
	var value bool
	if json.Unmarshal(raw, &value) != nil {
		return fallback
	}
	return value
}

func rawInt(patch SettingsPatch, key string, fallback int) int {
	raw, ok := patch[key]
	if !ok {
		return fallback
	}
	var value int
	if json.Unmarshal(raw, &value) != nil {
		return fallback
	}
	return value
}

// storyboard durations are integer seconds. Keep the measured audio value in
// settings, but plan against the same ceiling rule used by script generation.
func measuredAudioDurationSeconds(patch SettingsPatch) float64 {
	// Full TTS and line timing are part of every new production flow. Keep
	// explicit false values for historical snapshots, but default missing
	// configuration to enabled so audio source only controls the final mix.
	if !rawBool(patch, "audioPlanningEnabled", true) {
		return 0
	}
	raw, ok := patch["audioDurationSeconds"]
	if !ok {
		return 0
	}
	var seconds float64
	if json.Unmarshal(raw, &seconds) != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds <= 0 {
		return 0
	}
	return seconds
}

func audioTargetSeconds(patch SettingsPatch) int {
	seconds := measuredAudioDurationSeconds(patch)
	if seconds <= 0 {
		return 0
	}
	return int(math.Ceil(seconds))
}

func audioMinimumVideoCount(targetSeconds, maxVideoDuration int) int {
	if targetSeconds <= 0 || maxVideoDuration <= 0 {
		return 0
	}
	return int(math.Ceil(float64(targetSeconds) / float64(maxVideoDuration)))
}

func bookFromBatch(batch Batch, bookID string) (Book, error) {
	for _, book := range batch.Books {
		if book.ID == bookID {
			return book, nil
		}
	}
	return Book{}, ErrNotFound
}

func snapshotForBook(batch Batch, book Book) (DirectorSnapshot, error) {
	return snapshotForBookWithAudioRequirement(batch, book, true)
}

// Asset extraction is independent of timeline planning. It can run before a
// voice measurement supplies the duration later required by director output.
func snapshotForAssetExtraction(batch Batch, book Book) (DirectorSnapshot, error) {
	return snapshotForBookWithAudioRequirement(batch, book, false)
}

func snapshotForBookWithAudioRequirement(batch Batch, book Book, requireMeasuredAudio bool) (DirectorSnapshot, error) {
	effective := ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	mode := rawString(effective, "productionMode", rawString(effective, "mode", "original"))
	if mode == "original_direct" {
		mode = "original"
	}
	if mode == "viral_hook" {
		mode = "viral"
	}
	maxDuration := rawInt(effective, "storyboardDurationLimit", 10)
	if maxDuration != 10 && maxDuration != 15 {
		return DirectorSnapshot{}, fmt.Errorf("%w: storyboardDurationLimit must be 10 or 15", ErrInvalid)
	}
	fixed := rawBool(effective, "fixedSingleVideo", false)
	// 新链路始终先做完整配音和逐行时长测量；audioPlanningEnabled 只作为
	// 历史任务的兼容开关。新 SD 原文时间轴任务缺省即开启，旧任务仍尊重
	// 明确保存的 false，避免历史快照被静默升级。
	audioPlanningDefault := rawString(effective, "storyTimelineVersion", "") == sdOriginalStoryVersion
	audioPlanning := rawBool(effective, "audioPlanningEnabled", audioPlanningDefault)
	if requireMeasuredAudio && fixed && audioPlanning {
		return DirectorSnapshot{}, fmt.Errorf("%w: 固定开头只生产 VIDEO01，不能同时启用分镜规划跟随配音", ErrConflict)
	}
	if requireMeasuredAudio && audioPlanning && audioTargetSeconds(effective) == 0 {
		return DirectorSnapshot{}, fmt.Errorf("%w: 音频规划已开启，请先生成配音并读取真实时长", ErrConflict)
	}
	aspect := EffectiveVideoAspectRatio(effective)
	if mode != "original" && mode != "viral" {
		return DirectorSnapshot{}, fmt.Errorf("%w: unsupported director mode", ErrInvalid)
	}
	return DirectorSnapshot{Effective: effective, Mode: mode, MaxVideoDuration: maxDuration, AudioDurationSeconds: measuredAudioDurationSeconds(effective), AudioTargetSeconds: audioTargetSeconds(effective), FixedSingleVideo: fixed, AspectRatio: aspect}, nil
}

func (s *DirectorService) validate() error {
	if s == nil || s.Store == nil || s.Provider == nil {
		return ErrUnavailable
	}
	return nil
}

// productionBook returns a copy that keeps sourceText immutable while allowing
// a user-approved per-book working front to drive Hook and Director generation.
// The draft is deliberately separate from captured source text: publishing can
// later choose either source or the approved rewrite without destructive edits.
func (s *DirectorService) productionBook(ctx context.Context, owner, batchID string, book Book) (Book, error) {
	draft, err := s.Store.GetDraft(ctx, owner, "working-front:"+book.ID, "working-front-content", batchID)
	if err == nil && strings.TrimSpace(draft.Content) != "" {
		book.SourceText = strings.TrimSpace(draft.Content)
		return book, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return Book{}, err
	}
	book.SourceText = productionTextForBook(book)
	return book, nil
}

func productionTextForBook(book Book) string {
	limit := 5
	if raw, ok := book.SourceMetadata["contentRangeLines"]; ok {
		switch value := raw.(type) {
		case float64:
			limit = int(value)
		case int:
			limit = value
		case int64:
			limit = int(value)
		case json.Number:
			if parsed, err := value.Int64(); err == nil {
				limit = int(parsed)
			}
		}
	}
	if limit < 1 {
		limit = 5
	} else if limit > 500 {
		limit = 500
	}
	lines := h3NonEmptyVideoSourceLines(book.SourceText)
	if len(lines) > limit {
		lines = lines[:limit]
	}
	return strings.Join(lines, "\n")
}

// RewriteWorkingFront generates a candidate only. The approved working front
// remains unchanged until the user explicitly saves it from the workbench.
func (s *DirectorService) RewriteWorkingFront(ctx context.Context, owner, batchID, bookID, currentText string, opening PresetSnapshot) (string, error) {
	if err := s.validate(); err != nil {
		return "", err
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return "", err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(currentText) != "" {
		book.SourceText = strings.TrimSpace(currentText)
	} else {
		book, err = s.productionBook(ctx, owner, batchID, book)
		if err != nil {
			return "", err
		}
	}
	if strings.TrimSpace(book.SourceText) == "" {
		return "", fmt.Errorf("%w: working front text is required", ErrInvalid)
	}
	contract := BuildWorkingFrontRewriteContract(book, opening)
	candidate, err := s.Provider.Complete(ctx, TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens})
	if err != nil {
		return "", err
	}
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return "", fmt.Errorf("%w: empty working-front rewrite", ErrInvalid)
	}
	if _, err := s.Store.SaveDraft(ctx, owner, Draft{Key: "working-front-candidate:" + book.ID, Kind: "working-front-viral-candidate", Scope: batchID, Content: candidate}); err != nil {
		return "", err
	}
	return candidate, nil
}

func (s *DirectorService) RunHook(ctx context.Context, owner, batchID, bookID string) (HookRevision, error) {
	if err := s.validate(); err != nil {
		return HookRevision{}, err
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return HookRevision{}, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return HookRevision{}, err
	}
	book, err = s.productionBook(ctx, owner, batchID, book)
	if err != nil {
		return HookRevision{}, err
	}
	snapshot, err := snapshotForBook(batch, book)
	if err != nil {
		return HookRevision{}, err
	}
	if snapshot.Mode != "viral" {
		return HookRevision{}, fmt.Errorf("%w: Hook is only available in viral mode", ErrConflict)
	}
	if strings.TrimSpace(book.SourceText) == "" {
		return HookRevision{}, fmt.Errorf("%w: source text is required", ErrInvalid)
	}
	contract := BuildHookContract(book, snapshot)
	text, err := s.Provider.Complete(ctx, TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens})
	if err != nil {
		return HookRevision{}, err
	}
	return s.Store.CreateHookRevision(ctx, owner, batchID, bookID, strings.TrimSpace(text), sourceDigest(book.SourceText))
}

func (s *DirectorService) ApproveHook(ctx context.Context, owner, batchID, bookID, hookID string) (HookRevision, error) {
	if err := s.validate(); err != nil {
		return HookRevision{}, err
	}
	return s.Store.ApproveHookRevision(ctx, owner, batchID, bookID, hookID)
}

func (s *DirectorService) RunDirector(ctx context.Context, owner, batchID, bookID string) (DirectorRevision, error) {
	return s.runDirector(ctx, owner, batchID, bookID, nil)
}

// RunDirectorWithSmartUnifiedStyle accepts only a bridge-derived, validated
// full-book visual analysis. The user controls the selected preset; the model
// result is kept with this director revision rather than mutable settings.
func (s *DirectorService) RunDirectorWithSmartUnifiedStyle(ctx context.Context, owner, batchID, bookID, style string) (DirectorRevision, error) {
	analysis, err := parseSmartUnifiedAnalysis(style)
	if err != nil {
		return DirectorRevision{}, err
	}
	return s.runDirector(ctx, owner, batchID, bookID, analysis)
}

func (s *DirectorService) runDirector(ctx context.Context, owner, batchID, bookID string, smartUnifiedAnalysis *SmartUnifiedAnalysis) (DirectorRevision, error) {
	if err := s.validate(); err != nil {
		return DirectorRevision{}, err
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return DirectorRevision{}, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return DirectorRevision{}, err
	}
	book, err = s.productionBook(ctx, owner, batchID, book)
	if err != nil {
		return DirectorRevision{}, err
	}
	if strings.TrimSpace(book.SourceText) == "" {
		return DirectorRevision{}, fmt.Errorf("%w: source text is required", ErrInvalid)
	}
	snapshot, err := snapshotForBook(batch, book)
	if err != nil {
		return DirectorRevision{}, err
	}
	if smartUnifiedAnalysis != nil {
		encoded, _ := json.Marshal(smartUnifiedAnalysis)
		snapshot.Effective["smartUnifiedStyle"] = encoded
	}
	hook := HookRevision{}
	if snapshot.Mode == "viral" {
		hook, err = s.Store.LatestHookRevision(ctx, owner, batchID, bookID)
		if err != nil {
			return DirectorRevision{}, fmt.Errorf("%w: viral mode requires approved Hook", ErrConflict)
		}
		if hook.Status != "approved" {
			return DirectorRevision{}, fmt.Errorf("%w: viral mode requires approved Hook", ErrConflict)
		}
	}
	// sd.original-story.v1 uses one SD video-prompt request. The response must
	// contain both the story-script line and the fixed director shots; keeping
	// the old SD_STORY_SCRIPT helper below preserves historical compatibility
	// without making a second model call for new tasks.
	contract, err := BuildDirectorContract(book, hook, snapshot)
	if err != nil {
		return DirectorRevision{}, err
	}
	completion, err := s.Provider.Complete(ctx, TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens, DisableJSONResponse: contract.SDTextProtocol})
	if err != nil {
		return DirectorRevision{}, err
	}
	var result DirectorResult
	if contract.SDTextProtocol {
		result, err = parseSDDirectorText(completion, contract.Normalization.MaxVideoDuration)
		if err != nil {
			return DirectorRevision{}, retryableModelOutputError(err)
		}
	} else {
		raw, err := ParseDirectorJSON(completion)
		if err != nil {
			// Some OpenAI-compatible gateways accept JSON mode but still return
			// explanatory text on the first attempt. Give the same frozen
			// contract one repair attempt without response_format before marking
			// the stage failed; this keeps a transient formatting issue retryable
			// without weakening the final validator.
			retryCompletion, retryErr := s.Provider.Complete(ctx, TextCompletionRequest{
				SystemPrompt:        contract.SystemPrompt + "\n上一轮输出无法解析。现在只输出一个完整、合法的 JSON 对象，不要解释，不要 Markdown。",
				UserPrompt:          contract.UserPrompt,
				Temperature:         0.2,
				MaxTokens:           contract.MaxTokens,
				DisableJSONResponse: true,
			})
			if retryErr == nil {
				raw, err = ParseDirectorJSON(retryCompletion)
			}
			if err != nil {
				return DirectorRevision{}, retryableModelOutputError(err)
			}
		}
		result, err = NormalizeDirectorOutput(raw, contract.Normalization)
		if err != nil {
			return DirectorRevision{}, retryableModelOutputError(err)
		}
	}
	if smartUnifiedAnalysis != nil {
		result.SmartUnifiedStyle = smartUnifiedAnalysis.Prompt
		if selectedSDVideoPreset(aiReasoningPromptConfig(snapshot.Effective).Video) && strings.TrimSpace(smartUnifiedAnalysis.SDUnifiedStyle) != "" {
			result.SmartUnifiedStyle = smartUnifiedAnalysis.SDUnifiedStyle
		}
		result.SmartUnifiedAnalysis = smartUnifiedAnalysis
	}
	return s.Store.PersistDirectorRevision(ctx, owner, book, snapshot, sourceDigest(book.SourceText), hook.ID, result)
}

// RunAssetExtraction runs the compact asset-only request. Existing director
// revisions, storyboard prompts and VIDEO records are intentionally untouched.
func (s *DirectorService) RunAssetExtraction(ctx context.Context, owner, batchID, bookID string) ([]BookAsset, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return nil, err
	}
	book, err = s.productionBook(ctx, owner, batchID, book)
	if err != nil {
		return nil, err
	}
	snapshot, err := snapshotForAssetExtraction(batch, book)
	if err != nil {
		return nil, err
	}
	config := aiReasoningPromptConfig(snapshot.Effective)
	if smartUnifiedSelectedForRevision(book, config) && strings.TrimSpace(rawString(snapshot.Effective, "h3StyleAnalysis", "")) == "" {
		styleContract, contractErr := BuildSmartUnifiedContract(book, snapshot)
		if contractErr != nil {
			return nil, contractErr
		}
		styleCompletion, completionErr := s.Provider.Complete(ctx, TextCompletionRequest{SystemPrompt: styleContract.SystemPrompt, UserPrompt: styleContract.UserPrompt, Temperature: styleContract.Temperature, MaxTokens: styleContract.MaxTokens})
		if completionErr != nil {
			return nil, completionErr
		}
		styleRaw, parseErr := ParseDirectorJSON(styleCompletion)
		if parseErr != nil {
			return nil, retryableModelOutputError(parseErr)
		}
		analysis, analysisErr := SmartUnifiedAnalysisFromAssetExtractionOutput(styleRaw)
		if analysisErr != nil || analysis == nil {
			if analysisErr == nil {
				analysisErr = fmt.Errorf("智能统一未返回完整视觉基线")
			}
			return nil, retryableModelOutputError(analysisErr)
		}
		encoded, _ := json.Marshal(analysis)
		analysisSetting, _ := json.Marshal(string(encoded))
		latest, latestErr := s.Store.GetBatch(ctx, owner, batchID)
		if latestErr != nil {
			return nil, latestErr
		}
		for _, candidate := range latest.Books {
			if candidate.ID != bookID {
				continue
			}
			if _, latestErr = s.Store.SaveSettings(ctx, owner, ScopeRef{Kind: ScopeBook, BatchID: batchID, BookID: bookID}, SettingsUpdate{
				Patch: SettingsPatch{"h3StyleAnalysis": json.RawMessage(analysisSetting)}, ExpectedRevision: candidate.Revision,
			}); latestErr != nil {
				return nil, latestErr
			}
			break
		}
		snapshot.Effective["h3StyleAnalysis"] = analysisSetting
	}
	contract, err := BuildAssetExtractionContract(book, snapshot)
	if err != nil {
		return nil, err
	}
	completion, err := s.Provider.Complete(ctx, TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens})
	if err != nil {
		return nil, err
	}
	raw, err := ParseDirectorJSON(completion)
	if err != nil {
		retryCompletion, retryErr := s.Provider.Complete(ctx, TextCompletionRequest{
			SystemPrompt:        contract.SystemPrompt + "\n上一轮输出无法解析。现在只输出一个完整、合法的 JSON 对象，不要解释，不要 Markdown。",
			UserPrompt:          contract.UserPrompt,
			Temperature:         0.2,
			MaxTokens:           contract.MaxTokens,
			DisableJSONResponse: true,
		})
		if retryErr == nil {
			raw, err = ParseDirectorJSON(retryCompletion)
		}
		if err != nil {
			return nil, retryableModelOutputError(err)
		}
	}
	assets, err := NormalizeAssetExtractionOutput(raw)
	if err != nil {
		return nil, retryableModelOutputError(err)
	}
	// The selected full H3 assets preset already returns detailed appearances
	// together with characters/scenes/props. Other renderer selections retain
	// their historical post-extraction compilation behavior.
	config = aiReasoningPromptConfig(snapshot.Effective)
	if !(config.Assets.appliesTo(book, config.Assets.Extraction) && usesCombinedAssetRenderer(config.Assets.Extraction)) {
		assets, err = s.compileH3AssetPrompts(ctx, book, snapshot, assets)
		if err != nil {
			return nil, err
		}
	}
	persisted, err := s.Store.PersistExtractedBookAssets(ctx, owner, book, snapshot, assets)
	if err != nil {
		return nil, err
	}
	return persisted, nil
}

// RunOpeningVariants generates alternative VIDEO01 openings. It silently
// returns no variants when the feature is off, the book has fewer than two
// storyboards, or the director output is not SD-direct; callers treat an empty
// result as "no opening replacement" and continue the normal pipeline.
func (s *DirectorService) RunOpeningVariants(ctx context.Context, owner, batchID, bookID string, meta PresetSnapshot) ([]OpeningVariant, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return nil, err
	}
	effective := ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	if !rawBool(effective, "openingEnabled", false) {
		return nil, nil
	}
	if len(book.Videos) < 2 || book.DirectorRevision == nil || len(book.DirectorRevision.Output.Storyboard) < 2 {
		return nil, nil
	}
	storyboard := book.DirectorRevision.Output.Storyboard
	first, followUps := storyboard[0], storyboard[1:]
	// SD 直出分镜的整段提示词存于 FinalPrompt；结构化分镜只有镜头/动作描述，
	// 由 storyboardVideoPrompt 渲染成“镜头画面”文本。两种形态都可以生成换开头：
	// 变体本身是分镜一整段开场的重写，编译时走整段提示词路径，外层约束照常注入。
	// 只有连可渲染的分镜一正文都拿不到（异常空数据）时才跳过。
	if strings.TrimSpace(storyboardVideoPrompt(first)) == "" {
		return nil, nil
	}
	count := rawInt(effective, "openingCount", 4)
	if count < 1 || count > 8 {
		return nil, fmt.Errorf("%w: openingCount must be between 1 and 8", ErrInvalid)
	}
	variantCount := count - 1
	if variantCount < 1 {
		return nil, nil
	}
	snapshot, err := snapshotForBook(batch, book)
	if err != nil {
		return nil, err
	}
	contract := BuildOpeningVariantsContract(book, first, followUps, meta, variantCount, snapshot.MaxVideoDuration)
	text, err := s.Provider.Complete(ctx, TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens, DisableJSONResponse: true})
	if err != nil {
		return nil, err
	}
	variants := parseOpeningVariants(text, snapshot.MaxVideoDuration, first.DurationSec, variantCount)
	merged, err := s.saveOpeningVariants(ctx, owner, batchID, bookID, book.Videos[0].ID, variants)
	if err != nil {
		return nil, err
	}
	if err := requireSuccessfulOpeningVariants(merged, variantCount); err != nil {
		return merged, err
	}
	return merged, nil
}

func requireSuccessfulOpeningVariants(variants []OpeningVariant, variantCount int) error {
	byIndex := make(map[int]OpeningVariant, len(variants))
	for _, variant := range variants {
		byIndex[variant.Index] = variant
	}
	failed := make([]string, 0)
	for index := 1; index <= variantCount; index++ {
		variant, ok := byIndex[index]
		if ok && variant.Status == "success" && strings.TrimSpace(variant.Prompt) != "" {
			continue
		}
		reason := "模型未输出该变体分段"
		if ok && strings.TrimSpace(variant.FailureReason) != "" {
			reason = strings.TrimSpace(variant.FailureReason)
		}
		failed = append(failed, fmt.Sprintf("换开头%d：%s", index, reason))
	}
	if len(failed) == 0 {
		return nil
	}
	return retryableModelOutputError(fmt.Errorf("换开头变体未全部生成；%s", strings.Join(failed, "；")))
}

// saveOpeningVariants writes the generated variants into VIDEO01's settings
// patch under openingVariants. A failed slot never overwrites a previously
// stored success at the same index, so a bad rerun cannot destroy usable
// variants; every other patch key is left untouched by the sparse update.
func (s *DirectorService) saveOpeningVariants(ctx context.Context, owner, batchID, bookID, videoID string, variants []OpeningVariant) ([]OpeningVariant, error) {
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return nil, err
	}
	var video *Video
	for i := range book.Videos {
		if book.Videos[i].ID == videoID {
			video = &book.Videos[i]
			break
		}
	}
	if video == nil {
		return nil, ErrNotFound
	}
	merged := append([]OpeningVariant(nil), variants...)
	if raw, ok := video.SettingsState.Patch["openingVariants"]; ok {
		var previous []OpeningVariant
		if json.Unmarshal(raw, &previous) == nil {
			for i, variant := range merged {
				if variant.Status != "failed" {
					continue
				}
				for _, old := range previous {
					if old.Index == variant.Index && old.Status == "success" {
						merged[i] = old
						break
					}
				}
			}
		}
	}
	encoded, err := json.Marshal(merged)
	if err != nil {
		return nil, err
	}
	_, err = s.Store.SaveSettings(ctx, owner, ScopeRef{Kind: ScopeVideo, BatchID: batchID, BookID: bookID, VideoID: videoID}, SettingsUpdate{
		Patch:            SettingsPatch{"openingVariants": encoded},
		ExpectedRevision: video.Revision,
	})
	if err != nil {
		return nil, err
	}
	return merged, nil
}
