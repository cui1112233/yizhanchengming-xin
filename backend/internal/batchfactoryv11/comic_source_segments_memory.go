package batchfactoryv11

import (
	"context"
	"fmt"
	"sort"
	"time"
)

func comicSourceSegmentWithBoundaries(segment ComicSourceSegment) ComicSourceSegment {
	local := EnumerateComicSourceBoundaries(segment.SourceText)
	segment.SafeBoundaries = make([]ComicSourceBoundary, 0, len(local))
	for _, boundary := range local {
		absolute := segment.StartOffset + boundary.Offset
		if absolute <= segment.StartOffset || absolute >= segment.EndOffset {
			continue
		}
		preview := segment.SourceText[:boundary.Offset]
		previewRunes := []rune(preview)
		if len(previewRunes) > 18 {
			previewRunes = previewRunes[len(previewRunes)-18:]
		}
		segment.SafeBoundaries = append(segment.SafeBoundaries, ComicSourceBoundary{Offset: absolute, Type: boundary.Type, Preview: string(previewRunes)})
	}
	return segment
}

func (s *MemoryStore) ReplaceComicSourceSegments(_ context.Context, owner, batchID, bookID, digest string, drafts []ComicSourceSegmentDraft) (ComicSourceSegmentation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.batches[batchID]
	if !ok || owned.Owner != owner {
		return ComicSourceSegmentation{}, ErrNotFound
	}
	bookIndex := -1
	for index := range owned.Value.Books {
		if owned.Value.Books[index].ID == bookID {
			bookIndex = index
			break
		}
	}
	if bookIndex < 0 {
		return ComicSourceSegmentation{}, ErrNotFound
	}
	book := &owned.Value.Books[bookIndex]
	source := book.SourceText
	if draft, exists := s.drafts[draftKey(owner, "working-front:"+book.ID, "working-front-content", batchID)]; exists && draft.Content != "" {
		source = draft.Content
	}
	if digest != sourceDigest(source) {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 漫画正文已变化", ErrConflict)
	}
	if err := ValidateComicSourceSegments(source, drafts); err != nil {
		return ComicSourceSegmentation{}, err
	}
	version := int64(1)
	for _, stored := range s.comicSourceSegments {
		if stored.Owner == owner && stored.Value.BatchID == batchID && stored.Value.BookID == bookID && stored.Value.SegmentationVersion >= version {
			version = stored.Value.SegmentationVersion + 1
		}
	}
	for index := range book.Videos {
		book.Videos[index].CompatibilityState = "orphaned"
	}
	now := time.Now().UTC()
	videos := make([]Video, 0, len(drafts))
	for ordinal, draft := range drafts {
		videoID := s.id("video")
		video := Video{ID: videoID, BatchID: batchID, BookID: bookID, Label: fmt.Sprintf("漫画页 %02d", ordinal+1), CompatibilityState: "active", Revision: 1}
		segment := ComicSourceSegment{VideoID: videoID, BatchID: batchID, BookID: bookID, Ordinal: ordinal, StartOffset: draft.StartOffset, EndOffset: draft.EndOffset, SourceText: draft.SourceText, CharacterCount: draft.CharacterCount, BoundaryType: draft.BoundaryType, Reason: draft.Reason, SourceDigest: digest, SegmentationVersion: version, Revision: 1}
		s.comicSourceSegments[videoID] = memoryOwned[ComicSourceSegment]{Owner: owner, Value: segment}
		videos = append(videos, video)
	}
	book.Videos = videos
	book.Revision++
	owned.Value.UpdatedAt = now
	s.batches[batchID] = owned
	return s.getComicSourceSegmentationLocked(owner, batchID, bookID)
}

func (s *MemoryStore) BindComicSourceSegments(_ context.Context, owner, batchID, bookID, digest string, videoIDs []string, drafts []ComicSourceSegmentDraft) (ComicSourceSegmentation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.batches[batchID]
	if !ok || owned.Owner != owner {
		return ComicSourceSegmentation{}, ErrNotFound
	}
	bookIndex := -1
	for index := range owned.Value.Books {
		if owned.Value.Books[index].ID == bookID {
			bookIndex = index
			break
		}
	}
	if bookIndex < 0 {
		return ComicSourceSegmentation{}, ErrNotFound
	}
	book := &owned.Value.Books[bookIndex]
	source := book.SourceText
	if draft, exists := s.drafts[draftKey(owner, "working-front:"+book.ID, "working-front-content", batchID)]; exists && draft.Content != "" {
		source = draft.Content
	}
	if digest != sourceDigest(source) {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 漫画正文已变化", ErrConflict)
	}
	if len(videoIDs) != len(drafts) || len(videoIDs) == 0 {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 漫画页与原文分段数量不一致", ErrInvalid)
	}
	if err := ValidateComicSourceSegments(source, drafts); err != nil {
		return ComicSourceSegmentation{}, err
	}
	activeIDs := make([]string, 0, len(book.Videos))
	for _, video := range book.Videos {
		if video.CompatibilityState == "active" {
			activeIDs = append(activeIDs, video.ID)
		}
	}
	if len(activeIDs) != len(videoIDs) {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 漫画页已变化", ErrConflict)
	}
	for index := range activeIDs {
		if activeIDs[index] != videoIDs[index] {
			return ComicSourceSegmentation{}, fmt.Errorf("%w: 漫画页顺序已变化", ErrConflict)
		}
	}
	version := int64(1)
	for _, stored := range s.comicSourceSegments {
		if stored.Owner == owner && stored.Value.BatchID == batchID && stored.Value.BookID == bookID && stored.Value.SegmentationVersion >= version {
			version = stored.Value.SegmentationVersion + 1
		}
	}
	for ordinal, draft := range drafts {
		segment := ComicSourceSegment{VideoID: videoIDs[ordinal], BatchID: batchID, BookID: bookID, Ordinal: ordinal, StartOffset: draft.StartOffset, EndOffset: draft.EndOffset, SourceText: draft.SourceText, CharacterCount: draft.CharacterCount, BoundaryType: draft.BoundaryType, Reason: draft.Reason, SourceDigest: digest, SegmentationVersion: version, Revision: 1}
		s.comicSourceSegments[segment.VideoID] = memoryOwned[ComicSourceSegment]{Owner: owner, Value: segment}
	}
	book.Revision++
	owned.Value.UpdatedAt = time.Now().UTC()
	s.batches[batchID] = owned
	return s.getComicSourceSegmentationLocked(owner, batchID, bookID)
}

func (s *MemoryStore) GetComicSourceSegmentation(_ context.Context, owner, batchID, bookID string) (ComicSourceSegmentation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getComicSourceSegmentationLocked(owner, batchID, bookID)
}

func (s *MemoryStore) getComicSourceSegmentationLocked(owner, batchID, bookID string) (ComicSourceSegmentation, error) {
	owned, ok := s.batches[batchID]
	if !ok || owned.Owner != owner {
		return ComicSourceSegmentation{}, ErrNotFound
	}
	book, err := bookFromBatch(owned.Value, bookID)
	if err != nil {
		return ComicSourceSegmentation{}, err
	}
	source := book.SourceText
	if draft, exists := s.drafts[draftKey(owner, "working-front:"+book.ID, "working-front-content", batchID)]; exists && draft.Content != "" {
		source = draft.Content
	}
	result := ComicSourceSegmentation{BatchID: batchID, BookID: bookID, CurrentDigest: sourceDigest(source)}
	for ordinal, video := range book.Videos {
		if video.CompatibilityState != "active" {
			continue
		}
		stored, ok := s.comicSourceSegments[video.ID]
		if !ok || stored.Owner != owner {
			continue
		}
		segment := stored.Value
		segment.Ordinal = ordinal
		segment = comicSourceSegmentWithBoundaries(segment)
		result.Segments = append(result.Segments, segment)
		result.SourceDigest = segment.SourceDigest
		if segment.SegmentationVersion > result.Version {
			result.Version = segment.SegmentationVersion
		}
	}
	if len(result.Segments) == 0 {
		return ComicSourceSegmentation{}, ErrNotFound
	}
	result.Stale = result.SourceDigest != result.CurrentDigest
	return result, nil
}

func (s *MemoryStore) SplitComicSourceSegment(_ context.Context, owner, batchID, bookID, videoID string, input ComicSourceSplitInput) (ComicSourceSegmentation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.batches[batchID]
	if !ok || owned.Owner != owner {
		return ComicSourceSegmentation{}, ErrNotFound
	}
	bookIndex := -1
	for index := range owned.Value.Books {
		if owned.Value.Books[index].ID == bookID {
			bookIndex = index
			break
		}
	}
	if bookIndex < 0 {
		return ComicSourceSegmentation{}, ErrNotFound
	}
	book := &owned.Value.Books[bookIndex]
	source := book.SourceText
	if draft, exists := s.drafts[draftKey(owner, "working-front:"+book.ID, "working-front-content", batchID)]; exists && draft.Content != "" {
		source = draft.Content
	}
	videoIndex := -1
	for index := range book.Videos {
		if book.Videos[index].ID == videoID && book.Videos[index].CompatibilityState == "active" {
			videoIndex = index
			break
		}
	}
	stored, exists := s.comicSourceSegments[videoID]
	if videoIndex < 0 || !exists || stored.Owner != owner {
		return ComicSourceSegmentation{}, ErrNotFound
	}
	current := stored.Value
	if input.ExpectedRevision != current.Revision {
		return ComicSourceSegmentation{}, ErrConflict
	}
	if input.Direction != "left" && input.Direction != "right" {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 拆分方向无效", ErrInvalid)
	}
	validBoundary := false
	boundaryType := "sentence"
	for _, boundary := range comicSourceSegmentWithBoundaries(current).SafeBoundaries {
		if boundary.Offset == input.BoundaryOffset {
			validBoundary, boundaryType = true, boundary.Type
			break
		}
	}
	if !validBoundary {
		return ComicSourceSegmentation{}, fmt.Errorf("%w: 只能在完整句子或连续对白结束处拆分", ErrInvalid)
	}
	prefix := current.SourceText[:input.BoundaryOffset-current.StartOffset]
	suffix := current.SourceText[input.BoundaryOffset-current.StartOffset:]
	newID := s.id("video")
	newSegment := ComicSourceSegment{VideoID: newID, BatchID: batchID, BookID: bookID, SourceDigest: current.SourceDigest, SegmentationVersion: current.SegmentationVersion + 1, PromptStale: true, MediaStale: true, Revision: 1, Reason: "manual_split"}
	current.SegmentationVersion++
	current.PromptStale, current.MediaStale, current.Revision, current.Reason = true, true, current.Revision+1, "manual_split"
	currentVideo := book.Videos[videoIndex]
	currentVideo.VideoPrompt, currentVideo.VisualPrompt, currentVideo.ComicPrompt, currentVideo.ComicImagePrompt = "", "", "", ""
	currentVideo.Revision++
	newVideo := Video{ID: newID, BatchID: batchID, BookID: bookID, CompatibilityState: "active", Revision: 1}
	if input.Direction == "left" {
		newSegment.StartOffset, newSegment.EndOffset, newSegment.SourceText, newSegment.BoundaryType = current.StartOffset, input.BoundaryOffset, prefix, boundaryType
		current.StartOffset, current.SourceText = input.BoundaryOffset, suffix
		book.Videos = append(book.Videos[:videoIndex], append([]Video{newVideo, currentVideo}, book.Videos[videoIndex+1:]...)...)
	} else {
		current.EndOffset, current.SourceText, current.BoundaryType = input.BoundaryOffset, prefix, boundaryType
		newSegment.StartOffset, newSegment.EndOffset, newSegment.SourceText, newSegment.BoundaryType = input.BoundaryOffset, current.EndOffset+len(suffix), suffix, stored.Value.BoundaryType
		book.Videos = append(book.Videos[:videoIndex], append([]Video{currentVideo, newVideo}, book.Videos[videoIndex+1:]...)...)
	}
	current.CharacterCount = comicSourceCharacterCount(current.SourceText)
	newSegment.CharacterCount = comicSourceCharacterCount(newSegment.SourceText)
	version := current.SegmentationVersion
	for index := range book.Videos {
		book.Videos[index].Label = fmt.Sprintf("漫画页 %02d", index+1)
		if segment, ok := s.comicSourceSegments[book.Videos[index].ID]; ok && segment.Owner == owner {
			value := segment.Value
			value.Ordinal, value.SegmentationVersion = index, version
			s.comicSourceSegments[value.VideoID] = memoryOwned[ComicSourceSegment]{Owner: owner, Value: value}
		}
	}
	s.comicSourceSegments[current.VideoID] = memoryOwned[ComicSourceSegment]{Owner: owner, Value: current}
	s.comicSourceSegments[newID] = memoryOwned[ComicSourceSegment]{Owner: owner, Value: newSegment}
	for pageID, page := range s.comicPages {
		if page.Owner == owner && page.Value.VideoID == current.VideoID {
			value := page.Value
			value.Status, value.Revision, value.UpdatedAt = "prompt_stale", value.Revision+1, time.Now().UTC()
			s.comicPages[pageID] = memoryOwned[ComicPage]{Owner: owner, Value: value}
		}
	}
	drafts := make([]ComicSourceSegmentDraft, 0, len(book.Videos))
	for _, video := range book.Videos {
		segment := current
		if video.ID == newID {
			segment = newSegment
		} else if video.ID != current.VideoID {
			segment = s.comicSourceSegments[video.ID].Value
		}
		drafts = append(drafts, ComicSourceSegmentDraft{StartOffset: segment.StartOffset, EndOffset: segment.EndOffset, SourceText: segment.SourceText})
	}
	sort.SliceStable(drafts, func(i, j int) bool { return drafts[i].StartOffset < drafts[j].StartOffset })
	if err := ValidateComicSourceSegments(source, drafts); err != nil {
		return ComicSourceSegmentation{}, err
	}
	book.Revision++
	s.batches[batchID] = owned
	return s.getComicSourceSegmentationLocked(owner, batchID, bookID)
}

var _ ComicSourceSegmentStore = (*MemoryStore)(nil)
