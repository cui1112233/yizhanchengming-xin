package aicut

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrAssetFolderNotFound = errors.New("asset folder not found")
var ErrInvalidAssetFolder = errors.New("invalid asset folder")

type AssetFolder struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// Folders only organize project assets. Deleting a folder never deletes media.
type AssetFolderStore interface {
	ListAssetFolders(context.Context, string, string) ([]AssetFolder, error)
	CreateAssetFolder(context.Context, string, AssetFolder) error
	RenameAssetFolder(context.Context, string, string, string, string) error
	DeleteAssetFolder(context.Context, string, string, string) error
	MoveAssetToFolder(context.Context, string, string, string, string) error
	AssetFolderAssignments(context.Context, string, string) (map[string]string, error)
}

func validAssetFolderName(name string) bool {
	return strings.TrimSpace(name) != "" && utf8.RuneCountInString(strings.TrimSpace(name)) <= 100 && !strings.ContainsAny(name, "\x00\r\n")
}
func (s *MemoryStore) folderProjectLocked(owner, project string) error {
	p, ok := s.projects[aicutProjectKey(owner, project)]
	if !ok || p.project.DeletedAt != nil {
		return ErrProjectNotFound
	}
	return nil
}
func (s *MemoryStore) ListAssetFolders(_ context.Context, owner, project string) ([]AssetFolder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.folderProjectLocked(owner, project); err != nil {
		return nil, err
	}
	out := make([]AssetFolder, 0)
	prefix := aicutAssetKey(owner, project, "")
	for k, v := range s.assetFolders {
		if strings.HasPrefix(k, prefix) {
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out, nil
}
func (s *MemoryStore) CreateAssetFolder(_ context.Context, owner string, folder AssetFolder) error {
	if folder.ID == "" || !validAssetFolderName(folder.Name) {
		return ErrInvalidAssetFolder
	}
	folder.Name = strings.TrimSpace(folder.Name)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.folderProjectLocked(owner, folder.ProjectID); err != nil {
		return err
	}
	if s.assetFolders == nil {
		s.assetFolders = make(map[string]AssetFolder)
	}
	key := aicutAssetKey(owner, folder.ProjectID, folder.ID)
	if _, ok := s.assetFolders[key]; ok {
		return ErrInvalidAssetFolder
	}
	s.assetFolders[key] = folder
	return nil
}
func (s *MemoryStore) RenameAssetFolder(_ context.Context, owner, project, id, name string) error {
	if !validAssetFolderName(name) {
		return ErrInvalidAssetFolder
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.folderProjectLocked(owner, project); err != nil {
		return err
	}
	key := aicutAssetKey(owner, project, id)
	folder, ok := s.assetFolders[key]
	if !ok {
		return ErrAssetFolderNotFound
	}
	folder.Name = strings.TrimSpace(name)
	s.assetFolders[key] = folder
	return nil
}
func (s *MemoryStore) DeleteAssetFolder(_ context.Context, owner, project, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.folderProjectLocked(owner, project); err != nil {
		return err
	}
	key := aicutAssetKey(owner, project, id)
	if _, ok := s.assetFolders[key]; !ok {
		return ErrAssetFolderNotFound
	}
	delete(s.assetFolders, key)
	prefix := aicutAssetKey(owner, project, "")
	for k, v := range s.assetFolderAssignments {
		if strings.HasPrefix(k, prefix) && v == id {
			delete(s.assetFolderAssignments, k)
		}
	}
	return nil
}
func (s *MemoryStore) MoveAssetToFolder(_ context.Context, owner, project, asset, folder string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.folderProjectLocked(owner, project); err != nil {
		return err
	}
	key := aicutAssetKey(owner, project, asset)
	if _, ok := s.assets[key]; !ok {
		return ErrAssetNotFound
	}
	if folder != "" {
		if _, ok := s.assetFolders[aicutAssetKey(owner, project, folder)]; !ok {
			return ErrAssetFolderNotFound
		}
	}
	if s.assetFolderAssignments == nil {
		s.assetFolderAssignments = make(map[string]string)
	}
	if folder == "" {
		delete(s.assetFolderAssignments, key)
	} else {
		s.assetFolderAssignments[key] = folder
	}
	return nil
}
func (s *MemoryStore) AssetFolderAssignments(_ context.Context, owner, project string) (map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.folderProjectLocked(owner, project); err != nil {
		return nil, err
	}
	out := map[string]string{}
	prefix := aicutAssetKey(owner, project, "")
	for k, v := range s.assetFolderAssignments {
		if strings.HasPrefix(k, prefix) {
			out[strings.TrimPrefix(k, prefix)] = v
		}
	}
	return out, nil
}
