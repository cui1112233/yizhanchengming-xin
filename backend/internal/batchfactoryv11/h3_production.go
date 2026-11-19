package batchfactoryv11

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

func (s *PromptCompilerService) CompileH3ForProduction(ctx context.Context, owner, batchID, bookID, videoID, compilationID string, referenceLimit int) (FinalPrompt, error) {
	if s == nil || s.Store == nil {
		return FinalPrompt{}, ErrUnavailable
	}
	repository, ok := s.Store.(H3Repository)
	if !ok {
		return FinalPrompt{}, ErrUnavailable
	}
	revision, err := repository.GetH3VideoCompilation(ctx, owner, compilationID)
	if err != nil {
		return FinalPrompt{}, err
	}
	if revision.BatchID != batchID || revision.BookID != bookID {
		return FinalPrompt{}, ErrNotFound
	}
	effective, book, video, ordinal, err := s.ResolveEffective(ctx, owner, batchID, bookID, videoID)
	if err != nil {
		return FinalPrompt{}, err
	}
	if book.DirectorRevision == nil || book.DirectorRevision.ID != revision.Compilation.DirectorRevisionID || ordinal >= len(revision.Compilation.Segments) {
		return FinalPrompt{}, fmt.Errorf("%w: VIDEO does not belong to the frozen H3 compilation", ErrConflict)
	}
	segment := revision.Compilation.Segments[ordinal]
	references, downgraded, err := s.h3ReferenceImages(ctx, owner, batchID, bookID, video, book, effective.Values, referenceLimit)
	if err != nil {
		return FinalPrompt{}, err
	}
	durationSeconds := int((segment.RequestDurationMS + 999) / 1000)
	trace := segment.CompileTrace
	compiledPrompt := segment.CompiledPrompt
	displayPrompt := segment.EditableCopy
	snapshotHash := segment.CompiledPromptHash
	// 用户手动编辑或自动去敏改写的视频提示词优先于冻结编译内容；否则重试
	// 会再次提交被审核拒绝的原文，改写等于没生效。
	if override, ok := videoPromptOverrideFor(video, 0); ok {
		compiledPrompt = override
		displayPrompt = override
		snapshotHash = fmt.Sprintf("%x", sha256.Sum256([]byte(strings.TrimSpace(override))))
		trace.EditableCopySource = "video_prompt_override"
	}
	return FinalPrompt{
		BatchID:               batchID,
		BookID:                bookID,
		VideoID:               videoID,
		DirectorRevisionID:    revision.Compilation.DirectorRevisionID,
		SnapshotHash:          snapshotHash,
		DisplayPrompt:         displayPrompt,
		CompiledPrompt:        compiledPrompt,
		Components:            []PromptComponent{{Key: "h3", Label: "", Content: compiledPrompt}},
		EffectiveSettings:     effective,
		DurationSeconds:       durationSeconds,
		ReferenceImageURLs:    references,
		DowngradedAssetIDs:    downgraded,
		CompilationID:         revision.ID,
		CompilationSegmentKey: segment.SegmentKey,
		CompileTrace:          &trace,
	}, nil
}

func (s *PromptCompilerService) h3ReferenceImages(ctx context.Context, owner, batchID, bookID string, video Video, book Book, values SettingsPatch, frozenLimit int) ([]string, []string, error) {
	limit := frozenLimit
	if limit <= 0 {
		limit = referenceImageLimit(values)
	}
	references := []string{}
	for _, value := range append(rawStrings(values, "imageUrls"), rawStrings(values, "referenceImages")...) {
		if imageURL := providerAccessibleImageURL(value); imageURL != "" && len(references) < limit {
			references = append(references, imageURL)
		}
	}
	records := recordsByKindAndName(book, map[string][]NamedPrompt{"character": book.Assets.Characters, "scene": book.Assets.Scenes, "prop": book.Assets.Props})
	images, err := s.primaryAssetImages(ctx, owner, batchID, bookID, records)
	if err != nil {
		return nil, nil, err
	}
	selection, hasSelection := decodeVideoAssetSelection(video.SettingsState.Patch)
	selectedIDs := selection.effectiveAssetIDSet()
	downgraded := []string{}
	recordKeys := make([]string, 0, len(records))
	for key := range records {
		recordKeys = append(recordKeys, key)
	}
	sort.Strings(recordKeys)
	for _, key := range recordKeys {
		record := records[key]
		if record.ID == "" || (hasSelection && !selectedIDs[record.ID]) {
			continue
		}
		imageURL := images[record.ID]
		if imageURL == "" {
			continue
		}
		if len(references) < limit {
			references = append(references, imageURL)
		} else {
			downgraded = append(downgraded, record.ID)
		}
	}
	return references, downgraded, nil
}

func cloneH3CompileTrace(value *H3CompileTrace) *H3CompileTrace {
	if value == nil {
		return nil
	}
	clone := *value
	clone.InjectedLayers = append([]string(nil), value.InjectedLayers...)
	clone.OmittedLayers = append([]string(nil), value.OmittedLayers...)
	clone.SourceSlices = append([]H3SourceSlice(nil), value.SourceSlices...)
	return &clone
}
