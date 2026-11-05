package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"qiantie/backend/internal/prompt"
)

// BuildComicBootstrapContract combines the selected asset and page rules for
// the first whole-book run. The model returns assets, pagination and every
// page prompt together, while the service still persists them independently.
func BuildComicBootstrapContract(book Book, snapshot DirectorSnapshot) (PromptContract, error) {
	if strings.TrimSpace(book.SourceText) == "" {
		return PromptContract{}, fmt.Errorf("%w: source text is required", ErrInvalid)
	}
	if normalizeProductionMode(snapshot.Effective) != "comic" {
		return PromptContract{}, NewConfigRequiredError("请先打开漫画生成")
	}
	config := aiReasoningPromptConfig(snapshot.Effective).Comic
	if strings.TrimSpace(config.Assets.Body) == "" || strings.TrimSpace(config.Body) == "" {
		return PromptContract{}, NewConfigRequiredError("请先选择漫画资产提示词和漫画分镜提示词")
	}

	sections := []string{
		prompt.ComicBootstrapProtocol,
		"【当前漫画资产推导规则】\n" + strings.TrimSpace(config.Assets.Body),
		"【当前漫画分页与页面提示词规则】\n" + expandRecoveredComicPreset(strings.TrimSpace(config.Body), book, strings.TrimSpace(book.SourceText), !AssetImagesEnabled(snapshot.Effective)),
		comicPromptTextForAspect(comicPageStructureLock, EffectiveImageAspectRatio(snapshot.Effective)),
	}
	if body := strings.TrimSpace(config.ImageDerivation.Body); body != "" {
		sections = append(sections, "【图片描述词推导规则】\n"+body)
	}
	if body := strings.TrimSpace(config.Style.Body); body != "" {
		sections = append(sections, "【所选画风】\n"+body)
	}
	if body := strings.TrimSpace(config.ImagePromptPrefix); body != "" {
		sections = append(sections, "【图片描述词全局前缀】\n"+body)
	}
	for _, constraint := range config.Constraints {
		if body := strings.TrimSpace(constraint.Body); body != "" {
			sections = append(sections, body)
		}
	}
	boundaries := EnumerateComicSourceBoundaries(book.SourceText)
	fallbackSegments, err := FallbackComicSourceSegments(book.SourceText, boundaries)
	if err != nil {
		return PromptContract{}, err
	}
	totalCharacters := comicSourceCharacterCount(book.SourceText)
	minimumPageCount := (totalCharacters + 149) / 150
	if minimumPageCount < 1 {
		minimumPageCount = 1
	}
	payload, err := json.Marshal(map[string]any{
		"book_id":                book.ID,
		"book_title":             book.Title,
		"source_text":            book.SourceText,
		"source_safe_boundaries": comicSegmentationBoundaryPrompt(book.SourceText, boundaries),
		"page_plan": map[string]any{
			"minimum_page_count":     minimumPageCount,
			"recommended_page_count": len(fallbackSegments),
			"target_characters":      "90-120",
			"upper_limit_characters": 150,
		},
		"aspect_ratio": EffectiveImageAspectRatio(snapshot.Effective),
	})
	if err != nil {
		return PromptContract{}, err
	}
	return PromptContract{SystemPrompt: strings.Join(sections, "\n\n"), UserPrompt: string(payload), Temperature: 0.2, MaxTokens: 20000}, nil
}

type ComicBootstrapResult struct {
	Assets   []BookAsset
	Revision DirectorRevision
	Videos   []Video
}

// comicAssetStageCanReuse reports whether the missing-only comic asset action
// already has an automatic comic asset set to keep. A recorded comic revision
// must still point at the current source; old records without a revision stay
// reusable for backwards compatibility.
func comicAssetStageCanReuse(book Book) bool {
	for _, asset := range book.AssetRecords {
		if strings.TrimSpace(asset.Source) != "comic" || strings.TrimSpace(asset.Prompt) == "" {
			continue
		}
		if book.DirectorRevision != nil && strings.TrimSpace(book.DirectorRevision.SourceDigest) != "" && book.DirectorRevision.SourceDigest != sourceDigest(book.SourceText) {
			return false
		}
		return true
	}
	return false
}

// comicBootstrapCanReuse keeps the normal comic extraction action idempotent.
// A complete, current analysis is reused; a missing page, missing segment or
// changed source falls through to the normal bootstrap request.
func comicBootstrapCanReuse(book Book, requireAssets bool) bool {
	if (requireAssets && !comicAssetStageCanReuse(book)) || len(book.Videos) == 0 {
		return false
	}
	for _, video := range book.Videos {
		if strings.TrimSpace(video.ComicPrompt) == "" || strings.TrimSpace(video.ComicImagePrompt) == "" || video.ComicSourceSegment == nil || strings.TrimSpace(video.ComicSourceSegment.SourceText) == "" || video.ComicSourceSegment.SourceDigest != sourceDigest(book.SourceText) {
			return false
		}
	}
	return true
}

type comicBootstrapPage struct {
	EndBoundaryID string            `json:"end_boundary_id"`
	RoleStage     map[string]string `json:"role_stage"`
	ScreenPrompt  string            `json:"screen_prompt"`
	ImagePrompt   string            `json:"image_prompt"`
	Roles         []string          `json:"roles"`
	ScenePresets  []string          `json:"scene_presets"`
}

func parseComicBootstrapOutput(raw json.RawMessage, snapshot DirectorSnapshot, source string, boundaries []ComicSourceBoundary) (DirectorResult, []comicPromptResult, []ComicSourceSegmentDraft, error) {
	normalizedRaw, err := normalizeComicBootstrapTiming(raw)
	if err != nil {
		return DirectorResult{}, nil, nil, retryableModelOutputError(err)
	}
	director, err := NormalizeDirectorOutput(normalizedRaw, DirectorSettings{
		MaxVideoDuration: 1,
		AspectRatio:      EffectiveImageAspectRatio(snapshot.Effective),
	})
	if err != nil {
		return DirectorResult{}, nil, nil, retryableModelOutputError(err)
	}
	var envelope struct {
		Storyboard []comicBootstrapPage `json:"storyboard"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil || len(envelope.Storyboard) != len(director.Storyboard) {
		return DirectorResult{}, nil, nil, retryableModelOutputError(fmt.Errorf("漫画分页提示词数量不一致"))
	}
	boundaryIDs := make([]string, 0, len(envelope.Storyboard))
	light := aiReasoningPromptConfig(snapshot.Effective).Comic.ID == "batch-comic-page-light"
	pages := make([]comicPromptResult, 0, len(envelope.Storyboard))
	for index, page := range envelope.Storyboard {
		page.ScreenPrompt = strings.TrimSpace(page.ScreenPrompt)
		page.ImagePrompt = strings.TrimSpace(page.ImagePrompt)
		if page.ScreenPrompt == "" || (!light && page.ImagePrompt == "") {
			return DirectorResult{}, nil, nil, retryableModelOutputError(fmt.Errorf("storyboard[%d] 漫画提示词为空", index))
		}
		boundaryIDs = append(boundaryIDs, strings.TrimSpace(page.EndBoundaryID))
		pages = append(pages, comicPromptResult{
			RoleStage: normalizedStringMap(page.RoleStage), ScreenPrompt: page.ScreenPrompt,
			ImagePrompt: page.ImagePrompt, Roles: normalizedStringList(page.Roles),
			ScenePresets: normalizedStringList(page.ScenePresets),
		})
	}
	segments, err := comicSegmentsFromBoundaryIDs(source, boundaries, boundaryIDs, "bootstrap_content_boundary")
	if err != nil {
		fallback, fallbackErr := FallbackComicSourceSegments(source, boundaries)
		lastBoundaryID := fmt.Sprintf("b%d", len(boundaries))
		if boundaryIDs[len(boundaryIDs)-1] != lastBoundaryID {
			return DirectorResult{}, nil, nil, retryableModelOutputError(fmt.Errorf("漫画正文没有完整覆盖: %w", err))
		}
		if fallbackErr == nil && len(fallback) == len(pages) {
			segments = fallback
		} else if segments, fallbackErr = comicSourceSegmentsForPageCount(source, boundaries, len(pages)); fallbackErr != nil {
			return DirectorResult{}, nil, nil, retryableModelOutputError(fmt.Errorf("漫画正文没有完整覆盖（模型返回 %d 页，安全纠偏失败：%v）: %w", len(pages), fallbackErr, err))
		}
	}
	return director, pages, segments, nil
}

func normalizeComicBootstrapTiming(raw json.RawMessage) (json.RawMessage, error) {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, err
	}
	storyboard, ok := root["storyboard"].([]any)
	if !ok {
		return raw, nil
	}
	for _, item := range storyboard {
		page, ok := item.(map[string]any)
		if !ok {
			continue
		}
		description := firstDirectorText(page, "video_desc", "videoDesc", "screen_prompt", "screenPrompt", "image_prompt", "imagePrompt")
		page["duration_sec"] = 1
		page["prefix_key"] = "general_anime"
		page["video_desc"] = description
		page["shots"] = []any{map[string]any{
			"start_sec": 0, "end_sec": 1, "shot_type": "漫画页", "camera": "漫画页", "description": description,
		}}
	}
	return json.Marshal(root)
}

func (s *DirectorService) RunComicBootstrap(ctx context.Context, owner, batchID, bookID string) (ComicBootstrapResult, error) {
	if err := s.validate(); err != nil {
		return ComicBootstrapResult{}, err
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return ComicBootstrapResult{}, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return ComicBootstrapResult{}, err
	}
	book, err = s.comicProductionBook(ctx, owner, batchID, book)
	if err != nil {
		return ComicBootstrapResult{}, err
	}
	effective := ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	assetImages := AssetImagesEnabled(effective)
	if assetImages && !comicAssetStageCanReuse(book) {
		return ComicBootstrapResult{}, fmt.Errorf("%w: 请先提取漫画资产", ErrConflict)
	}
	if comicBootstrapCanReuse(book, assetImages) {
		revision := DirectorRevision{}
		if book.DirectorRevision != nil {
			revision = *book.DirectorRevision
		}
		return ComicBootstrapResult{Assets: book.AssetRecords, Revision: revision, Videos: book.Videos}, nil
	}
	segmentStore, ok := s.Store.(ComicSourceSegmentStore)
	if !ok {
		return ComicBootstrapResult{}, fmt.Errorf("%w: 漫画原文分段存储未配置", ErrUnavailable)
	}
	segmentation, segmentationErr := segmentStore.GetComicSourceSegmentation(ctx, owner, batchID, bookID)
	if errors.Is(segmentationErr, ErrNotFound) || (segmentationErr == nil && segmentation.Stale) {
		segmentation, segmentationErr = s.RunComicSourceSegmentation(ctx, owner, batchID, bookID)
	}
	if segmentationErr != nil {
		return ComicBootstrapResult{}, segmentationErr
	}
	if len(segmentation.Segments) == 0 {
		return ComicBootstrapResult{}, fmt.Errorf("%w: 漫画分页为空", ErrConflict)
	}
	if _, err := s.RunMissingComicPromptExtraction(ctx, owner, batchID, bookID); err != nil {
		return ComicBootstrapResult{}, err
	}
	latestBatch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return ComicBootstrapResult{}, err
	}
	latestBook, err := bookFromBatch(latestBatch, bookID)
	if err != nil {
		return ComicBootstrapResult{}, err
	}
	latestBook, err = s.comicProductionBook(ctx, owner, batchID, latestBook)
	if err != nil {
		return ComicBootstrapResult{}, err
	}
	revision := DirectorRevision{}
	if latestBook.DirectorRevision != nil {
		revision = *latestBook.DirectorRevision
	}
	return ComicBootstrapResult{Assets: latestBook.AssetRecords, Revision: revision, Videos: latestBook.Videos}, nil
}
