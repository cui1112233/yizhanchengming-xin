package batchfactoryv11

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// applyComicPagePatch merges a client patch into an existing page row. Only
// fields that were actually provided are replaced; the final row still has to
// satisfy the comic page invariant (both prompts non-empty).
func applyComicPagePatch(base UpsertComicPageInput, patch PatchComicPageInput) (UpsertComicPageInput, error) {
	if patch.ComicPrompt != nil {
		base.ComicPrompt = strings.TrimSpace(*patch.ComicPrompt)
	}
	if patch.ImagePrompt != nil {
		base.ImagePrompt = strings.TrimSpace(*patch.ImagePrompt)
	}
	if patch.ModelID != nil {
		if value := strings.TrimSpace(*patch.ModelID); value != "" {
			base.ModelID = value
		}
	}
	if patch.AspectRatio != nil {
		if value := strings.TrimSpace(*patch.AspectRatio); value != "" {
			base.AspectRatio = value
		}
	}
	if patch.Status != nil {
		// An empty status means "leave it alone"; a real status is applied.
		if value := strings.TrimSpace(*patch.Status); value != "" {
			base.Status = value
		}
	}
	if patch.ReferenceImageURLs != nil {
		base.ReferenceImageURLs = normalizeComicReferenceURLs(*patch.ReferenceImageURLs)
	}
	base.ComicPrompt = strings.TrimSpace(base.ComicPrompt)
	base.ImagePrompt = strings.TrimSpace(base.ImagePrompt)
	if base.ComicPrompt == "" || base.ImagePrompt == "" {
		return UpsertComicPageInput{}, ErrInvalid
	}
	base.ModelID = strings.TrimSpace(base.ModelID)
	base.AspectRatio = strings.TrimSpace(base.AspectRatio)
	if base.AspectRatio == "" {
		base.AspectRatio = "9:16"
	}
	base.Status = strings.TrimSpace(base.Status)
	if base.Status == "" {
		base.Status = "prompt_ready"
	}
	base.SnapshotHash = strings.TrimSpace(base.SnapshotHash)
	base.ReferenceImageURLs = normalizeComicReferenceURLs(base.ReferenceImageURLs)
	return base, nil
}

func normalizeComicPageInput(input UpsertComicPageInput) (UpsertComicPageInput, error) {
	input.ComicPrompt = strings.TrimSpace(input.ComicPrompt)
	input.ImagePrompt = strings.TrimSpace(input.ImagePrompt)
	input.ModelID = strings.TrimSpace(input.ModelID)
	input.AspectRatio = strings.TrimSpace(input.AspectRatio)
	input.Status = strings.TrimSpace(input.Status)
	input.SnapshotHash = strings.TrimSpace(input.SnapshotHash)
	input.ReferenceImageURLs = normalizeComicReferenceURLs(input.ReferenceImageURLs)
	if input.ComicPrompt == "" || input.ImagePrompt == "" {
		return UpsertComicPageInput{}, ErrInvalid
	}
	if input.AspectRatio == "" {
		input.AspectRatio = "9:16"
	}
	if input.Status == "" {
		input.Status = "prompt_ready"
	}
	return input, nil
}

func normalizeComicReferenceURLs(urls []string) []string {
	result := make([]string, 0, len(urls))
	seen := make(map[string]bool, len(urls))
	for _, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" || seen[url] {
			continue
		}
		seen[url] = true
		result = append(result, url)
	}
	return result
}

func normalizeComicPageImageInput(input CreateComicPageImageInput) (CreateComicPageImageInput, error) {
	valid, err := validBookAssetImageInput(CreateBookAssetImageInput(input))
	if err != nil {
		return CreateComicPageImageInput{}, err
	}
	return CreateComicPageImageInput(valid), nil
}

func cloneComicPage(page ComicPage) ComicPage {
	page.Images = append([]ComicPageImage(nil), page.Images...)
	page.ReferenceImageURLs = append([]string(nil), page.ReferenceImageURLs...)
	return page
}

func cloneComicCollection(collection ComicCollection) ComicCollection {
	collection.Items = append([]ComicCollectionItem(nil), collection.Items...)
	return collection
}

func (s *MemoryStore) ownedBookLocked(owner, batchID, bookID string) (Book, error) {
	entry, ok := s.batches[batchID]
	if !ok || entry.Owner != owner {
		return Book{}, ErrNotFound
	}
	for _, book := range entry.Value.Books {
		if book.ID == bookID {
			return book, nil
		}
	}
	return Book{}, ErrNotFound
}

func (s *MemoryStore) UpsertComicPage(_ context.Context, owner, batchID, bookID, videoID string, input UpsertComicPageInput) (ComicPage, error) {
	input, err := normalizeComicPageInput(input)
	if err != nil {
		return ComicPage{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	book, err := s.ownedBookLocked(owner, batchID, bookID)
	if err != nil {
		return ComicPage{}, err
	}
	foundVideo := false
	for _, video := range book.Videos {
		if video.ID == videoID {
			foundVideo = true
			break
		}
	}
	if !foundVideo {
		return ComicPage{}, ErrNotFound
	}
	now := time.Now().UTC()
	for id, entry := range s.comicPages {
		if entry.Owner == owner && entry.Value.BatchID == batchID && entry.Value.BookID == bookID && entry.Value.VideoID == videoID {
			page := entry.Value
			if input.ExpectedRevision > 0 && page.Revision != input.ExpectedRevision {
				return ComicPage{}, fmt.Errorf("%w: 漫画页版本已变化", ErrConflict)
			}
			page.ComicPrompt, page.ImagePrompt, page.ModelID = input.ComicPrompt, input.ImagePrompt, input.ModelID
			page.AspectRatio, page.Status, page.SnapshotHash, page.ReferenceImageURLs, page.UpdatedAt, page.Revision = input.AspectRatio, input.Status, input.SnapshotHash, input.ReferenceImageURLs, now, page.Revision+1
			s.comicPages[id] = memoryOwned[ComicPage]{Owner: owner, Value: page}
			return cloneComicPage(page), nil
		}
	}
	if input.ExpectedRevision > 0 {
		// 期望版本大于 0 说明调用方在写回一版已有页面；此时找不到行同样
		// 属于版本冲突（页面被删除或重建）。
		return ComicPage{}, fmt.Errorf("%w: 漫画页版本已变化", ErrConflict)
	}
	s.seq++
	page := ComicPage{ID: "comic-page-" + strconv.FormatInt(s.seq, 10), BatchID: batchID, BookID: bookID, VideoID: videoID, ComicPrompt: input.ComicPrompt, ImagePrompt: input.ImagePrompt, ModelID: input.ModelID, AspectRatio: input.AspectRatio, Status: input.Status, SnapshotHash: input.SnapshotHash, ReferenceImageURLs: input.ReferenceImageURLs, Revision: 1, CreatedAt: now, UpdatedAt: now}
	s.comicPages[page.ID] = memoryOwned[ComicPage]{Owner: owner, Value: page}
	return cloneComicPage(page), nil
}

func (s *MemoryStore) PatchComicPage(_ context.Context, owner, batchID, bookID, videoID string, patch PatchComicPageInput) (ComicPage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	book, err := s.ownedBookLocked(owner, batchID, bookID)
	if err != nil {
		return ComicPage{}, err
	}
	foundVideo := false
	for _, video := range book.Videos {
		if video.ID == videoID {
			foundVideo = true
			break
		}
	}
	if !foundVideo {
		return ComicPage{}, ErrNotFound
	}
	now := time.Now().UTC()
	for id, entry := range s.comicPages {
		if entry.Owner != owner || entry.Value.BatchID != batchID || entry.Value.BookID != bookID || entry.Value.VideoID != videoID {
			continue
		}
		page := entry.Value
		merged, mergeErr := applyComicPagePatch(UpsertComicPageInput{
			ComicPrompt: page.ComicPrompt, ImagePrompt: page.ImagePrompt, ReferenceImageURLs: page.ReferenceImageURLs,
			ModelID: page.ModelID, AspectRatio: page.AspectRatio, Status: page.Status, SnapshotHash: page.SnapshotHash,
		}, patch)
		if mergeErr != nil {
			return ComicPage{}, mergeErr
		}
		page.ComicPrompt, page.ImagePrompt, page.ModelID = merged.ComicPrompt, merged.ImagePrompt, merged.ModelID
		page.AspectRatio, page.Status, page.SnapshotHash, page.ReferenceImageURLs = merged.AspectRatio, merged.Status, merged.SnapshotHash, merged.ReferenceImageURLs
		page.UpdatedAt, page.Revision = now, page.Revision+1
		s.comicPages[id] = memoryOwned[ComicPage]{Owner: owner, Value: page}
		return cloneComicPage(page), nil
	}
	merged, err := applyComicPagePatch(UpsertComicPageInput{}, patch)
	if err != nil {
		return ComicPage{}, err
	}
	s.seq++
	page := ComicPage{ID: "comic-page-" + strconv.FormatInt(s.seq, 10), BatchID: batchID, BookID: bookID, VideoID: videoID, ComicPrompt: merged.ComicPrompt, ImagePrompt: merged.ImagePrompt, ModelID: merged.ModelID, AspectRatio: merged.AspectRatio, Status: merged.Status, SnapshotHash: merged.SnapshotHash, ReferenceImageURLs: merged.ReferenceImageURLs, Revision: 1, CreatedAt: now, UpdatedAt: now}
	s.comicPages[page.ID] = memoryOwned[ComicPage]{Owner: owner, Value: page}
	return cloneComicPage(page), nil
}

func (s *MemoryStore) SetComicPageReferenceImages(_ context.Context, owner, batchID, bookID, videoID string, urls []string) (ComicPage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.ownedBookLocked(owner, batchID, bookID); err != nil {
		return ComicPage{}, err
	}
	for id, entry := range s.comicPages {
		page := entry.Value
		if entry.Owner != owner || page.BatchID != batchID || page.BookID != bookID || page.VideoID != videoID {
			continue
		}
		page.ReferenceImageURLs = normalizeComicReferenceURLs(urls)
		page.Revision++
		page.UpdatedAt = time.Now().UTC()
		s.comicPages[id] = memoryOwned[ComicPage]{Owner: owner, Value: page}
		return cloneComicPage(page), nil
	}
	return ComicPage{}, ErrNotFound
}

func (s *MemoryStore) ApplyComicPageReferenceImages(_ context.Context, owner, batchID, bookID, sourceVideoID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.ownedBookLocked(owner, batchID, bookID); err != nil {
		return 0, err
	}
	var source ComicPage
	found := false
	for _, entry := range s.comicPages {
		page := entry.Value
		if entry.Owner == owner && page.BatchID == batchID && page.BookID == bookID && page.VideoID == sourceVideoID {
			source, found = page, true
			break
		}
	}
	if !found {
		return 0, ErrNotFound
	}
	now, updated := time.Now().UTC(), 0
	for id, entry := range s.comicPages {
		page := entry.Value
		if entry.Owner != owner || page.BatchID != batchID || page.BookID != bookID {
			continue
		}
		page.ReferenceImageURLs = append([]string(nil), source.ReferenceImageURLs...)
		page.Revision++
		page.UpdatedAt = now
		s.comicPages[id] = memoryOwned[ComicPage]{Owner: owner, Value: page}
		updated++
	}
	return updated, nil
}

func (s *MemoryStore) ListComicPages(_ context.Context, owner, batchID, bookID string) ([]ComicPage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	book, err := s.ownedBookLocked(owner, batchID, bookID)
	if err != nil {
		return nil, err
	}
	byVideo := map[string]ComicPage{}
	for _, entry := range s.comicPages {
		if entry.Owner == owner && entry.Value.BatchID == batchID && entry.Value.BookID == bookID {
			page := cloneComicPage(entry.Value)
			page.Images = nil
			for _, imageEntry := range s.comicPageImages {
				if imageEntry.Owner == owner && imageEntry.Value.ComicPageID == page.ID {
					page.Images = append(page.Images, imageEntry.Value)
				}
			}
			byVideo[page.VideoID] = page
		}
	}
	pages := make([]ComicPage, 0, len(byVideo))
	for _, video := range book.Videos {
		if page, ok := byVideo[video.ID]; ok {
			pages = append(pages, page)
		}
	}
	return pages, nil
}

func (s *MemoryStore) CreateComicPageImage(_ context.Context, owner, batchID, bookID, pageID string, input CreateComicPageImageInput) (ComicPageImage, error) {
	input, err := normalizeComicPageImageInput(input)
	if err != nil {
		return ComicPageImage{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	pageEntry, ok := s.comicPages[pageID]
	if !ok || pageEntry.Owner != owner || pageEntry.Value.BatchID != batchID || pageEntry.Value.BookID != bookID {
		return ComicPageImage{}, ErrNotFound
	}
	isPrimary := true
	for _, entry := range s.comicPageImages {
		if entry.Owner == owner && entry.Value.ComicPageID == pageID && entry.Value.IsPrimary {
			isPrimary = false
			break
		}
	}
	s.seq++
	now := time.Now().UTC()
	image := ComicPageImage{ID: "comic-image-" + strconv.FormatInt(s.seq, 10), ComicPageID: pageID, URL: input.URL, StorageRef: input.StorageRef, MediaType: input.MediaType, Source: input.Source, IsPrimary: isPrimary, Revision: 1, CreatedAt: now}
	s.comicPageImages[image.ID] = memoryOwned[ComicPageImage]{Owner: owner, Value: image}
	pageEntry.Value.Status = "image_ready"
	pageEntry.Value.Revision++
	pageEntry.Value.UpdatedAt = now
	s.comicPages[pageID] = pageEntry
	return image, nil
}

func (s *MemoryStore) SetPrimaryComicPageImage(_ context.Context, owner, batchID, bookID, pageID, imageID string) (ComicPageImage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pageEntry, ok := s.comicPages[pageID]
	if !ok || pageEntry.Owner != owner || pageEntry.Value.BatchID != batchID || pageEntry.Value.BookID != bookID {
		return ComicPageImage{}, ErrNotFound
	}
	target, ok := s.comicPageImages[imageID]
	if !ok || target.Owner != owner || target.Value.ComicPageID != pageID {
		return ComicPageImage{}, ErrNotFound
	}
	for id, entry := range s.comicPageImages {
		if entry.Owner == owner && entry.Value.ComicPageID == pageID && entry.Value.IsPrimary != (id == imageID) {
			entry.Value.IsPrimary = id == imageID
			entry.Value.Revision++
			s.comicPageImages[id] = entry
		}
	}
	return s.comicPageImages[imageID].Value, nil
}

func (s *MemoryStore) CreateComicCollection(_ context.Context, owner, batchID, bookID string, requested ...[]string) (ComicCollection, []string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	book, err := s.ownedBookLocked(owner, batchID, bookID)
	if err != nil {
		return ComicCollection{}, nil, err
	}
	var requestedIDs []string
	if len(requested) > 0 {
		requestedIDs = requested[0]
	}
	selectedIDs, unknown := SelectComicCollectionVideoIDs(book.Videos, requestedIDs)
	if len(unknown) > 0 || len(selectedIDs) == 0 {
		return ComicCollection{}, unknown, ErrConflict
	}
	selectedSet := make(map[string]struct{}, len(selectedIDs))
	for _, id := range selectedIDs {
		selectedSet[id] = struct{}{}
	}
	items := make([]ComicCollectionItem, 0, len(selectedIDs))
	missing := make([]string, 0)
	for _, video := range book.Videos {
		if _, ok := selectedSet[video.ID]; !ok {
			continue
		}
		var page ComicPage
		for _, entry := range s.comicPages {
			if entry.Owner == owner && entry.Value.BatchID == batchID && entry.Value.BookID == bookID && entry.Value.VideoID == video.ID {
				page = entry.Value
				break
			}
		}
		var primary ComicPageImage
		for _, entry := range s.comicPageImages {
			if entry.Owner == owner && entry.Value.ComicPageID == page.ID && entry.Value.IsPrimary {
				primary = entry.Value
				break
			}
		}
		if page.ID == "" || primary.ID == "" {
			missing = append(missing, video.ID)
			continue
		}
		items = append(items, ComicCollectionItem{Ordinal: len(items) + 1, VideoID: video.ID, ComicPageID: page.ID, ImageID: primary.ID, StorageRef: primary.StorageRef, MediaType: primary.MediaType})
	}
	if len(missing) > 0 {
		return ComicCollection{}, missing, ErrConflict
	}
	for _, entry := range s.comicCollections {
		if entry.Owner == owner && entry.Value.BatchID == batchID && entry.Value.BookID == bookID && sameComicCollectionItems(entry.Value.Items, items) {
			return cloneComicCollection(entry.Value), nil, nil
		}
	}
	s.seq++
	collection := ComicCollection{ID: "comic-collection-" + strconv.FormatInt(s.seq, 10), BatchID: batchID, BookID: bookID, Status: "ready", Items: items, CreatedAt: time.Now().UTC()}
	s.comicCollections[collection.ID] = memoryOwned[ComicCollection]{Owner: owner, Value: collection}
	return cloneComicCollection(collection), nil, nil
}

func (s *MemoryStore) ListComicCollections(_ context.Context, owner, batchID, bookID string) ([]ComicCollection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.ownedBookLocked(owner, batchID, bookID); err != nil {
		return nil, err
	}
	collections := []ComicCollection{}
	for _, entry := range s.comicCollections {
		if entry.Owner == owner && entry.Value.BatchID == batchID && entry.Value.BookID == bookID {
			collections = append(collections, cloneComicCollection(entry.Value))
		}
	}
	sort.SliceStable(collections, func(i, j int) bool {
		return collections[i].CreatedAt.Before(collections[j].CreatedAt)
	})
	return latestDistinctComicCollections(collections), nil
}
