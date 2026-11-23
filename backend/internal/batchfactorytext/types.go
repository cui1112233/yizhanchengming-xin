package batchfactorytext

import (
	"encoding/json"
	"strings"
)

// Scope identifies the owning account, batch and internal book (not the bookstore ID).
type Scope struct{ Owner, BatchID, BookID string }

type LayoutConfig struct {
	AspectRatio      string  `json:"aspectRatio"`
	TargetCharacters int     `json:"targetCharacters"`
	FontID           string  `json:"fontId"`
	FontSize         float64 `json:"fontSize"`
	FontWeight       int     `json:"fontWeight"`
	LineHeight       float64 `json:"lineHeight"`
	ParagraphSpacing float64 `json:"paragraphSpacing"`
	Padding          int     `json:"padding"`
	Align            string  `json:"align"`
	TextColor        string  `json:"textColor"`
	HighlightColor   string  `json:"highlightColor"`
	MaxHeightRatio   float64 `json:"maxHeightRatio"`
}

type Background struct {
	Kind  string `json:"kind"`
	Color string `json:"color,omitempty"`
	ID    string `json:"id,omitempty"`
}

type Document struct {
	ID                string          `json:"id"`
	OriginalText      string          `json:"originalText"`
	ActiveVersionID   string          `json:"activeVersionId"`
	PaginationVersion int64           `json:"paginationVersion"`
	ConfigSnapshot    json.RawMessage `json:"configSnapshot,omitempty"`
	CoverArtifactID   string          `json:"coverArtifactId,omitempty"`
	CoverLocked       bool            `json:"coverLocked"`
	Revision          int64           `json:"revision"`
}

type Version struct {
	ID             string          `json:"id"`
	ProductionText string          `json:"productionText"`
	SensitiveMode  string          `json:"sensitiveMode"`
	ChangeTrace    json.RawMessage `json:"changeTrace,omitempty"`
	ContentDigest  string          `json:"contentDigest"`
}

type VersionInput struct {
	ExpectedRevision              int64
	ProductionText, SensitiveMode string
	ChangeTrace                   json.RawMessage
}

// NormalizeLineEndings supplies the shared coordinate space for pagination,
// persisted offsets and page edits. The original and version still retain
// the exact input, including CRLF in uploaded files.
func NormalizeLineEndings(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}

// Offsets are half-open rune offsets into the newline-normalized version.
type PageDraft struct {
	Text        string        `json:"text"`
	StartOffset int           `json:"startOffset"`
	EndOffset   int           `json:"endOffset"`
	Layout      *LayoutConfig `json:"layout,omitempty"`
	Background  *Background   `json:"background,omitempty"`
}

type Page struct {
	PageDraft
	ID                string `json:"id"`
	VersionID         string `json:"versionId"`
	PaginationVersion int64  `json:"paginationVersion"`
	Ordinal           int    `json:"ordinal"`
	ContentDigest     string `json:"contentDigest"`
	SourceKind        string `json:"sourceKind"`
	Locked            bool   `json:"locked"`
	RenderVersion     int64  `json:"renderVersion"`
	ActiveArtifactID  string `json:"activeArtifactId,omitempty"`
	Status            string `json:"status"`
	FailedStage       string `json:"failedStage,omitempty"`
	ErrorType         string `json:"errorType,omitempty"`
	LastError         string `json:"lastError,omitempty"`
	Revision          int64  `json:"revision"`
}

type PaginationUpdate struct {
	ExpectedRevision int64
	VersionID        string
	ConfigSnapshot   json.RawMessage
	Pages            []PageDraft
}

// Text edits create a new Version and pagination. These edits only redraw a page.
type PagePatch struct {
	Layout     *LayoutConfig
	Background *Background
}

type ArtifactFile struct {
	StorageRef string `json:"storageRef"`
	MediaType  string `json:"mediaType"`
	ByteSize   int64  `json:"byteSize"`
	SHA256     string `json:"sha256"`
}

type PageArtifact struct {
	ArtifactFile
	ID            string `json:"id"`
	PageID        string `json:"pageId,omitempty"`
	Kind          string `json:"kind"`
	SourceKind    string `json:"sourceKind"`
	RenderVersion int64  `json:"renderVersion"`
}

type CollectionItem struct {
	Ordinal    int    `json:"ordinal"`
	PageID     string `json:"pageId"`
	ArtifactID string `json:"artifactId"`
	StorageRef string `json:"storageRef"`
	MediaType  string `json:"mediaType"`
	SHA256     string `json:"sha256"`
}

type Collection struct {
	ID                string           `json:"id"`
	VersionID         string           `json:"versionId"`
	PaginationVersion int64            `json:"paginationVersion"`
	CoverArtifactID   string           `json:"coverArtifactId"`
	Status            string           `json:"status"`
	StorageRef        string           `json:"storageRef,omitempty"`
	ConfigSnapshot    json.RawMessage  `json:"configSnapshot,omitempty"`
	Items             []CollectionItem `json:"items,omitempty"`
}

type CollectionInput struct {
	VersionID         string
	PaginationVersion int64
	PageIDs           []string // Empty selects every page, otherwise uses book order.
	ConfigSnapshot    json.RawMessage
}
