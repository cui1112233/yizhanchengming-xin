package agentworkspace

import (
	"context"
	"sort"
	"strings"
	"time"
)

func libraryFolderKey(owner, id string) string { return owner + "\x00" + id }
func libraryItemKey(owner, id string) string   { return owner + "\x00" + id }

func (s *MemoryStore) CreateLibraryFolder(_ context.Context, owner string, input CreateLibraryFolderInput) (LibraryFolder, error) {
	input, err := normalizeLibraryFolderInput(input)
	if err != nil || strings.TrimSpace(owner) == "" {
		return LibraryFolder{}, ErrLibraryInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := libraryFolderKey(owner, input.ID)
	if _, exists := s.libraryFolders[key]; exists {
		return LibraryFolder{}, ErrLibraryFolderExists
	}
	folder := LibraryFolder{ID: input.ID, Name: input.Name, CreatedAt: input.CreatedAt.UTC(), UpdatedAt: input.CreatedAt.UTC()}
	s.libraryFolders[key] = folder
	return folder, nil
}

func (s *MemoryStore) ListLibraryFolders(_ context.Context, owner string) ([]LibraryFolder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]LibraryFolder, 0)
	for key, folder := range s.libraryFolders {
		if strings.HasPrefix(key, strings.TrimSpace(owner)+"\x00") {
			items = append(items, folder)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].ID < items[j].ID
		}
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	return items, nil
}

func (s *MemoryStore) SaveLibraryItem(_ context.Context, owner string, input SaveLibraryItemInput) (LibraryItem, error) {
	input, err := normalizeLibraryItemInput(input)
	if err != nil || strings.TrimSpace(owner) == "" {
		return LibraryItem{}, ErrLibraryInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.libraryFolders[libraryFolderKey(owner, input.FolderID)]; !exists {
		return LibraryItem{}, ErrLibraryFolderNotFound
	}
	if _, exists := s.libraryItems[libraryItemKey(owner, input.ID)]; exists {
		return LibraryItem{}, ErrLibraryItemExists
	}
	for key, item := range s.libraryItems {
		if strings.HasPrefix(key, owner+"\x00") && librarySourceKey(item.FolderID, item.SourceProjectID, item.SourceAssetID) == librarySourceKey(input.FolderID, input.SourceProjectID, input.SourceAssetID) {
			return LibraryItem{}, ErrLibraryItemExists
		}
	}
	item := LibraryItem{ID: input.ID, FolderID: input.FolderID, SourceProjectID: input.SourceProjectID, SourceAssetID: input.SourceAssetID, Kind: input.Kind, Name: input.Name, Tags: append([]string(nil), input.Tags...), CreatedAt: input.CreatedAt.UTC(), UpdatedAt: input.CreatedAt.UTC()}
	s.libraryItems[libraryItemKey(owner, item.ID)] = item
	return item, nil
}

func (s *MemoryStore) ListLibraryItems(_ context.Context, owner, folderID string, filter LibraryItemFilter) ([]LibraryItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	folderID = strings.TrimSpace(folderID)
	if _, exists := s.libraryFolders[libraryFolderKey(owner, folderID)]; !exists {
		return nil, ErrLibraryFolderNotFound
	}
	kind := strings.TrimSpace(filter.Kind)
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	items := make([]LibraryItem, 0)
	for key, item := range s.libraryItems {
		if !strings.HasPrefix(key, owner+"\x00") || item.FolderID != folderID || (kind != "" && item.Kind != kind) || !libraryItemMatches(item, query) {
			continue
		}
		item.Tags = append([]string(nil), item.Tags...)
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
	return items, nil
}

func (s *MemoryStore) LibraryItem(_ context.Context, owner, id string) (LibraryItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, exists := s.libraryItems[libraryItemKey(owner, strings.TrimSpace(id))]
	if !exists {
		return LibraryItem{}, ErrLibraryItemNotFound
	}
	item.Tags = append([]string(nil), item.Tags...)
	return item, nil
}

func (s *MemoryStore) MoveLibraryItem(_ context.Context, owner, itemID, folderID string, now time.Time) (LibraryItem, error) {
	if now.IsZero() {
		return LibraryItem{}, ErrLibraryInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	folderID = strings.TrimSpace(folderID)
	if _, exists := s.libraryFolders[libraryFolderKey(owner, folderID)]; !exists {
		return LibraryItem{}, ErrLibraryFolderNotFound
	}
	key := libraryItemKey(owner, strings.TrimSpace(itemID))
	item, exists := s.libraryItems[key]
	if !exists {
		return LibraryItem{}, ErrLibraryItemNotFound
	}
	for candidateKey, candidate := range s.libraryItems {
		if candidateKey != key && strings.HasPrefix(candidateKey, owner+"\x00") && librarySourceKey(folderID, candidate.SourceProjectID, candidate.SourceAssetID) == librarySourceKey(folderID, item.SourceProjectID, item.SourceAssetID) {
			return LibraryItem{}, ErrLibraryItemExists
		}
	}
	item.FolderID = folderID
	item.UpdatedAt = now.UTC()
	s.libraryItems[key] = item
	return item, nil
}

func libraryItemMatches(item LibraryItem, query string) bool {
	if query == "" || strings.Contains(strings.ToLower(item.Name), query) {
		return true
	}
	for _, tag := range item.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}
