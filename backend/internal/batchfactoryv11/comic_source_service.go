package batchfactoryv11

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"qiantie/backend/internal/prompt"
)

type comicSegmentationSelection struct {
	BoundaryIDs []string `json:"boundaryIds"`
}

func (s *DirectorService) RunComicSourceSegmentation(ctx context.Context, owner, batchID, bookID string) (ComicSourceSegmentation, error) {
	if s == nil || s.Store == nil {
		return ComicSourceSegmentation{}, ErrUnavailable
	}
	segmentStore, ok := s.Store.(ComicSourceSegmentStore)
	if !ok {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 漫画原文分段存储未配置", ErrUnavailable)
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return ComicSourceSegmentation{}, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return ComicSourceSegmentation{}, err
	}
	if draft, draftErr := s.Store.GetDraft(ctx, owner, "working-front:"+book.ID, "working-front-content", batchID); draftErr == nil && strings.TrimSpace(draft.Content) != "" {
		book.SourceText = draft.Content
	} else if draftErr != nil && draftErr != ErrNotFound {
		return ComicSourceSegmentation{}, draftErr
	}
	if book.SourceText == "" {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 漫画正文为空", ErrInvalid)
	}
	boundaries := EnumerateComicSourceBoundaries(book.SourceText)
	drafts, err := FallbackComicSourceSegments(book.SourceText, boundaries)
	if err != nil {
		return ComicSourceSegmentation{}, err
	}
	if s.Provider != nil {
		request := TextCompletionRequest{SystemPrompt: prompt.ComicSegmentationSystemPrompt, UserPrompt: comicSegmentationBoundaryPrompt(book.SourceText, boundaries), Temperature: 0.2, MaxTokens: 4096}
		if completion, completeErr := s.Provider.Complete(ctx, request); completeErr == nil {
			if selected, selectErr := comicSegmentsFromModelSelection(book.SourceText, boundaries, completion); selectErr == nil {
				drafts = selected
			}
		}
	}
	drafts, err = LimitComicSourceSegments(book.SourceText, drafts, ComicMaxPageCount)
	if err != nil {
		return ComicSourceSegmentation{}, err
	}
	return segmentStore.ReplaceComicSourceSegments(ctx, owner, batchID, bookID, sourceDigest(book.SourceText), drafts)
}

func comicSegmentationBoundaryPrompt(source string, boundaries []ComicSourceBoundary) string {
	var output strings.Builder
	output.WriteString("按顺序选择漫画页结束边界：\n")
	start := 0
	for index, boundary := range boundaries {
		if boundary.Offset <= start || boundary.Offset > len(source) {
			continue
		}
		fmt.Fprintf(&output, "[b%d | %s | 累计%d字] %s\n", index+1, boundary.Type, comicSourceCharacterCount(source[start:boundary.Offset]), source[start:boundary.Offset])
		start = boundary.Offset
	}
	return output.String()
}

func comicSegmentsFromModelSelection(source string, boundaries []ComicSourceBoundary, completion string) ([]ComicSourceSegmentDraft, error) {
	raw, err := ParseDirectorJSON(completion)
	if err != nil {
		return nil, err
	}
	var selection comicSegmentationSelection
	if err := json.Unmarshal(raw, &selection); err != nil || len(selection.BoundaryIDs) == 0 {
		return nil, fmt.Errorf("invalid boundary selection")
	}
	return comicSegmentsFromBoundaryIDs(source, boundaries, selection.BoundaryIDs, "model_selected_boundary")
}

func comicSegmentsFromBoundaryIDs(source string, boundaries []ComicSourceBoundary, boundaryIDs []string, reason string) ([]ComicSourceSegmentDraft, error) {
	if len(boundaryIDs) == 0 {
		return nil, fmt.Errorf("boundary selection is empty")
	}
	byID := make(map[string]ComicSourceBoundary, len(boundaries))
	for index, boundary := range boundaries {
		byID[fmt.Sprintf("b%d", index+1)] = boundary
	}
	drafts := make([]ComicSourceSegmentDraft, 0, len(boundaryIDs))
	start := 0
	for index, id := range boundaryIDs {
		boundary, ok := byID[id]
		if !ok || boundary.Offset <= start {
			return nil, fmt.Errorf("unsafe or reordered boundary %q", id)
		}
		text := source[start:boundary.Offset]
		count := comicSourceCharacterCount(text)
		if index < len(boundaryIDs)-1 && count < comicPageTextMin {
			return nil, fmt.Errorf("page below target range")
		}
		if count > comicPageRepairMax && hasComicBoundaryBetween(boundaries, start, boundary.Offset) {
			return nil, fmt.Errorf("page exceeds upper limit despite safe boundary")
		}
		drafts = append(drafts, ComicSourceSegmentDraft{StartOffset: start, EndOffset: boundary.Offset, SourceText: text, CharacterCount: count, BoundaryType: boundary.Type, Reason: reason})
		start = boundary.Offset
	}
	if start != len(source) {
		return nil, fmt.Errorf("model selection does not cover source")
	}
	if err := ValidateComicSourceSegments(source, drafts); err != nil {
		return nil, err
	}
	return drafts, nil
}

func hasComicBoundaryBetween(boundaries []ComicSourceBoundary, start, end int) bool {
	for _, boundary := range boundaries {
		if boundary.Offset > start && boundary.Offset < end {
			return true
		}
	}
	return false
}
