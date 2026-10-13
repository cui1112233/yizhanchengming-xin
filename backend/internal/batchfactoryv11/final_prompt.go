package batchfactoryv11

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// CompatibilityNotice reports a setting that cannot be applied to the
// current Director/video identity. Notices are read-only: the compiler never
// silently rewrites a user's sparse patch.
type CompatibilityNotice struct {
	Field  string `json:"field"`
	State  string `json:"state"`
	Reason string `json:"reason"`
	Source string `json:"source,omitempty"`
}

type EffectiveSettings struct {
	Values             SettingsPatch         `json:"values"`
	SourceByField      map[string]string     `json:"sourceByField"`
	Compatibility      []CompatibilityNotice `json:"compatibility"`
	SnapshotHash       string                `json:"snapshotHash"`
	DirectorRevisionID string                `json:"directorRevisionId,omitempty"`
}

type PromptComponent struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Content string `json:"content"`
}

type FinalPrompt struct {
	BatchID            string `json:"batchId"`
	BookID             string `json:"bookId"`
	VideoID            string `json:"videoId"`
	DirectorRevisionID string `json:"directorRevisionId"`
	SnapshotHash       string `json:"snapshotHash"`
	// DisplayPrompt is the concise, editable director-card text shown in the
	// workbench. CompiledPrompt is the private submission payload sent to the
	// video provider and may include H3's full visual baseline and safeguards.
	DisplayPrompt string `json:"displayPrompt"`
	// BaseSetupPrompt is the read-only people-and-scenes context displayed with
	// a storyboard card. It is intentionally separate from DisplayPrompt so an
	// edit to a card's VIDEO text cannot overwrite shared asset definitions.
	BaseSetupPrompt       string            `json:"baseSetupPrompt,omitempty"`
	CompiledPrompt        string            `json:"compiledPrompt"`
	Components            []PromptComponent `json:"components"`
	EffectiveSettings     EffectiveSettings `json:"effectiveSettings"`
	DurationSeconds       int               `json:"durationSeconds"`
	ReferenceImageURLs    []string          `json:"referenceImageUrls,omitempty"`
	DowngradedAssetIDs    []string          `json:"downgradedAssetIds,omitempty"`
	CompilationID         string            `json:"compilationId,omitempty"`
	CompilationSegmentKey string            `json:"compilationSegmentKey,omitempty"`
	CompileTrace          *H3CompileTrace   `json:"compileTrace,omitempty"`
	// This is UI metadata only. A placeholder must never enter the VIDEO prompt.
	SmartUnifiedPending bool `json:"smartUnifiedPending,omitempty"`
}

type PromptCompilerService struct{ Store Store }

var systemSettings = SettingsPatch{
	"productionMode":          json.RawMessage(`"original"`),
	"aspectRatio":             json.RawMessage(`"9:16"`),
	"storyboardDurationLimit": json.RawMessage(`10`),
	"maxVideoDuration":        json.RawMessage(`10`),
	"durationMode":            json.RawMessage(`"auto"`),
	"fixedSingleVideo":        json.RawMessage(`false`),
	"prefixMode":              json.RawMessage(`"auto"`),
	"subtitlePolicy":          json.RawMessage(`"forbid-auto-dialogue-subtitle"`),
	"negativeMergeMode":       json.RawMessage(`"append"`),
	"injectBaseSettings":      json.RawMessage(`true`),
	"injectCharacterPrompt":   json.RawMessage(`true`),
	"injectScenePrompt":       json.RawMessage(`true`),
	"injectPropPrompt":        json.RawMessage(`true`),
}

func applyLayer(values SettingsPatch, sources map[string]string, layer SettingsPatch, source string) {
	keys := make([]string, 0, len(layer))
	for key := range layer {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if key == "aiPromptConfig" {
			values[key] = mergePromptConfig(values[key], layer[key])
		} else {
			values[key] = append([]byte(nil), layer[key]...)
		}
		sources[key] = source
	}
}

// mergePromptConfig keeps the batch-level AI modules when a book only changes
// one module (for example its storyboard composition). Arrays deliberately
// replace as a whole, while nested objects overlay their explicit fields.
func mergePromptConfig(base, override json.RawMessage) json.RawMessage {
	var baseObject, overrideObject map[string]json.RawMessage
	if len(base) == 0 || json.Unmarshal(base, &baseObject) != nil || baseObject == nil || json.Unmarshal(override, &overrideObject) != nil || overrideObject == nil {
		return append([]byte(nil), override...)
	}
	merged := make(map[string]json.RawMessage, len(baseObject)+len(overrideObject))
	for key, value := range baseObject {
		merged[key] = append([]byte(nil), value...)
	}
	for key, value := range overrideObject {
		if existing, found := merged[key]; found {
			merged[key] = mergePromptConfig(existing, value)
			continue
		}
		merged[key] = append([]byte(nil), value...)
	}
	encoded, err := json.Marshal(merged)
	if err != nil {
		return append([]byte(nil), override...)
	}
	return encoded
}

func configPatch(raw json.RawMessage) SettingsPatch {
	var direct SettingsPatch
	if json.Unmarshal(raw, &direct) != nil {
		return SettingsPatch{}
	}
	if nested, ok := direct["settings"]; ok {
		var settings SettingsPatch
		if json.Unmarshal(nested, &settings) == nil {
			return settings
		}
	}
	delete(direct, "source")
	return direct
}

func rawStrings(patch SettingsPatch, key string) []string {
	var values []string
	if raw, ok := patch[key]; ok {
		_ = json.Unmarshal(raw, &values)
	}
	return values
}

func videoFromBook(book Book, videoID string) (Video, int, error) {
	for index, video := range book.Videos {
		if video.ID == videoID && video.CompatibilityState == "active" {
			return video, index, nil
		}
	}
	return Video{}, -1, ErrNotFound
}

func snapshotHash(values SettingsPatch, directorRevisionID string) string {
	payload, _ := json.Marshal(struct {
		Values             SettingsPatch `json:"values"`
		DirectorRevisionID string        `json:"directorRevisionId"`
	}{values, directorRevisionID})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func (s *PromptCompilerService) ResolveEffective(ctx context.Context, owner, batchID, bookID, videoID string) (EffectiveSettings, Book, Video, int, error) {
	if s == nil || s.Store == nil {
		return EffectiveSettings{}, Book{}, Video{}, -1, ErrUnavailable
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return EffectiveSettings{}, Book{}, Video{}, -1, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return EffectiveSettings{}, Book{}, Video{}, -1, err
	}
	video, ordinal, err := videoFromBook(book, videoID)
	if err != nil {
		return EffectiveSettings{}, Book{}, Video{}, -1, err
	}

	values, sources := SettingsPatch{}, map[string]string{}
	applyLayer(values, sources, systemSettings, "system")
	versionID := rawString(batch.SettingsState.Patch, "versionConfigId", "")
	if versionID != "" {
		versions, listErr := s.Store.ConfigVersions(ctx, owner)
		if listErr != nil {
			return EffectiveSettings{}, Book{}, Video{}, -1, listErr
		}
		for _, version := range versions {
			if version.ID == versionID {
				applyLayer(values, sources, configPatch(version.Config), "config-version:"+version.ID)
				break
			}
		}
	}
	applyLayer(values, sources, batch.SettingsState.Patch, "batch")
	applyLayer(values, sources, book.SettingsState.Patch, "book")
	applyLayer(values, sources, video.SettingsState.Patch, "video")

	directorID := ""
	if book.DirectorRevision != nil {
		directorID = book.DirectorRevision.ID
	}
	notices := []CompatibilityNotice{}
	if directorID == "" {
		notices = append(notices, CompatibilityNotice{Field: "directorRevisionId", State: "missing", Reason: "当前 VIDEO 没有关联 Director revision"})
	}
	duration := rawInt(values, "duration", int(video.DurationSeconds))
	maxDuration := rawInt(values, "storyboardDurationLimit", 10)
	if duration > 0 && maxDuration > 0 && duration > maxDuration {
		notices = append(notices, CompatibilityNotice{Field: "duration", State: "incompatible", Reason: fmt.Sprintf("VIDEO 时长 %ds 超过当前模型上限 %ds", duration, maxDuration), Source: sources["duration"]})
	}
	if videoAspect := rawString(values, "videoAspectRatio", ""); videoAspect != "" && videoAspect != "9:16" && videoAspect != "16:9" {
		notices = append(notices, CompatibilityNotice{Field: "videoAspectRatio", State: "incompatible", Reason: "当前视频画幅不受支持", Source: sources["videoAspectRatio"]})
	}
	effective := EffectiveSettings{Values: values, SourceByField: sources, Compatibility: notices, DirectorRevisionID: directorID}
	effective.SnapshotHash = snapshotHash(values, directorID)
	return effective, book, video, ordinal, nil
}

func namedPromptMap(values []NamedPrompt) map[string]string {
	out := make(map[string]string, len(values))
	for _, value := range values {
		out[value.Name] = value.Prompt
	}
	return out
}

func (s *PromptCompilerService) namedPromptMapWithDrafts(ctx context.Context, owner, batchID, category string, values []NamedPrompt) (map[string]string, error) {
	out := namedPromptMap(values)
	if s == nil || s.Store == nil {
		return out, nil
	}
	for _, value := range values {
		key := "asset:" + category + ":" + value.Name
		draft, err := s.Store.GetDraft(ctx, owner, key, "asset-prompt", batchID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				continue
			}
			return nil, err
		}
		// A found draft is authoritative even when intentionally blank: the
		// user can clear an inherited asset prompt instead of silently restoring it.
		out[value.Name] = strings.TrimSpace(draft.Content)
	}
	return out, nil
}

func selectPrompts(names []string, library map[string]string) string {
	parts := []string{}
	for _, name := range names {
		if prompt := strings.TrimSpace(library[name]); prompt != "" {
			parts = append(parts, name+"："+prompt)
		}
	}
	return strings.Join(parts, "；")
}

type compiledAsset struct {
	ID       string
	Kind     string
	Name     string
	Prompt   string
	ImageURL string
}

func assetKey(kind, name string) string { return kind + "\x00" + name }

func recordsByKindAndName(book Book, fallback map[string][]NamedPrompt) map[string]compiledAsset {
	records := map[string]compiledAsset{}
	for _, asset := range book.AssetRecords {
		if strings.TrimSpace(asset.Name) == "" || strings.TrimSpace(asset.Kind) == "" {
			continue
		}
		records[assetKey(asset.Kind, asset.Name)] = compiledAsset{ID: asset.ID, Kind: asset.Kind, Name: asset.Name, Prompt: strings.TrimSpace(asset.Prompt)}
	}
	for kind, values := range fallback {
		for _, value := range values {
			key := assetKey(kind, value.Name)
			if _, exists := records[key]; !exists {
				records[key] = compiledAsset{Kind: kind, Name: value.Name, Prompt: strings.TrimSpace(value.Prompt)}
			}
		}
	}
	return records
}

func providerAccessibleImageURL(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
		return ""
	}
	return parsed.String()
}

func (s *PromptCompilerService) primaryAssetImages(ctx context.Context, owner, batchID, bookID string, records map[string]compiledAsset) (map[string]string, error) {
	images := map[string]string{}
	for _, asset := range records {
		if asset.ID == "" {
			continue
		}
		versions, err := s.Store.ListBookAssetImages(ctx, owner, batchID, bookID, asset.ID)
		if err != nil {
			return nil, err
		}
		for _, version := range versions {
			if version.IsPrimary {
				if imageURL := providerAccessibleImageURL(version.URL); imageURL != "" {
					images[asset.ID] = imageURL
				}
				break
			}
		}
	}
	return images, nil
}

func namedPromptNames(items []NamedPrompt) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		if name := strings.TrimSpace(item.Name); name != "" {
			names = append(names, name)
		}
	}
	return names
}

func selectedAssets(kind string, names []string, records map[string]compiledAsset, images map[string]string, selectedIDs map[string]bool, hasSelection bool) []compiledAsset {
	selected := make([]compiledAsset, 0, len(names))
	seen := map[string]bool{}
	for _, name := range names {
		asset, ok := records[assetKey(kind, name)]
		if !ok || (hasSelection && !selectedIDs[asset.ID]) {
			continue
		}
		seenKey := asset.ID
		if seenKey == "" {
			seenKey = assetKey(asset.Kind, asset.Name)
		}
		if seen[seenKey] {
			continue
		}
		asset.ImageURL = images[asset.ID]
		if strings.TrimSpace(asset.Prompt) == "" && asset.ImageURL == "" {
			continue
		}
		selected = append(selected, asset)
		seen[seenKey] = true
	}
	if !hasSelection {
		return selected
	}
	additional := make([]compiledAsset, 0)
	for _, asset := range records {
		if asset.Kind != kind || asset.ID == "" || !selectedIDs[asset.ID] || seen[asset.ID] {
			continue
		}
		asset.ImageURL = images[asset.ID]
		if strings.TrimSpace(asset.Prompt) == "" && asset.ImageURL == "" {
			continue
		}
		additional = append(additional, asset)
	}
	sort.Slice(additional, func(i, j int) bool { return additional[i].Name < additional[j].Name })
	selected = append(selected, additional...)
	return selected
}

func storyboardAssetMentions(groups ...[]compiledAsset) string {
	seen := map[string]bool{}
	mentions := []string{}
	for _, group := range groups {
		for _, asset := range group {
			name := strings.TrimSpace(asset.Name)
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			mentions = append(mentions, "@"+name)
		}
	}
	return strings.Join(mentions, " ")
}

func withStoryboardAssetMentions(prompt, mentions string) string {
	prompt = strings.TrimSpace(prompt)
	mentions = strings.TrimSpace(mentions)
	if prompt == "" || mentions == "" {
		return prompt
	}
	allPresent := true
	for _, mention := range strings.Fields(mentions) {
		if !strings.Contains(prompt, mention) {
			allPresent = false
			break
		}
	}
	if allPresent {
		return prompt
	}
	return prompt + "\n资产：" + mentions
}

func baseSetupEnabled(values SettingsPatch) bool {
	raw, exists := values["aiPromptConfig"]
	if !exists {
		return true
	}
	var config struct {
		Constraints struct {
			Enabled   *bool `json:"enabled"`
			BaseSetup struct {
				Enabled *bool `json:"enabled"`
			} `json:"baseSetup"`
		} `json:"constraints"`
	}
	if json.Unmarshal(raw, &config) != nil {
		return true
	}
	if config.Constraints.Enabled != nil && !*config.Constraints.Enabled {
		return false
	}
	if config.Constraints.BaseSetup.Enabled == nil {
		return true
	}
	return *config.Constraints.BaseSetup.Enabled
}

// constraintsApply is deliberately stricter than the historical per-field
// flags. Once an AI-reasoning constraint module is saved and switched off,
// none of its layers (including base assets) may silently survive in VIDEO
// compilation. Legacy books without that module retain their old settings.
func constraintsApply(values SettingsPatch, config AIReasoningPromptConfig, book Book) bool {
	if _, configured := values["aiPromptConfig"]; !configured {
		return true
	}
	return config.Constraints.appliesTo(book)
}

func constraintCategory(selection PresetSnapshot) string {
	if category := strings.TrimSpace(selection.ConstraintCategory); category != "" {
		return category
	}
	return strings.TrimPrefix(strings.TrimSpace(selection.Slot), "script.constraint.")
}

func selectedConstraintBody(config AIReasoningPromptConfig, book Book, category string) string {
	if !config.Constraints.appliesTo(book) {
		return ""
	}
	for _, selection := range config.Constraints.Selections {
		if constraintCategory(selection) == category {
			return strings.TrimSpace(selection.Body)
		}
	}
	return ""
}

const smartUnifiedPrefixPresetID = "script-constraint-prefix-smart-unified"

func smartUnifiedStyleForRevision(book Book, config AIReasoningPromptConfig) string {
	if !smartUnifiedSelectedForRevision(book, config) {
		return ""
	}
	if book.DirectorRevision != nil && strings.TrimSpace(book.DirectorRevision.Output.SmartUnifiedStyle) != "" {
		return strings.TrimSpace(book.DirectorRevision.Output.SmartUnifiedStyle)
	}
	if analysis, err := parseSmartUnifiedAnalysis(rawString(book.SettingsState.Patch, "h3StyleAnalysis", "")); err == nil && analysis != nil {
		if selectedSDVideoPreset(config.Video) && strings.TrimSpace(analysis.SDUnifiedStyle) != "" {
			return strings.TrimSpace(analysis.SDUnifiedStyle)
		}
		return analysis.Prompt
	}
	return ""
}

func smartUnifiedSelectedForRevision(book Book, config AIReasoningPromptConfig) bool {
	if !config.Constraints.appliesTo(book) {
		return false
	}
	for _, selection := range config.Constraints.Selections {
		if constraintCategory(selection) == "prefix" && selection.ID == smartUnifiedPrefixPresetID {
			return true
		}
	}
	return false
}

func allEnabledBookAssets(kind string, records map[string]compiledAsset, excluded map[string]bool) []compiledAsset {
	assets := make([]compiledAsset, 0)
	for _, asset := range records {
		if asset.Kind != kind || asset.ID == "" || excluded[asset.ID] || strings.TrimSpace(asset.Prompt) == "" {
			continue
		}
		assets = append(assets, asset)
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].Name < assets[j].Name })
	return assets
}

func selectResolvedPrompts(assets []compiledAsset, included map[string]bool) string {
	parts := []string{}
	for _, asset := range assets {
		if asset.ID != "" && included[asset.ID] {
			// 图片确实作为参考输入发给了视频模型：走图片生视频，
			// 不再重复发文字设定。图片因上限没发出去时，文字仍兜底保留。
			continue
		}
		if prompt := strings.TrimSpace(asset.Prompt); prompt != "" {
			parts = append(parts, asset.Name+"："+prompt)
		}
	}
	return strings.Join(parts, "；")
}

// assetsWithoutImages drops assets whose image is actually sent, so H3 subject
// definitions are emitted only for text-to-video (or limit-dropped) assets.
func assetsWithoutImages(assets []compiledAsset, included map[string]bool) []compiledAsset {
	out := make([]compiledAsset, 0, len(assets))
	for _, asset := range assets {
		if asset.ID != "" && included[asset.ID] {
			continue
		}
		out = append(out, asset)
	}
	return out
}

func referenceImageLimit(values SettingsPatch) int {
	limit := rawInt(values, "referenceImageLimit", 3)
	if limit < 0 {
		return 0
	}
	if limit > 16 {
		return 16
	}
	return limit
}

func addComponent(out *[]PromptComponent, key, label, content string) {
	content = strings.TrimSpace(content)
	if content != "" {
		*out = append(*out, PromptComponent{Key: key, Label: label, Content: content})
	}
}

func nonEmptyStrings(values []string) []string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func (s *PromptCompilerService) Compile(ctx context.Context, owner, batchID, bookID, videoID string) (FinalPrompt, error) {
	// Preview is deliberately permissive: an old storyboard that no longer
	// matches a newly selected duration rule must still show its compiled
	// prompt and its compatibility warning in the card. Submission keeps the
	// strict validation path below.
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return FinalPrompt{}, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return FinalPrompt{}, err
	}
	if book.DirectorRevision != nil && book.DirectorRevision.Output.H3Director != nil {
		repository, ok := s.Store.(H3Repository)
		if !ok {
			return FinalPrompt{}, fmt.Errorf("%w: H3 compilation repository is unavailable", ErrUnavailable)
		}
		latest, latestErr := repository.LatestH3VideoCompilation(ctx, owner, batchID, bookID, book.DirectorRevision.ID)
		if latestErr == nil {
			return s.CompileH3ForProduction(ctx, owner, batchID, bookID, videoID, latest.ID, -1)
		}
		if !errors.Is(latestErr, ErrNotFound) {
			return FinalPrompt{}, latestErr
		}
	}
	return s.compile(ctx, owner, batchID, bookID, videoID, -1, false, 0)
}

// CompileForProduction applies all effective-setting compatibility checks
// before a provider task can be created. It is used by adapters that do not
// expose a concrete reference-image limit.
func (s *PromptCompilerService) CompileForProduction(ctx context.Context, owner, batchID, bookID, videoID string) (FinalPrompt, error) {
	return s.compile(ctx, owner, batchID, bookID, videoID, -1, true, 0)
}

// CompileWithReferenceImageLimit is used by production after the concrete
// video model is frozen. Preview compilation keeps the saved per-book limit.
func (s *PromptCompilerService) CompileWithReferenceImageLimit(ctx context.Context, owner, batchID, bookID, videoID string, limit int) (FinalPrompt, error) {
	return s.compile(ctx, owner, batchID, bookID, videoID, limit, true, 0)
}

// CompileForOpeningVariant compiles VIDEO01 with the given opening variant
// prompt replacing the original director prompt. The variant must exist in the
// video's settings patch with status "success".
func (s *PromptCompilerService) CompileForOpeningVariant(ctx context.Context, owner, batchID, bookID, videoID string, variantIndex, frozenReferenceLimit int) (FinalPrompt, error) {
	return s.compile(ctx, owner, batchID, bookID, videoID, frozenReferenceLimit, true, variantIndex)
}

func (s *PromptCompilerService) compile(ctx context.Context, owner, batchID, bookID, videoID string, frozenReferenceLimit int, strict bool, variantIndex int) (FinalPrompt, error) {
	effective, book, video, ordinal, err := s.ResolveEffective(ctx, owner, batchID, bookID, videoID)
	if err != nil {
		return FinalPrompt{}, err
	}
	if effective.DirectorRevisionID == "" || book.DirectorRevision == nil || ordinal >= len(book.DirectorRevision.Output.Storyboard) {
		return FinalPrompt{}, fmt.Errorf("%w: active Director revision is required", ErrConflict)
	}
	if strict {
		for _, notice := range effective.Compatibility {
			if notice.State == "incompatible" {
				return FinalPrompt{}, fmt.Errorf("%w: incompatible effective settings", ErrConflict)
			}
		}
	}
	draft := book.DirectorRevision.Output.Storyboard[ordinal]
	if variantIndex > 0 {
		if ordinal != 0 {
			return FinalPrompt{}, fmt.Errorf("%w: opening variants only apply to VIDEO01", ErrInvalid)
		}
		variant, found := openingVariantFor(video, variantIndex)
		if !found {
			return FinalPrompt{}, fmt.Errorf("%w: opening variant %d is not available", ErrConflict, variantIndex)
		}
		draft.FinalPrompt = variant.Prompt
		draft.Shots = nil
	}
	values := effective.Values
	characterRefs := rawStrings(values, "characterRefs")
	if len(characterRefs) == 0 {
		characterRefs = draft.Characters
	}
	if len(characterRefs) == 0 {
		// SD 纯文本分镜没有逐镜人物字段，也没有 VIDEO 级手动选择时，
		// 默认带整本书已提取的全部人物，跟剧本生成一致。
		characterRefs = namedPromptNames(book.Assets.Characters)
	}
	sceneRefs := rawStrings(values, "sceneRefs")
	if len(sceneRefs) == 0 && draft.Scene != "" {
		sceneRefs = []string{draft.Scene}
	}
	if len(sceneRefs) == 0 {
		sceneRefs = namedPromptNames(book.Assets.Scenes)
	}
	propRefs := rawStrings(values, "propRefs")
	if len(propRefs) == 0 {
		propRefs = draft.Props
	}
	if len(propRefs) == 0 {
		propRefs = namedPromptNames(book.Assets.Props)
	}

	records := recordsByKindAndName(book, map[string][]NamedPrompt{
		"character": book.Assets.Characters,
		"scene":     book.Assets.Scenes,
		"prop":      book.Assets.Props,
	})
	primaryImages, err := s.primaryAssetImages(ctx, owner, batchID, bookID, records)
	if err != nil {
		return FinalPrompt{}, err
	}
	selection, hasSelection := decodeVideoAssetSelection(video.SettingsState.Patch)
	selectedIDs := selection.effectiveAssetIDSet()
	characters := selectedAssets("character", characterRefs, records, primaryImages, selectedIDs, hasSelection)
	scenes := selectedAssets("scene", sceneRefs, records, primaryImages, selectedIDs, hasSelection)
	props := selectedAssets("prop", propRefs, records, primaryImages, selectedIDs, hasSelection)
	assetMentions := storyboardAssetMentions(characters, scenes, props)
	limit := referenceImageLimit(values)
	if frozenReferenceLimit >= 0 {
		limit = frozenReferenceLimit
	}
	referenceImages := []string{}
	for _, visualImage := range append(rawStrings(values, "imageUrls"), rawStrings(values, "referenceImages")...) {
		if url := providerAccessibleImageURL(visualImage); url != "" && len(referenceImages) < limit {
			referenceImages = append(referenceImages, url)
		}
	}
	included := map[string]bool{}
	downgraded := []string{}
	for _, asset := range append(append(characters, scenes...), props...) {
		if asset.ImageURL == "" {
			continue
		}
		if len(referenceImages) < limit {
			referenceImages = append(referenceImages, asset.ImageURL)
			included[asset.ID] = true
		} else if asset.ID != "" {
			downgraded = append(downgraded, asset.ID)
		}
	}
	components := []PromptComponent{}
	// visualPrompt is an image-generation input. It intentionally never enters
	// the final video prompt: image composition and motion/video direction are
	// separate controls, and mixing them makes a regenerated still silently
	// change the video request.
	// Base setup controls text fallback for assets that have no usable image.
	// Existing asset images remain reference inputs regardless of this switch.
	// prefix_key was emitted by the retired video-style-prefix flow. Ignore it
	// for both new and historical director revisions; the active picture-prefix
	// constraint below remains the sole prefix source for VIDEO compilation.
	config := aiReasoningPromptConfig(values)
	h3Renderer := usesH3VideoRenderer(config.Video)
	// SD 直出：分镜卡正文即 AI 按预设模板写好的提示词，外层开关组件
	// （智能统一、基础设定、画质约束、负面提示词）照常注入，与 SD 模板
	// 移除的统一风格/统一人物/画质/负面段落互补，由开关控制是否显示。
	sdDirect := strings.TrimSpace(draft.FinalPrompt) != ""
	constraintsActive := constraintsApply(values, config, book)
	constraintBody := func(category, legacyEnabledKey, legacyValueKey string) string {
		if category == "prefix" && smartUnifiedSelectedForRevision(book, config) {
			// "智能统一" is a meta prompt that asks the model to analyse the
			// whole book. Its own body is never a VIDEO prefix. Until a director
			// generation produces the analysis, leave this component empty and
			// report SmartUnifiedPending to the UI instead of falling back to it.
			return smartUnifiedStyleForRevision(book, config)
		}
		if constraintsActive {
			if body := selectedConstraintBody(config, book, category); body != "" {
				return body
			}
		}
		if constraintsActive && rawBool(values, legacyEnabledKey, false) {
			return strings.TrimSpace(rawString(values, legacyValueKey, ""))
		}
		return ""
	}
	// The final VIDEO request has one stable, user-visible sequence. Transport
	// fields such as duration and subtitle policy are passed separately to the
	// provider and must not masquerade as prompt content.
	// 画面前缀统一由约束开关注入。SD 的兼容卡片把它和画质规则合并
	// 成一个“画面质量提示词”段，保留旧卡片的阅读和提交顺序；H3 仍以
	// 独立前缀组件编译，保持其现有时间轴协议不变。
	prefixPrompt := constraintBody("prefix", "prefixEnabled", "prefix")
	if !sdDirect && (smartUnifiedSelectedForRevision(book, config) || rawBool(values, "prefixEnabled", false)) {
		addComponent(&components, "prefix", "画面前缀", prefixPrompt)
	}
	// Base setup is its own user switch. A batch may choose an H3 VIDEO preset
	// before opening the constraint editor; that must retain the legacy default
	// of including asset definitions rather than treating an absent module as off.
	baseSetupActive := baseSetupEnabled(values) && rawBool(values, "injectBaseSettings", true)
	baseSetupPrompt := ""
	if baseSetupActive {
		base := []string{}
		if value := selectResolvedPrompts(characters, included); value != "" {
			base = append(base, "人物："+value)
		}
		if value := selectResolvedPrompts(scenes, included); value != "" {
			base = append(base, "场景："+value)
		}
		baseSetupPrompt = strings.Join(base, "\n")
		if value := selectResolvedPrompts(props, included); value != "" {
			base = append(base, "道具："+value)
		}
		// 基础设定统一由开关注入，H3 不再跳过。H3 下的人物定义保留 <Subject N>
		// 格式（供时间轴引用），但作为外层“基础设定”组件按开关注入；非 H3 沿用
		// 通用“人物/场景/道具”文本。图片真发出去的资产不出文字（图片生视频）。
		if h3Renderer {
			if definitions := h3SubjectDefinitions(h3Subjects(assetsWithoutImages(characters, included))); definitions != "" {
				addComponent(&components, "baseSetup", "基础设定", definitions)
			}
		} else {
			addComponent(&components, "baseSetup", "基础设定", strings.Join(base, "\n"))
		}
	}
	var displayPrompt, videoPrompt, videoLabel string
	if sdDirect {
		videoPrompt = draft.FinalPrompt
		displayPrompt = videoPrompt
		videoLabel = "视频提示词"
		// 旧 SD 卡片始终以“基础设定 → 画面质量提示词 → 视频提示词”
		// 展示和提交。智能统一/普通前缀与画质约束都属于同一视觉质量段，
		// 不再在 SD 卡片顶层露出一个“画面前缀”。
		visualQuality := []string{strings.TrimSpace(prefixPrompt), strings.TrimSpace(constraintBody("quality", "qualityEnabled", "quality"))}
		addComponent(&components, "quality", "画面质量提示词", strings.Join(nonEmptyStrings(visualQuality), "\n"))
		addComponent(&components, "video", videoLabel, videoPrompt)
		addComponent(&components, "restriction", "画面限制", constraintBody("restriction", "restrictionEnabled", "restriction"))
		addComponent(&components, "negative", "负面提示词", constraintBody("negative", "negativeEnabled", "negative"))
	} else {
		addComponent(&components, "quality", "画面约束提示词", constraintBody("quality", "qualityEnabled", "quality"))
		displayPrompt = storyboardVideoPrompt(draft)
		videoPrompt = displayPrompt
		if h3Renderer {
			// compileH3VideoPrompt 只产出 H3 骨架（视听呈现+时间轴）；人物定义、画风
			// 由外层开关统一注入。基础设定开关关闭时不传人物/场景，时间轴里的
			// <Subject N> 指代也随之取消。H3 编译结果是程序生成的结构化内容，不剥离。
			h3Characters, h3Scenes := characters, scenes
			if !baseSetupActive {
				h3Characters, h3Scenes = nil, nil
			}
			videoPrompt = compileH3VideoPrompt(draft, h3Characters, h3Scenes)
		} else {
			// 非 H3 正文先剥离模型/粘贴自带的共享设定段，再由开关统一注入，对齐剧本生成。
			videoPrompt = stripSharedPromptSections(videoPrompt)
			displayPrompt = videoPrompt
		}
		// A video-scope edit is the sole exception: it is an intentional user
		// replacement for this one card.  Stored pre-fix video_desc summaries are
		// not overrides and must never win over the director's verified timeline.
		// 用户手动覆盖的非 H3 文本同样剥离共享段，防止绕过开关与程序注入重复。
		if override, ok := optionalPromptValue(video.SettingsState.Patch, "videoPrompt"); ok {
			if !h3Renderer {
				override = stripSharedPromptSections(override)
			}
			videoPrompt = override
			displayPrompt = override
		}
		videoLabel = "视频提示词"
		if h3Renderer {
			// H3's canonical output already has a top-level grammar. Adding the V11
			// card label would make the provider receive a wrapper H3 never used.
			videoLabel = ""
		}
		addComponent(&components, "video", videoLabel, videoPrompt)
		addComponent(&components, "restriction", "画面限制", constraintBody("restriction", "restrictionEnabled", "restriction"))
		addComponent(&components, "negative", "负面提示词", constraintBody("negative", "negativeEnabled", "negative"))
	}
	displayPrompt = withStoryboardAssetMentions(displayPrompt, assetMentions)
	videoPrompt = withStoryboardAssetMentions(videoPrompt, assetMentions)
	for index := range components {
		if components[index].Key == "video" {
			components[index].Content = videoPrompt
		}
	}
	duration := rawInt(values, "duration", int(video.DurationSeconds))

	lines := make([]string, 0, len(components))
	for _, component := range components {
		if component.Label == "" {
			lines = append(lines, component.Content)
			continue
		}
		lines = append(lines, component.Label+"："+component.Content)
	}
	compiled := strings.Join(lines, "\n")
	return FinalPrompt{BatchID: batchID, BookID: bookID, VideoID: videoID, DirectorRevisionID: effective.DirectorRevisionID, SnapshotHash: effective.SnapshotHash, DisplayPrompt: displayPrompt, BaseSetupPrompt: baseSetupPrompt, CompiledPrompt: compiled, Components: components, EffectiveSettings: effective, DurationSeconds: duration, ReferenceImageURLs: referenceImages, DowngradedAssetIDs: downgraded, SmartUnifiedPending: smartUnifiedSelectedForRevision(book, config) && smartUnifiedStyleForRevision(book, config) == ""}, nil
}
