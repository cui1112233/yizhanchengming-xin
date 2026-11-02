package agentworkspace

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrLibraryFolderNotFound = errors.New("agent library folder not found")
	ErrLibraryFolderExists   = errors.New("agent library folder already exists")
	ErrLibraryItemNotFound   = errors.New("agent library item not found")
	ErrLibraryItemExists     = errors.New("agent library item already exists")
	ErrLibraryInvalid        = errors.New("invalid agent library item")
)

// LibraryFolder is an owner-scoped collection of reusable project assets.
type LibraryFolder struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// LibraryItem stores a source asset reference only. Its object bytes remain
// owned by the original immutable agent asset.
type LibraryItem struct {
	ID              string    `json:"id"`
	FolderID        string    `json:"folderId"`
	SourceProjectID string    `json:"sourceProjectId"`
	SourceAssetID   string    `json:"sourceAssetId"`
	Kind            string    `json:"kind"`
	Name            string    `json:"name"`
	Tags            []string  `json:"tags"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CreateLibraryFolderInput struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type SaveLibraryItemInput struct {
	ID              string
	FolderID        string
	SourceProjectID string
	SourceAssetID   string
	Kind            string
	Name            string
	Tags            []string
	CreatedAt       time.Time
}

type LibraryItemFilter struct {
	Kind  string
	Query string
}

type LibraryStore interface {
	CreateLibraryFolder(context.Context, string, CreateLibraryFolderInput) (LibraryFolder, error)
	ListLibraryFolders(context.Context, string) ([]LibraryFolder, error)
	SaveLibraryItem(context.Context, string, SaveLibraryItemInput) (LibraryItem, error)
	ListLibraryItems(context.Context, string, string, LibraryItemFilter) ([]LibraryItem, error)
	LibraryItem(context.Context, string, string) (LibraryItem, error)
	MoveLibraryItem(context.Context, string, string, string, time.Time) (LibraryItem, error)
}

func normalizeLibraryFolderInput(input CreateLibraryFolderInput) (CreateLibraryFolderInput, error) {
	input.ID = strings.TrimSpace(input.ID)
	input.Name = strings.TrimSpace(input.Name)
	if !assetSafeID.MatchString(input.ID) || input.Name == "" || len(input.Name) > 120 || input.CreatedAt.IsZero() {
		return CreateLibraryFolderInput{}, ErrLibraryInvalid
	}
	return input, nil
}

func normalizeLibraryItemInput(input SaveLibraryItemInput) (SaveLibraryItemInput, error) {
	input.ID = strings.TrimSpace(input.ID)
	input.FolderID = strings.TrimSpace(input.FolderID)
	input.SourceProjectID = strings.TrimSpace(input.SourceProjectID)
	input.SourceAssetID = strings.TrimSpace(input.SourceAssetID)
	input.Kind = strings.TrimSpace(input.Kind)
	input.Name = strings.TrimSpace(input.Name)
	if !assetSafeID.MatchString(input.ID) || !assetSafeID.MatchString(input.FolderID) || !assetSafeID.MatchString(input.SourceProjectID) || !assetSafeID.MatchString(input.SourceAssetID) || !ValidAssetKind(input.Kind) || input.Name == "" || len(input.Name) > 191 || input.CreatedAt.IsZero() {
		return SaveLibraryItemInput{}, ErrLibraryInvalid
	}
	tags := make([]string, 0, len(input.Tags))
	seen := make(map[string]struct{})
	for _, tag := range input.Tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if len(tag) > 48 || len(tags) >= 12 {
			return SaveLibraryItemInput{}, ErrLibraryInvalid
		}
		key := strings.ToLower(tag)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		tags = append(tags, tag)
	}
	input.Tags = tags
	return input, nil
}

func librarySourceKey(folderID, projectID, assetID string) string {
	return folderID + "\x00" + projectID + "\x00" + assetID
}
