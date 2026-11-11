package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/prompt/luoshui"
)

// MaxAssetConcurrency 对应规格 §1.4：中途最多 20 路并发。
const MaxAssetConcurrency = 20

// AssetChunkSize 对应规格 §1.4：每 5 章（或每 5 集）一批。
const AssetChunkSize = 5

type AssetRequest struct {
	Owner       string      `json:"owner"`
	ProjectID   string      `json:"project_id"`
	Type        AssetType   `json:"type"`
	Degree      AssetDegree `json:"degree"`
	Content     string      `json:"content"`
	Chapters    []string    `json:"chapters"`
	Model       string      `json:"model"`
	Concurrency int         `json:"concurrency"`
	Confirmed   bool        `json:"confirmed"`
}

type AssetResult struct {
	Batch                AssetBatch      `json:"batch"`
	Assets               []Asset         `json:"assets"`
	ChunkCount           int             `json:"chunk_count"`
	FailedChunks         int             `json:"failed_chunks"`
	Warnings             []string        `json:"warnings,omitempty"`
	PartialFailure       bool            `json:"partial_failure"`
	FailedBatches        []AssetBatchFailure `json:"failed_batches,omitempty"`
	RequiresConfirmation bool            `json:"requires_confirmation"`
	PendingActions       []PendingAction `json:"pending_actions,omitempty"`
}

// AssetBatchFailure 记录失败批次的位置和原因。批次号与章节区间一起返回，
// 上层可以直接提示「第 N 批（第 X-Y 章）失败」，不必再从 warnings 文本里猜。
type AssetBatchFailure struct {
	BatchIndex   int    `json:"batch_index"`
	ChapterStart int    `json:"chapter_start"`
	ChapterEnd   int    `json:"chapter_end"`
	Error        string `json:"error"`
}

// AssetService 实现能力 3/4/5：人物（含音色）、场景、道具提取，四档可切。
type AssetService struct {
	Store     Store
	Generator TextGenerator
	Prompts   *luoshui.Registry
}

func (s *AssetService) Extract(ctx context.Context, req AssetRequest) (AssetResult, error) {
	if !validAssetType(req.Type) {
		return AssetResult{}, fmt.Errorf("%w: 未知资产类型 %q", ErrInvalidArgument, req.Type)
	}
	if !validDegree(req.Degree) {
		return AssetResult{}, fmt.Errorf("%w: 未知档位 %q", ErrInvalidArgument, req.Degree)
	}
	project, err := s.Store.GetProject(ctx, req.Owner, req.ProjectID)
	if err != nil {
		return AssetResult{}, err
	}
	if plan, blocked := pendingPlan(project, req.Confirmed, "3_assets",
		fmt.Sprintf("提取%s（%s）", assetTypeLabel(req.Type), req.Degree),
		"调用资产提取提示词并写入资产库"); blocked {
		return AssetResult{Batch: AssetBatch{}, RequiresConfirmation: true, PendingActions: plan.PendingActions}, nil
	}

	chunks, err := s.resolveChunks(ctx, req)
	if err != nil {
		return AssetResult{}, err
	}
	promptID, err := assetPromptID(req.Type, req.Degree)
	if err != nil {
		return AssetResult{}, err
	}
	systemPrompt, ok := s.prompts().RuntimeBody(promptID)
	if !ok {
		return AssetResult{}, fmt.Errorf("%w: 缺少提示词 %s", ErrInvalidArgument, promptID)
	}

	concurrency := req.Concurrency
	if concurrency <= 0 || concurrency > MaxAssetConcurrency {
		concurrency = MaxAssetConcurrency
	}
	results := make([]chunkOutcome, len(chunks))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for index, chunk := range chunks {
		wg.Add(1)
		go func(index int, chunk string) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[index] = chunkOutcome{err: ctx.Err()}
				return
			}
			if err := ctx.Err(); err != nil {
				results[index] = chunkOutcome{err: err}
				return
			}
			assets, err := s.extractChunk(ctx, req, promptID, systemPrompt, chunk, index+1)
			results[index] = chunkOutcome{assets: assets, err: err}
		}(index, chunk)
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return AssetResult{}, fmt.Errorf("%w: %w", ErrUserCancelled, err)
	}

	result := AssetResult{ChunkCount: len(chunks)}
	merged := make([]Asset, 0)
	seen := map[string]int{}
	for index, outcome := range results {
		chapterStart := index*AssetChunkSize + 1
		chapterEnd := chapterStart + AssetChunkSize - 1
		batchRecord := AssetBatch{
			ID:            newScreenplayID("abatch"),
			ProjectID:     req.ProjectID,
			OwnerUsername: req.Owner,
			Type:          req.Type,
			Degree:        req.Degree,
			BatchIndex:    index + 1,
			ChapterStart:  chapterStart,
			ChapterEnd:    chapterEnd,
			Status:        "done",
			ItemCount:     len(outcome.assets),
			CreatedAt:     time.Now().UTC(),
		}
		if outcome.err != nil {
			batchRecord.Status = "failed"
			batchRecord.ErrorMessage = outcome.err.Error()
			result.FailedChunks++
			result.FailedBatches = append(result.FailedBatches, AssetBatchFailure{
				BatchIndex:   index + 1,
				ChapterStart: chapterStart,
				ChapterEnd:   chapterEnd,
				Error:        outcome.err.Error(),
			})
			result.Warnings = append(result.Warnings, fmt.Sprintf("第 %d 批提取失败：%v", index+1, outcome.err))
		} else if len(outcome.assets) == 0 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("第 %d 批未识别到%s", index+1, assetTypeLabel(req.Type)))
		}
		if err := s.Store.SaveAssetBatch(ctx, batchRecord); err != nil {
			return AssetResult{}, err
		}
		if outcome.err != nil {
			continue
		}
		for _, asset := range outcome.assets {
			key := asset.Type.String() + "\x00" + asset.Name
			if position, exists := seen[key]; exists {
				mergeAsset(&merged[position], asset)
				continue
			}
			seen[key] = len(merged)
			merged = append(merged, asset)
		}
	}
	if result.FailedChunks == len(chunks) && len(chunks) > 0 {
		return result, fmt.Errorf("%w: 全部 %d 批资产提取失败", ErrAIRequest, len(chunks))
	}
	for i := range merged {
		merged[i].ID = newScreenplayID("asset")
		merged[i].ProjectID = req.ProjectID
		merged[i].OwnerUsername = req.Owner
		merged[i].Degree = req.Degree
		merged[i].Type = req.Type
	}
	saved, err := s.Store.MergeAssets(ctx, req.Owner, req.ProjectID, merged)
	if err != nil {
		return AssetResult{}, err
	}
	result.Assets = saved
	result.Batch = AssetBatch{
		ProjectID: req.ProjectID, OwnerUsername: req.Owner, Type: req.Type, Degree: req.Degree,
		BatchIndex: 0, Status: "done", ItemCount: len(saved), CreatedAt: time.Now().UTC(),
	}
	if result.FailedChunks > 0 {
		// 已有批次成功落库，但整体不能算成功：把部分失败标记和失败批次号
		// 一并交给上层，由上层决定阻断还是逐批提示。
		result.PartialFailure = true
		numbers := make([]string, 0, len(result.FailedBatches))
		for _, failure := range result.FailedBatches {
			numbers = append(numbers, fmt.Sprintf("%d", failure.BatchIndex))
		}
		return result, fmt.Errorf("%w：第 %s 批资产提取失败，已保存 %d 条", ErrPartialFailure, strings.Join(numbers, "、"), len(saved))
	}
	return result, nil
}

func (s *AssetService) prompts() *luoshui.Registry {
	if s.Prompts != nil {
		return s.Prompts
	}
	return luoshui.MustDefault()
}

type chunkOutcome struct {
	assets []Asset
	err    error
}

func (s *AssetService) extractChunk(ctx context.Context, req AssetRequest, promptID, systemPrompt, chunk string, batchIndex int) ([]Asset, error) {
	response, err := s.Generator.Generate(ctx, TextRequest{
		Model:        defaultString(req.Model, "gpt-4o-mini"),
		SystemPrompt: systemPrompt,
		UserPrompt:   chunk,
		JSONResponse: true,
	})
	if err != nil {
		return nil, wrapAIError(err)
	}
	assets, err := parseAssetJSON(req.Type, strings.TrimSpace(response.Content))
	if err != nil {
		return nil, err
	}
	for i := range assets {
		assets[i].SourceBatch = batchIndex
	}
	return assets, nil
}

func (s *AssetService) resolveChunks(ctx context.Context, req AssetRequest) ([]string, error) {
	if len(req.Chapters) > 0 {
		chunks := make([]string, 0, len(req.Chapters))
		for _, chapter := range req.Chapters {
			trimmed := strings.TrimSpace(chapter)
			if trimmed == "" {
				return nil, ErrEmptyContent
			}
			chunks = append(chunks, trimmed)
		}
		return chunks, nil
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		doc, err := s.Store.ActiveDocument(ctx, req.Owner, req.ProjectID)
		if err != nil {
			return nil, err
		}
		content = strings.TrimSpace(doc.Content)
	}
	if content == "" {
		return nil, ErrEmptyContent
	}
	return SplitIntoBatches(content, AssetChunkSize), nil
}

var (
	chapterHeadingRe      = regexp.MustCompile(`(?m)^[ \t]*第[0-9一二三四五六七八九十百千万两]+[章回][ \t]*[^\n]*$`)
	batchEpisodeHeadingRe = regexp.MustCompile(`(?m)^[ \t]*#[ \t]*第[0-9]+[ \t]*集[ \t]*[:：]?[^\n]*$`)
)

// SplitIntoBatches 优先按「第N章」切分，其次按「# 第N集」，都没有则整段一批。
// 返回的每一块包含 size 个连续章节/集。
func SplitIntoBatches(content string, size int) []string {
	if size <= 0 {
		size = AssetChunkSize
	}
	for _, pattern := range []*regexp.Regexp{chapterHeadingRe, batchEpisodeHeadingRe} {
		positions := pattern.FindAllStringIndex(content, -1)
		if len(positions) == 0 {
			continue
		}
		units := make([]string, 0, len(positions))
		for i, position := range positions {
			end := len(content)
			if i+1 < len(positions) {
				end = positions[i+1][0]
			}
			units = append(units, strings.TrimSpace(content[position[0]:end]))
		}
		chunks := make([]string, 0, (len(units)+size-1)/size)
		for start := 0; start < len(units); start += size {
			end := start + size
			if end > len(units) {
				end = len(units)
			}
			chunks = append(chunks, strings.Join(units[start:end], "\n\n"))
		}
		return chunks
	}
	return []string{strings.TrimSpace(content)}
}

// assetPromptID 是四档 × 三类型的派发表。
//
// TODO(需运行确认): 规格 §1.3 只给出 8 套模板偏移，未逐档标注 more_diff 与
// standard/master_real 的场景/道具对应关系。当前按「真人/3D 用 V4，标准/差异化用 V3」
// 落表，差异需在对照测试中校正。
func assetPromptID(kind AssetType, degree AssetDegree) (string, error) {
	switch kind {
	case AssetCharacter:
		switch degree {
		case DegreeStandard:
			return "asset_character_standard_v5", nil
		case DegreeMoreDiff:
			return "asset_character_lite", nil
		case DegreeMaster:
			return "asset_character_v6_3d", nil
		case DegreeMasterReal:
			return "asset_character_real_v7", nil
		}
	case AssetScene:
		if degree == DegreeStandard || degree == DegreeMoreDiff {
			return "asset_scene_v3", nil
		}
		return "asset_scene_v4", nil
	case AssetProp:
		if degree == DegreeStandard || degree == DegreeMoreDiff {
			return "asset_prop_v3", nil
		}
		return "asset_prop_v4", nil
	}
	return "", fmt.Errorf("%w: 未知资产类型 %q", ErrInvalidArgument, kind)
}

func validAssetType(kind AssetType) bool {
	return kind == AssetCharacter || kind == AssetScene || kind == AssetProp
}

func validDegree(degree AssetDegree) bool {
	switch degree {
	case DegreeStandard, DegreeMoreDiff, DegreeMaster, DegreeMasterReal:
		return true
	}
	return false
}

func assetTypeLabel(kind AssetType) string {
	switch kind {
	case AssetCharacter:
		return "人物"
	case AssetScene:
		return "场景"
	case AssetProp:
		return "道具"
	}
	return string(kind)
}

// parseAssetJSON 只接受提示词声明的顶层键，容忍代码块包裹与前后说明文字。
func parseAssetJSON(kind AssetType, raw string) ([]Asset, error) {
	if raw == "" {
		return nil, ErrAIEmptyContent
	}
	object, ok := extractJSONObject(raw)
	if !ok {
		return nil, ErrAIInvalidJSON
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal([]byte(object), &envelope); err != nil {
		return nil, fmt.Errorf("%w %v", ErrAIJSONParse, err)
	}
	key := "characters"
	switch kind {
	case AssetScene:
		key = "scenes"
	case AssetProp:
		key = "props"
	}
	itemsRaw, ok := envelope[key]
	if !ok {
		// 提示词允许道具为空数组；其它类型缺键视为无效。
		if kind == AssetProp {
			return []Asset{}, nil
		}
		return nil, ErrAIInvalidJSON
	}
	var items []map[string]any
	if err := json.Unmarshal(itemsRaw, &items); err != nil {
		return nil, fmt.Errorf("%w %v", ErrAIJSONParse, err)
	}
	assets := make([]Asset, 0, len(items))
	for _, item := range items {
		asset, ok := assetFromMap(kind, item)
		if ok {
			assets = append(assets, asset)
		}
	}
	return assets, nil
}

func assetFromMap(kind AssetType, item map[string]any) (Asset, bool) {
	name := stringField(item, "name")
	if name == "" {
		return Asset{}, false
	}
	extra := map[string]string{}
	for key, value := range item {
		text, ok := value.(string)
		if !ok {
			continue
		}
		switch key {
		case "name", "identity", "description", "prompt", "voice", "颜值定位", "出现集数", "空间要点", "用途", "视觉要点", "tag":
		default:
			if strings.TrimSpace(text) != "" {
				extra[key] = text
			}
		}
	}
	asset := Asset{
		Type:     kind,
		Name:     name,
		Prompt:   stringField(item, "prompt"),
		Episodes: stringField(item, "出现集数"),
		Tag:      stringField(item, "tag"),
		Extra:    extra,
	}
	switch kind {
	case AssetCharacter:
		asset.Identity = stringField(item, "identity")
		asset.Description = stringField(item, "颜值定位")
		asset.Voice = stringField(item, "voice")
	case AssetScene:
		asset.Identity = stringField(item, "空间要点")
		asset.Description = stringField(item, "description")
	case AssetProp:
		asset.Identity = stringField(item, "用途")
		asset.Description = stringField(item, "description")
		if value := stringField(item, "视觉要点"); value != "" {
			if asset.Extra == nil {
				asset.Extra = map[string]string{}
			}
			asset.Extra["视觉要点"] = value
		}
	}
	return asset, true
}

func stringField(item map[string]any, key string) string {
	value, ok := item[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

// mergeAsset 同名合并：只补空字段，已有内容不被低质量结果覆盖。
func mergeAsset(target *Asset, incoming Asset) {
	if target.Prompt == "" || len([]rune(incoming.Prompt)) > len([]rune(target.Prompt)) {
		target.Prompt = incoming.Prompt
	}
	if target.Identity == "" {
		target.Identity = incoming.Identity
	}
	if target.Description == "" {
		target.Description = incoming.Description
	}
	if target.Voice == "" {
		target.Voice = incoming.Voice
	}
	if target.Episodes == "" {
		target.Episodes = incoming.Episodes
	}
	if target.Tag == "" {
		target.Tag = incoming.Tag
	}
	if target.Extra == nil && incoming.Extra != nil {
		target.Extra = map[string]string{}
	}
	for key, value := range incoming.Extra {
		if target.Extra[key] == "" {
			target.Extra[key] = value
		}
	}
}

func extractJSONObject(raw string) (string, bool) {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return "", false
	}
	return raw[start : end+1], true
}

func (t AssetType) String() string { return string(t) }
