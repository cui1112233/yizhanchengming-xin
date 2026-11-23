package batchfactoryv11

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type H3DirectorPreset struct {
	Key        string `json:"key"`
	Revision   int64  `json:"revision"`
	PromptBody string `json:"prompt_body,omitempty"`
}

type H3DirectorRunRequest struct {
	VideoSource       H3VideoSource    `json:"video_source"`
	Preset            H3DirectorPreset `json:"preset"`
	SmartUnifiedStyle string           `json:"smart_unified_style,omitempty"`
}

// H3DirectorCharacterAsset is the only persisted-asset context sent to the
// director model. The director needs stable character IDs for slot binding;
// scenes and props remain in the book library and are injected later only when
// the final-prompt switches permit them.
type H3DirectorCharacterAsset struct {
	AssetID       string `json:"asset_id"`
	CanonicalName string `json:"canonical_name"`
	Appearance    string `json:"appearance"`
	Revision      int64  `json:"revision"`
}

func (s *DirectorService) RunH3Director(ctx context.Context, owner, batchID, bookID string, request H3DirectorRunRequest) (DirectorRevision, error) {
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
	if _, pinned := ctx.Value(preparedProductionContextKey{}).(preparedProductionVersion); pinned {
		if _, _, err := checkedProductionSource(ctx, s.Store, owner, batchID, book); err != nil {
			return DirectorRevision{}, err
		}
	}
	source, err := normalizeH3VideoSource(request.VideoSource)
	if err != nil {
		return DirectorRevision{}, err
	}
	if stage, _ := book.SourceMetadata["batchSensitiveStage"].(map[string]any); stage["status"] == "done" && !strings.Contains(strings.Join(h3NonEmptyVideoSourceLines(book.SourceText), "\n"), source.Text) {
		return DirectorRevision{}, NewConfigRequiredError("导演输入与当前生产文本不一致，请重新选择正文")
	}
	if strings.TrimSpace(request.Preset.Key) == "" || request.Preset.Revision <= 0 {
		return DirectorRevision{}, fmt.Errorf("%w: H3 director preset key and revision are required", ErrInvalid)
	}
	snapshot, err := h3SnapshotForBook(batch, book)
	if err != nil {
		return DirectorRevision{}, err
	}
	knownCharacters := h3DirectorCharacterAssets(book.AssetRecords)
	analysis, err := parseSmartUnifiedAnalysis(request.SmartUnifiedStyle)
	if err != nil {
		return DirectorRevision{}, err
	}
	contract := buildH3DirectorContract(source, request.Preset, knownCharacters)
	if analysis != nil {
		contract.UserPrompt += "\n\n已冻结的全片统一视觉风格（只用于导演一致性，不得改写剧情事实）：\n" + analysis.Prompt
	}
	completion, err := s.Provider.Complete(ctx, contract)
	if err != nil {
		return DirectorRevision{}, err
	}
	raw, err := ParseDirectorJSON(completion)
	if err != nil {
		return DirectorRevision{}, retryableModelOutputError(err)
	}
	raw, err = bindH3DirectorAssets(raw, book.AssetRecords)
	if err != nil {
		return DirectorRevision{}, retryableModelOutputError(err)
	}
	document, err := ParseH3DirectorDocument(raw, source)
	if err != nil {
		return DirectorRevision{}, retryableModelOutputError(err)
	}
	if document.DirectorPresetKey != request.Preset.Key || document.DirectorPresetRevision != request.Preset.Revision {
		return DirectorRevision{}, retryableModelOutputError(fmt.Errorf("H3 director output preset does not match frozen preset"))
	}
	if h3VisualBaselineText(document.VisualBaseline) == "" {
		return DirectorRevision{}, retryableModelOutputError(fmt.Errorf("H3 visual_baseline is required"))
	}
	if analysis != nil {
		// The style request is an independently frozen upstream result.  The
		// director may consume it for consistency but may not replace it.
		document.VisualBaseline = analysis.Prompt
	}
	result := DirectorResult{H3Director: &document}
	if analysis != nil {
		result.SmartUnifiedStyle = analysis.Prompt
		result.SmartUnifiedAnalysis = analysis
	}
	return s.Store.PersistDirectorRevision(ctx, owner, book, snapshot, source.Hash, "", result)
}

func h3DirectorCharacterAssets(assets []BookAsset) []H3DirectorCharacterAsset {
	characters := make([]H3DirectorCharacterAsset, 0, len(assets))
	for _, asset := range assets {
		if asset.Kind != "character" || strings.TrimSpace(asset.ID) == "" || strings.TrimSpace(asset.Name) == "" || strings.TrimSpace(asset.Prompt) == "" {
			continue
		}
		characters = append(characters, H3DirectorCharacterAsset{AssetID: asset.ID, CanonicalName: asset.Name, Appearance: asset.Prompt, Revision: asset.Revision})
	}
	return characters
}

// RunConfiguredH3Director is the stage-runner entrypoint for new V12 runs.
// The processed production text is frozen as the H3 video source; the selected
// H3 VIDEO preset opts into this path, while the director preset body remains
// reusable independently of the final VIDEO renderer.
func (s *DirectorService) RunConfiguredH3Director(ctx context.Context, owner, batchID, bookID, smartUnifiedStyle string) (DirectorRevision, error) {
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
	effective := ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	config := aiReasoningPromptConfig(effective)
	lines := h3NonEmptyVideoSourceLines(book.SourceText)
	if len(lines) == 0 {
		return DirectorRevision{}, fmt.Errorf("%w: processed video source has no non-empty lines", ErrInvalid)
	}
	normalizedSource := strings.Join(lines, "\n")
	hash := sourceDigest(normalizedSource)
	// The selected VIDEO preset owns the director's output rules.  The common
	// structured contract below is the shared Batch Factory pipeline boundary.
	directorPreset := config.Video
	if !directorPreset.videoAppliesTo(book) {
		// Existing V12 records did not persist a video selection. Keep their
		// frozen director preset readable rather than silently changing it.
		directorPreset = config.OriginalDirector
	}
	revision := int64(directorPreset.Version)
	if revision <= 0 {
		revision = int64(config.OriginalDirector.Version)
	}
	if revision <= 0 {
		revision = 1
	}
	return s.RunH3Director(ctx, owner, batchID, bookID, H3DirectorRunRequest{
		VideoSource:       H3VideoSource{Revision: "video-source-" + hash[:16] + "-r1", Hash: hash, Text: normalizedSource},
		Preset:            H3DirectorPreset{Key: configuredVideoPresetKey(directorPreset), Revision: revision, PromptBody: directorRulesFromVideoPreset(directorPreset.body())},
		SmartUnifiedStyle: strings.TrimSpace(smartUnifiedStyle),
	})
}

func directorRulesFromVideoPreset(body string) string {
	trimmed := strings.TrimSpace(body)
	cutAt := len(trimmed)
	// The original H3 normal template renders the final video prompt. It is not
	// a director-rule section and conflicts with the persisted V12 H3 document
	// that this stage must return. Keep the preceding original outline rules.
	for _, marker := range []string{
		"## 原 H3：h3.normal.template",
		"## 原 H3：h3.normal.template__",
		"【批量工厂最终 Prompt 模板】",
		"## 最终 Prompt 模板",
	} {
		if index := strings.Index(trimmed, marker); index >= 0 && index < cutAt {
			cutAt = index
		}
	}
	return strings.TrimSpace(trimmed[:cutAt])
}

func configuredVideoPresetKey(preset AIReasoningPromptModule) string {
	if key := strings.TrimSpace(preset.Key); key != "" {
		return key
	}
	if id := strings.TrimSpace(preset.ID); id != "" {
		return id
	}
	return "h3-director-normal"
}

func normalizeH3VideoSource(source H3VideoSource) (H3VideoSource, error) {
	source.Revision = strings.TrimSpace(source.Revision)
	if source.Revision == "" {
		return H3VideoSource{}, fmt.Errorf("%w: H3 video source revision is required", ErrInvalid)
	}
	lines := h3NonEmptyVideoSourceLines(source.Text)
	if len(lines) == 0 {
		return H3VideoSource{}, fmt.Errorf("%w: processed video source has no non-empty lines", ErrInvalid)
	}
	normalizedText := strings.Join(lines, "\n")
	computedHash := sourceDigest(normalizedText)
	if supplied := strings.TrimSpace(source.Hash); supplied != "" && supplied != computedHash {
		return H3VideoSource{}, fmt.Errorf("%w: H3 video source hash does not match processed non-empty lines", ErrInvalid)
	}
	return H3VideoSource{Revision: source.Revision, Hash: computedHash, Text: normalizedText}, nil
}

func h3SnapshotForBook(batch Batch, book Book) (DirectorSnapshot, error) {
	effective := ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	maxDuration := rawInt(effective, "storyboardDurationLimit", 10)
	if maxDuration != 10 && maxDuration != 15 {
		return DirectorSnapshot{}, fmt.Errorf("%w: storyboardDurationLimit must be 10 or 15", ErrInvalid)
	}
	aspect := EffectiveVideoAspectRatio(effective)
	return DirectorSnapshot{Effective: effective, Mode: "h3", MaxVideoDuration: maxDuration, AspectRatio: aspect}, nil
}

func buildH3DirectorContract(source H3VideoSource, preset H3DirectorPreset, knownCharacters []H3DirectorCharacterAsset) TextCompletionRequest {
	schema := `你是 H3 结构化导演内核。只输出一个 JSON 对象，不要 Markdown。
必须输出 schema_version="h3-director/v1"、writer="batch-factory-v12"、video_source_revision、video_source_hash、video_source_non_empty_line_count、director_preset_key、director_preset_revision、visual_baseline、character_roster[]、director_cards[]。
character_roster 每项={"slot_id":"C001","slot_token":"S1","canonical_name":"人物资产姓名","asset_id":"输入的人物资产ID","aliases":[]}；不得独立生成外形，appearance 由后端从绑定资产填入。只引用已有资产；不存在或有歧义时不得编造资产ID。
如果已有人物资产中存在同一人，canonical_name 必须优先使用已有人物资产的精确 name，将原文中的称呼、简称和关系称呼放入 aliases。不得把“江小姐”之类称呼另建为与真实姓名脱节的人物。
每个非空视频原文行必须且只能对应一张 director_card，顺序一致；卡内保存 source_index/source_key/source_text/source_text_hash/visual_context/preferred_duration/duration_weight/character_slot_ids/action/camera/movement/rhythm/audio/continuity/micro_shots。
character_slot_ids 必须显式输出数组，无人镜头使用 []；非空引用必须指向 character_roster.slot_id。continuity 必须是结构化对象，包含 scene_id/location/axis/light_direction/positions/facings/gazes/held_props/action_ends，禁止用 scene_memory 字符串替代。不得使用 ?、??、unknown、未知人物作为 slot_id 或连续性对象的键；没有可解析人物时这些对象必须输出 {}。
每张卡至少一个 micro_shot，微镜头必须包含 micro_shot_key/weight/shot_task/visual/action/character_slot_ids/camera/movement/rhythm/audio。
director_card 和 micro_shot 的复合字段必须使用以下 JSON 对象形状，禁止写成字符串：
camera={"shot_size":"...","shot_angle":"...","framing":"..."}
movement={"camera_movement":"...","subject_movement":"...","transition":"..."}
audio={"mode":"...","speaker_slot_id":"","dialogue":"","voice_over":"","sound_effects":[],"ambience":[]}
continuity={"scene_id":"...","location":"...","axis":"...","light_direction":"...","positions":{},"facings":{},"gazes":{},"held_props":{},"action_ends":{}}
preferred_duration、duration_weight 和 micro_shot.weight 必须使用 JSON 数字（例如 2.5），不得使用字符串、单位、空值或文字等级。
AI 只给出 preferred_duration 和正数 duration_weight 表达语义节奏；不得输出最终秒数、最终分段或随机 VIDEO 时长。视觉基线、人物分析和场景连续性必须始终生成并保存。`
	if body := strings.TrimSpace(preset.PromptBody); body != "" {
		schema += "\n\n导演预设补充约束：\n" + body
	}
	// Imported H3 presets retain the original filmmaking rules, but their old
	// terminal renderer (timeline_segments and Go-template placeholders) cannot
	// coexist with Batch Factory's durable H3 director document. Put this last
	// so it remains the single output authority even if an old rule calls itself
	// "highest priority".
	schema += "\n\n批量工厂运行期输出适配（最高优先级）：\n只输出 h3-director/v1 JSON 对象，不得输出 timeline_segments、timeline_segment、Markdown、Go 模板变量或最终 VIDEO Prompt。原 H3 规则中关于镜头质量、剧情可视化、微镜头、动作、机位、运镜、声音和连续性的要求必须保留；其历史终端输出格式仅映射为本协议的 director_cards[] 与 micro_shots[]。"
	meta, _ := json.Marshal(map[string]any{
		"video_source_revision":    source.Revision,
		"video_source_hash":        source.Hash,
		"director_preset_key":      preset.Key,
		"director_preset_revision": preset.Revision,
		"known_character_assets":   knownCharacters,
	})
	return TextCompletionRequest{
		SystemPrompt: schema,
		UserPrompt:   string(meta) + "\n\n以下是 H3 已处理的视频原文，一行一张导演卡。director_cards 必须与下列顺序逐项对应；source_index 固定从 1 起（第一行=1，绝不能从 0 起）：\n" + source.Text,
		Temperature:  0.1,
		MaxTokens:    32000,
	}
}
