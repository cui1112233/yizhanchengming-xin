package screenplay

import (
	"context"
	"time"
)

// AssetLibraryEntry is CM's durable ordinary asset library adapter. The recovered
// frontend's image-binding semantics remain separate from this storage entity.
type AssetLibraryEntry struct {
	SourceMediaID   string                  `json:"source_media_id,omitempty"`
	ImageObject     AssetLibraryImageObject `json:"image_object"`
	ID              string                  `json:"id"`
	Owner           string                  `json:"-"`
	Type            AssetType               `json:"type"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	Category        string                  `json:"category"`
	Image           MediaItem               `json:"image"`
	Snapshot        Asset                   `json:"snapshot"`
	SourceProjectID string                  `json:"source_project_id"`
	SourceAssetID   string                  `json:"source_asset_id"`
	Revision        int64                   `json:"revision"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

// ImageObject is private persisted storage metadata. HTTP handlers return an
// explicit public DTO, never this object or a storage credential.
type AssetLibraryImageObject struct {
	ObjectKey   string `json:"object_key"`
	ContentType string `json:"content_type"`
	ByteSize    int64  `json:"byte_size"`
}
type AssetLibraryRepository interface {
	ListAssetLibrary(context.Context, string, AssetType, string, string) ([]AssetLibraryEntry, error)
	GetAssetLibraryEntry(context.Context, string, string) (AssetLibraryEntry, error)
	SaveAssetLibraryEntry(context.Context, AssetLibraryEntry, int64) (AssetLibraryEntry, error)
}
