package batchfactoryv11

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"qiantie/backend/internal/prompt"
)

type DirectorService struct {
	Store         Store
	Provider      DirectorProvider
	PresetCatalog *prompt.Catalog
}

// retryableModelOutputError keeps the production contract strict while making
// an occasional malformed model reply recoverable. Configuration and input
// validation continue to use ErrInvalid so they stop immediately.
func retryableModelOutputError(err error) error {
	if err == nil {
		return nil
	}
	// 解析器报错（如 JSON 语法错误）是英文原文，不能顶在主文案里；业务原因
	// 已经是中文，仍然直接跟在冒号后面展示。
	if hasCJK(err.Error()) {
		return fmt.Errorf("%w: 文本模型输出不符合当前格式要求，可重试：%v", ErrUnavailable, err)
	}
	return fmt.Errorf("%w: 文本模型输出不符合当前格式要求，可重试\n上游原文：%v", ErrUnavailable, err)
}

func hasCJK(text string) bool {
	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

func isTruncatedDirectorOutput(err error) bool {
	return IsModelOutputTruncated(err)
}

func sourceDigest(source string) string {
	sum := sha256.Sum256([]byte(source))
	return hex.EncodeToString(sum[:])
}

// SourceInputDigest 暴露导演输入摘要，让自动化在触发前就能判断
// “现有分镜是否由当前正文生成”。摘要一致时分镜阶段可以直接跳过，
// 不再把已完成的分镜重新生成一遍。
func SourceInputDigest(source string) string {
	return sourceDigest(source)
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

// audioPlanningRequired keeps the complete timing chain independent from the
// final audio mix choice. New SD timeline tasks cannot opt out; a historical
// snapshot with an explicit false remains on the legacy semantic timeline.
func audioPlanningRequired(patch SettingsPatch) bool {
	if rawString(patch, "storyTimelineVersion", "") == sdOriginalStoryVersion {
		return true
	}
	if source := rawString(patch, "audioSource", ""); source == "original" || source == "tts" {
		return true
	}
	if rawPositiveNumber(patch, "audioDurationSeconds") {
		return true
	}
	if _, exists := patch["audioPlanningEnabled"]; !exists {
		return false
	}
	return rawBool(patch, "audioPlanningEnabled", true)
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

func rawPositiveNumber(patch SettingsPatch, key string) bool {
	raw, ok := patch[key]
	if !ok {
		return false
	}
	var value float64
	if json.Unmarshal(raw, &value) != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return false
	}
	return value > 0
}

// storyboard durations are integer seconds. Keep the measured audio value in
// settings, but plan against the same ceiling rule used by script generation.
func measuredAudioDurationSeconds(patch SettingsPatch) float64 {
	// Full TTS and line timing are part of every marked production flow. Keep
	// unmarked historical snapshots compatible, while a saved audio source
	// makes the new flow explicit and leaves audio source only to final mixing.
	if !audioPlanningRequired(patch) {
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
	// productionMode now selects the output medium (video/comic); `mode` is the
	// director's original/viral storytelling choice. Preserve old books whose
	// productionMode held the director mode before this split was introduced.
	mode := rawString(effective, "mode", "original")
	if legacyMode := rawString(effective, "productionMode", ""); legacyMode == "original" || legacyMode == "viral" || legacyMode == "original_direct" || legacyMode == "viral_hook" {
		mode = legacyMode
	}
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
	audioPlanning := audioPlanningRequired(effective)
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

// hydrateVideoPromptPreset keeps saved settings aligned with the published
// server-owned video preset. New saves already freeze the body in the HTTP
// layer; this runtime fallback also refreshes older settings when their saved
// preset version is behind the currently published version.
func (s *DirectorService) hydrateVideoPromptPreset(snapshot DirectorSnapshot) (DirectorSnapshot, error) {
	if s == nil || s.PresetCatalog == nil {
		return snapshot, nil
	}
	raw, ok := snapshot.Effective["aiPromptConfig"]
	if !ok || len(raw) == 0 {
		return snapshot, nil
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil || config == nil {
		return snapshot, fmt.Errorf("%w: AI 提示词配置格式不正确", ErrInvalid)
	}
	video, ok := config["video"].(map[string]any)
	if !ok || video == nil {
		return snapshot, nil
	}
	id, _ := video["presetId"].(string)
	id = strings.TrimSpace(id)
	if id == "" {
		return snapshot, nil
	}
	preset, found, err := s.PresetCatalog.PublishedBody("batch-factory", id)
	if err != nil {
		return snapshot, fmt.Errorf("%w: 读取视频提示词预设失败", ErrUnavailable)
	}
	if !found || strings.TrimSpace(preset.Body) == "" {
		return snapshot, fmt.Errorf("%w: 视频提示词预设不存在或未发布", ErrInvalid)
	}
	storedVersion, _ := video["presetVersion"].(float64)
	storedBody, _ := video["body"].(string)
	if strings.TrimSpace(storedBody) != "" && int(storedVersion) == preset.Version {
		return snapshot, nil
	}
	video["presetId"] = preset.ID
	video["presetName"] = preset.Name
	video["presetSlot"] = preset.Slot
	video["presetVersion"] = preset.Version
	video["body"] = preset.Body
	video["constraintCategory"] = preset.ConstraintCategory
	encoded, err := json.Marshal(config)
	if err != nil {
		return snapshot, fmt.Errorf("%w: 保存视频提示词配置失败", ErrInvalid)
	}
	snapshot.Effective["aiPromptConfig"] = encoded
	return snapshot, nil
}

// productionBook returns a copy that keeps sourceText immutable while allowing
// a user-approved per-book working front to drive Hook and Director generation.
// The draft is deliberately separate from captured source text: publishing can
// later choose either source or the approved rewrite without destructive edits.
func (s *DirectorService) productionBook(ctx context.Context, owner, batchID string, book Book) (Book, error) {
	return s.productionBookWithSourceMode(ctx, owner, batchID, book, false)
}

func (s *DirectorService) comicProductionBook(ctx context.Context, owner, batchID string, book Book) (Book, error) {
	book, err := s.productionBookWithSourceMode(ctx, owner, batchID, book, true)
	if err != nil {
		return Book{}, err
	}
	book.SourceText = comicSourceTextForBook(book)
	return book, nil
}

func comicSourceTextForBook(book Book) string {
	source := strings.TrimSpace(book.SourceText)
	limit := 2000
	if raw, ok := book.SourceMetadata["comicAnalysisCharacters"]; ok {
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
	if limit < 100 {
		limit = 100
	}
	if limit > 100000 {
		limit = 100000
	}
	runes := []rune(source)
	if len(runes) <= limit {
		return source
	}
	next := -1
	for index := limit; index < len(runes); index++ {
		if strings.ContainsRune("。！？!?；;\n", runes[index]) {
			next = index
			break
		}
	}
	previous := -1
	for index := limit - 1; index >= 0; index-- {
		if strings.ContainsRune("。！？!?；;\n", runes[index]) {
			previous = index
			break
		}
	}
	if previous >= 0 && (next < 0 || limit-previous <= next-limit) {
		return string(runes[:previous+1])
	}
	if next >= 0 {
		return string(runes[:next+1])
	}
	return string(runes[:limit])
}

func (s *DirectorService) productionBookWithSourceMode(ctx context.Context, owner, batchID string, book Book, completeSource bool) (Book, error) {
	draft, err := s.Store.GetDraft(ctx, owner, "working-front:"+book.ID, "working-front-content", batchID)
	if err == nil && strings.TrimSpace(draft.Content) != "" {
		book.SourceText = strings.TrimSpace(draft.Content)
		return book, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return Book{}, err
	}
	if completeSource {
		book.SourceText = strings.TrimSpace(book.SourceText)
		return book, nil
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
	snapshot, err := snapshotForBook(batch, book)
	if err != nil {
		return DirectorRevision{}, err
	}
	snapshot, err = s.hydrateVideoPromptPreset(snapshot)
	if err != nil {
		return DirectorRevision{}, err
	}
	if normalizeProductionMode(snapshot.Effective) == "comic" {
		book, err = s.comicProductionBook(ctx, owner, batchID, book)
	} else {
		book, err = s.productionBook(ctx, owner, batchID, book)
	}
	if err != nil {
		return DirectorRevision{}, err
	}
	if strings.TrimSpace(book.SourceText) == "" {
		return DirectorRevision{}, fmt.Errorf("%w: source text is required", ErrInvalid)
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
	result, err := s.completeDirectorContract(ctx, contract)
	if err != nil {
		return DirectorRevision{}, err
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

// CreateDirectorPromptCandidate regenerates prompts for the existing VIDEO
// rows without creating a director revision or replacing any accepted prompt.
func (s *DirectorService) CreateDirectorPromptCandidate(ctx context.Context, owner, batchID, bookID string, expectedBookRevision int64) (DirectorPromptCandidate, error) {
	if err := s.validate(); err != nil {
		return DirectorPromptCandidate{}, err
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return DirectorPromptCandidate{}, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return DirectorPromptCandidate{}, err
	}
	if book.Revision != expectedBookRevision {
		return DirectorPromptCandidate{}, ErrConflict
	}
	book, err = s.productionBook(ctx, owner, batchID, book)
	if err != nil {
		return DirectorPromptCandidate{}, err
	}
	if strings.TrimSpace(book.SourceText) == "" {
		return DirectorPromptCandidate{}, fmt.Errorf("%w: source text is required", ErrInvalid)
	}
	snapshot, err := snapshotForBook(batch, book)
	if err != nil {
		return DirectorPromptCandidate{}, err
	}
	snapshot, err = s.hydrateVideoPromptPreset(snapshot)
	if err != nil {
		return DirectorPromptCandidate{}, err
	}
	hook := HookRevision{}
	if snapshot.Mode == "viral" {
		hook, err = s.Store.LatestHookRevision(ctx, owner, batchID, bookID)
		if err != nil || hook.Status != "approved" {
			return DirectorPromptCandidate{}, fmt.Errorf("%w: viral mode requires approved Hook", ErrConflict)
		}
	}
	contract, err := BuildDirectorContract(book, hook, snapshot)
	if err != nil {
		return DirectorPromptCandidate{}, err
	}
	contract = directorPromptCandidateContract(contract, len(book.Videos), 0)
	result, err := s.completeDirectorContract(ctx, contract)
	if err != nil {
		return DirectorPromptCandidate{}, err
	}
	if len(result.Storyboard) != len(book.Videos) {
		// Candidate regeneration only rewrites prompts for existing VIDEO rows.
		// Never pad or truncate here: either operation could attach a prompt to
		// the wrong VIDEO. Give the model one bounded correction attempt with
		// the actual count and the observed mismatch made explicit.
		retryContract := directorPromptCandidateContract(contract, len(book.Videos), len(result.Storyboard))
		result, err = s.completeDirectorContract(ctx, retryContract)
		if err != nil {
			return DirectorPromptCandidate{}, err
		}
	}
	if len(result.Storyboard) != len(book.Videos) {
		return DirectorPromptCandidate{}, retryableModelOutputError(fmt.Errorf("候选重生成返回 %d 个 storyboard，当前已有 %d 个 VIDEO", len(result.Storyboard), len(book.Videos)))
	}
	latest, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return DirectorPromptCandidate{}, err
	}
	latestBook, err := bookFromBatch(latest, bookID)
	if err != nil {
		return DirectorPromptCandidate{}, err
	}
	if latestBook.Revision != expectedBookRevision {
		return DirectorPromptCandidate{}, ErrConflict
	}
	now := time.Now().UTC()
	candidate := DirectorPromptCandidate{
		ID:                 "director-prompt-candidate-" + sourceDigest(fmt.Sprintf("%s:%s:%d", batchID, bookID, now.UnixNano()))[:16],
		SourceBookRevision: latestBook.Revision,
		CreatedAt:          now,
		VideoPrompts:       make(map[string]string, len(latestBook.Videos)),
	}
	for index, video := range latestBook.Videos {
		candidate.VideoPrompts[video.ID] = storyboardVideoPrompt(result.Storyboard[index])
	}
	encoded, err := json.Marshal(candidate)
	if err != nil {
		return DirectorPromptCandidate{}, err
	}
	if _, err := s.Store.SaveSettings(ctx, owner, ScopeRef{Kind: ScopeBook, BatchID: batchID, BookID: bookID}, SettingsUpdate{
		Patch: SettingsPatch{"directorPromptCandidate": encoded}, ExpectedRevision: latestBook.Revision,
	}); err != nil {
		return DirectorPromptCandidate{}, err
	}
	return candidate, nil
}

// directorPromptCandidateContract narrows the normal director contract to a
// prompt-only regeneration for the already-created VIDEO rows. The existing
// rows are the source of truth for cardinality and ordering; this operation
// must never create, delete, or silently remap storyboard cards.
func directorPromptCandidateContract(contract PromptContract, existingVideoCount, previousStoryboardCount int) PromptContract {
	instruction := fmt.Sprintf(`
候选提示词重生成模式（最高优先级）：这不是重新分页，也不是重新创建 VIDEO。当前已经存在恰好 %d 个 VIDEO，必须只返回恰好 %d 个 storyboard，按现有 VIDEO 的顺序一一对应。
- 不得增加、删除、合并或拆分 storyboard；不得改变既有 VIDEO 的卡片边界。
- 每个 storyboard 只负责生成对应 VIDEO 的候选提示词，必须保留该段剧情承接、人物、场景、道具和时间轴。
- 返回数量不等于 %d 时，本次结果不可用。`, existingVideoCount, existingVideoCount, existingVideoCount)
	if previousStoryboardCount > 0 {
		instruction += fmt.Sprintf("\n上一轮返回了 %d 个 storyboard，数量不正确。请完整重做，并严格返回 %d 个 storyboard。", previousStoryboardCount, existingVideoCount)
	}
	contract.SystemPrompt += instruction

	var payload map[string]any
	if err := json.Unmarshal([]byte(contract.UserPrompt), &payload); err == nil {
		payload["candidate_regeneration"] = map[string]any{
			"existing_video_count":      existingVideoCount,
			"preserve_video_boundaries": true,
			"preserve_video_order":      true,
		}
		if encoded, err := json.MarshalIndent(payload, "", "  "); err == nil {
			contract.UserPrompt = string(encoded)
		}
	} else {
		contract.UserPrompt += fmt.Sprintf("\n候选重生成约束：现有 VIDEO 数量为 %d，必须返回恰好 %d 个 storyboard，按现有顺序一一对应。", existingVideoCount, existingVideoCount)
	}
	return contract
}

func (s *DirectorService) DirectorPromptCandidate(ctx context.Context, owner, batchID, bookID string) (DirectorPromptCandidate, error) {
	if s == nil || s.Store == nil {
		return DirectorPromptCandidate{}, ErrUnavailable
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return DirectorPromptCandidate{}, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return DirectorPromptCandidate{}, err
	}
	return directorPromptCandidateFromBook(book)
}

func (s *DirectorService) AcceptDirectorPromptCandidate(ctx context.Context, owner, batchID, bookID, videoID, candidateID string, expectedVideoRevision int64) (Video, error) {
	if s == nil || s.Store == nil {
		return Video{}, ErrUnavailable
	}
	if err := s.Store.AcceptDirectorPromptCandidate(ctx, owner, batchID, bookID, videoID, candidateID, expectedVideoRevision); err != nil {
		return Video{}, err
	}
	latest, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return Video{}, err
	}
	latestBook, err := bookFromBatch(latest, bookID)
	if err != nil {
		return Video{}, err
	}
	for _, video := range latestBook.Videos {
		if video.ID == videoID {
			return video, nil
		}
	}
	return Video{}, ErrNotFound
}

func (s *DirectorService) CancelDirectorPromptCandidate(ctx context.Context, owner, batchID, bookID string, expectedBookRevision int64) error {
	if s == nil || s.Store == nil {
		return ErrUnavailable
	}
	if _, err := s.DirectorPromptCandidate(ctx, owner, batchID, bookID); err != nil {
		return err
	}
	_, err := s.Store.SaveSettings(ctx, owner, ScopeRef{Kind: ScopeBook, BatchID: batchID, BookID: bookID}, SettingsUpdate{
		RestoreKeys: []string{"directorPromptCandidate"}, ExpectedRevision: expectedBookRevision,
	})
	return err
}

func directorPromptCandidateFromBook(book Book) (DirectorPromptCandidate, error) {
	return directorPromptCandidateFromPatch(book.SettingsState.Patch)
}

func directorPromptCandidateFromPatch(patch SettingsPatch) (DirectorPromptCandidate, error) {
	raw, exists := patch["directorPromptCandidate"]
	if !exists {
		return DirectorPromptCandidate{}, ErrNotFound
	}
	var candidate DirectorPromptCandidate
	if err := json.Unmarshal(raw, &candidate); err != nil || strings.TrimSpace(candidate.ID) == "" || len(candidate.VideoPrompts) == 0 {
		return DirectorPromptCandidate{}, fmt.Errorf("%w: malformed director prompt candidate", ErrInvalid)
	}
	return candidate, nil
}

func (s *DirectorService) completeDirectorContract(ctx context.Context, contract PromptContract) (DirectorResult, error) {
	completion, err := s.Provider.Complete(ctx, TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens, DisableJSONResponse: contract.SDTextProtocol})
	if err != nil {
		return DirectorResult{}, err
	}
	if contract.SDTextProtocol {
		result, err := parseSDDirectorText(completion, contract.Normalization.MaxVideoDuration)
		if err == nil {
			return result, nil
		}
		retryCompletion, retryErr := s.Provider.Complete(ctx, TextCompletionRequest{SystemPrompt: contract.SystemPrompt + "\n上一轮输出无法解析。现在仅输出完整有效的 SD 分镜文本，不要解释，不要 Markdown。", UserPrompt: contract.UserPrompt, Temperature: 0.2, MaxTokens: contract.MaxTokens, DisableJSONResponse: true})
		if retryErr == nil {
			result, err = parseSDDirectorText(retryCompletion, contract.Normalization.MaxVideoDuration)
		}
		if err != nil {
			return DirectorResult{}, retryableModelOutputError(err)
		}
		return result, nil
	}
	raw, err := ParseDirectorJSON(completion)
	if err != nil {
		retryCompletion, retryErr := s.Provider.Complete(ctx, TextCompletionRequest{SystemPrompt: contract.SystemPrompt + "\n上一轮输出无法解析。现在只输出一个完整、合法的 JSON 对象，不要解释，不要 Markdown。", UserPrompt: contract.UserPrompt, Temperature: 0.2, MaxTokens: contract.MaxTokens, DisableJSONResponse: true})
		if retryErr == nil {
			raw, err = ParseDirectorJSON(retryCompletion)
		}
		if err != nil {
			return DirectorResult{}, retryableModelOutputError(err)
		}
	}
	result, err := NormalizeDirectorOutput(raw, contract.Normalization)
	if err != nil {
		return DirectorResult{}, retryableModelOutputError(err)
	}
	return result, nil
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
	if err != nil && !isTruncatedDirectorOutput(err) {
		return nil, err
	}
	raw, parseErr := ParseDirectorJSON(completion)
	if err != nil || parseErr != nil {
		retryPrompt := "\n上一轮输出无法解析。现在只输出一个完整、合法的 JSON 对象，不要解释，不要 Markdown。"
		if isTruncatedDirectorOutput(err) {
			retryPrompt = "\n上一轮输出因长度截断。请保留必要的角色、场景和道具，使用更紧凑的短语并减少输出量；现在只输出一个完整、合法的 JSON 对象，不要解释，不要 Markdown。"
		}
		retryCompletion, retryErr := s.Provider.Complete(ctx, TextCompletionRequest{
			SystemPrompt:        contract.SystemPrompt + retryPrompt,
			UserPrompt:          contract.UserPrompt,
			Temperature:         0.2,
			MaxTokens:           contract.MaxTokens,
			DisableJSONResponse: true,
		})
		if retryErr != nil {
			if isTruncatedDirectorOutput(retryErr) {
				return nil, retryErr
			}
			if parseErr != nil {
				return nil, retryableModelOutputError(parseErr)
			}
			return nil, retryErr
		}
		raw, parseErr = ParseDirectorJSON(retryCompletion)
		if parseErr != nil {
			return nil, retryableModelOutputError(parseErr)
		}
	}
	assets, err := NormalizeAssetExtractionOutput(raw)
	if err != nil {
		return nil, retryableModelOutputError(err)
	}
	// The book classification is a deterministic fallback for an ambiguous
	// first-person role. Apply it before and after optional prompt compilation
	// so the final persisted asset cannot lose the resolved gender anchor.
	assets = applyBookGenderFallback(book, assets)
	// The selected full H3 assets preset already returns detailed appearances
	// together with characters/scenes/props. Other renderer selections retain
	// their historical post-extraction compilation behavior.
	config = aiReasoningPromptConfig(snapshot.Effective)
	if config.Assets.appliesTo(book, config.Assets.Extraction) && usesCMAssetRenderer(config.Assets.Extraction) {
		assets, err = s.extractCMSceneAndPropPrompts(ctx, book, snapshot, config.Assets.Extraction, assets)
		if err != nil {
			return nil, err
		}
	}
	if !(config.Assets.appliesTo(book, config.Assets.Extraction) && usesCombinedAssetRenderer(config.Assets.Extraction)) {
		assets, err = s.compileH3AssetPrompts(ctx, book, snapshot, assets)
		if err != nil {
			return nil, err
		}
	}
	assets = applyBookGenderFallback(book, assets)
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
	// Attempt 1 asks for the full set. Attempt 2 is a single corrective retry:
	// truncation gets the compact-output instruction, a format/parse miss gets
	// an explicit missing-index + marker instruction. Partial successes from
	// either attempt are merged by index, so a model that only fixes 换开头2 on
	// the retry does not lose the already valid 换开头1.
	const maxOpeningVariantAttempts = 2
	// Start from the variants already persisted on VIDEO01. A rerun triggered only
	// to fill a missing slot must not discard slots that are already successful.
	variants := openingVariants(book.Videos[0])
	if len(variants) == 0 {
		variants = nil
	}
	truncatedRetry := false
	for attempt := 1; attempt <= maxOpeningVariantAttempts; attempt++ {
		request := TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens, DisableJSONResponse: true}
		if attempt > 1 {
			if truncatedRetry {
				request.SystemPrompt += fmt.Sprintf("\n上一轮输出因长度截断。请更紧凑地输出全部 %d 个变体，保留完整结尾，不要解释，不要 Markdown。", variantCount)
			} else {
				request.SystemPrompt += openingVariantCorrection(variants, variantCount)
			}
		}
		text, completionErr := s.Provider.Complete(ctx, request)
		if completionErr != nil {
			if isTruncatedDirectorOutput(completionErr) && attempt < maxOpeningVariantAttempts {
				truncatedRetry = true
				continue
			}
			return nil, completionErr
		}
		parsed := parseOpeningVariants(text, snapshot.MaxVideoDuration, first.DurationSec, variantCount)
		variants = mergeOpeningVariants(variants, parsed, variantCount)
		if err := requireSuccessfulOpeningVariants(variants, variantCount); err == nil {
			break
		}
		if attempt < maxOpeningVariantAttempts {
			missing := missingOpeningVariantIndices(variants, variantCount)
			log.Printf("opening variants incomplete attempt=%d missing=%v raw=%s", attempt, missing, directorHTTPBodySnippet([]byte(text), 400))
		}
	}
	merged, err := s.saveOpeningVariants(ctx, owner, batchID, bookID, book.Videos[0].ID, variants)
	if err != nil {
		return nil, err
	}
	if err := requireSuccessfulOpeningVariants(merged, variantCount); err != nil {
		return merged, err
	}
	return merged, nil
}

func mergeOpeningVariants(base, next []OpeningVariant, variantCount int) []OpeningVariant {
	baseByIndex := make(map[int]OpeningVariant, len(base))
	for _, variant := range base {
		baseByIndex[variant.Index] = variant
	}
	nextByIndex := make(map[int]OpeningVariant, len(next))
	for _, variant := range next {
		nextByIndex[variant.Index] = variant
	}
	out := make([]OpeningVariant, 0, variantCount)
	for index := 1; index <= variantCount; index++ {
		if previous, ok := baseByIndex[index]; ok && previous.Status == "success" && strings.TrimSpace(previous.Prompt) != "" {
			out = append(out, previous)
			continue
		}
		if candidate, ok := nextByIndex[index]; ok {
			out = append(out, candidate)
			continue
		}
		if previous, ok := baseByIndex[index]; ok {
			out = append(out, previous)
			continue
		}
		out = append(out, OpeningVariant{Index: index, Label: openingVariantLabel(index), Status: "failed", FailureReason: "模型未输出该变体分段"})
	}
	return out
}

func missingOpeningVariantIndices(variants []OpeningVariant, variantCount int) []int {
	byIndex := make(map[int]OpeningVariant, len(variants))
	for _, variant := range variants {
		byIndex[variant.Index] = variant
	}
	missing := make([]int, 0, variantCount)
	for index := 1; index <= variantCount; index++ {
		variant, ok := byIndex[index]
		if ok && variant.Status == "success" && strings.TrimSpace(variant.Prompt) != "" {
			continue
		}
		missing = append(missing, index)
	}
	return missing
}

func openingVariantCorrection(variants []OpeningVariant, variantCount int) string {
	missing := missingOpeningVariantIndices(variants, variantCount)
	labels := make([]string, 0, len(missing))
	for _, index := range missing {
		labels = append(labels, fmt.Sprintf("换开头%d", index))
	}
	return fmt.Sprintf("\n上一轮输出缺少 %s。请只重新输出这些缺失变体：每个变体必须单独一行写 ===换开头N===，下一行写 时长：X秒，再写完整分镜正文；不要解释，不要 Markdown，不要输出已成功的变体。", strings.Join(labels, "、"))
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
