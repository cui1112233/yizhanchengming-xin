package batchfactorytext

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
)

// Scope identifies the owning account, batch and internal book (not the bookstore ID).
type Scope struct{ Owner, BatchID, BookID string }

type LayoutConfig struct {
	AspectRatio                     string  `json:"aspectRatio"`
	TargetCharacters                int     `json:"targetCharacters"`
	FontID                          string  `json:"fontId"`
	FontSize                        float64 `json:"fontSize"`
	FontWeight                      int     `json:"fontWeight"`
	LineHeight                      float64 `json:"lineHeight"`
	ParagraphSpacing                float64 `json:"paragraphSpacing"`
	Padding                         int     `json:"padding"`
	Align                           string  `json:"align"`
	TextColor                       string  `json:"textColor"`
	HighlightColor                  string  `json:"highlightColor"`
	MaxHeightRatio                  float64 `json:"maxHeightRatio"`
	paddingSet, paragraphSpacingSet bool
}

// Wire fields use pointers for the two measurements for which zero has
// an independent meaning. Old snapshots can contain either a full layout
// or a sparse object, and explicit zeros must survive the next save.
type layoutValues LayoutConfig
type layoutWire struct {
	layoutValues
	Padding          *int     `json:"padding,omitempty"`
	ParagraphSpacing *float64 `json:"paragraphSpacing,omitempty"`
}

func (c *LayoutConfig) UnmarshalJSON(data []byte) error {
	var w layoutWire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	*c = LayoutConfig(w.layoutValues)
	if w.Padding != nil {
		c.Padding = *w.Padding
		c.paddingSet = c.Padding == 0
	}
	if w.ParagraphSpacing != nil {
		c.ParagraphSpacing = *w.ParagraphSpacing
		c.paragraphSpacingSet = c.ParagraphSpacing == 0
	}
	return nil
}
func (c LayoutConfig) MarshalJSON() ([]byte, error) {
	w := layoutWire{layoutValues: layoutValues(c)}
	if c.Padding != 0 || c.paddingSet {
		w.Padding = &c.Padding
	}
	if c.ParagraphSpacing != 0 || c.paragraphSpacingSet {
		w.ParagraphSpacing = &c.ParagraphSpacing
	}
	return json.Marshal(w)
}

// Nonzero spacing is intrinsically present, independent of construction in Go
// or decoding from JSON. Canonical overrides compare equal across persistence.
func (c LayoutConfig) canonicalSpacing() LayoutConfig {
	c.paddingSet = c.Padding == 0 && c.paddingSet
	c.paragraphSpacingSet = c.ParagraphSpacing == 0 && c.paragraphSpacingSet
	return c
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
	// Only a page-local edit can identify a retained source page after its offsets shift.
	PreviousPageID     string        `json:"-"`
	VerifiedArtifactID string        `json:"-"`
	Text               string        `json:"text"`
	StartOffset        int           `json:"startOffset"`
	EndOffset          int           `json:"endOffset"`
	Layout             *LayoutConfig `json:"layout,omitempty"`
	Background         *Background   `json:"background,omitempty"`
	ClearLayout        bool          `json:"clearLayout,omitempty"`
	ClearBackground    bool          `json:"clearBackground,omitempty"`
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
	OriginalText     string // A zero expected revision creates the document and activates its pages atomically.
	ExpectedRevision int64
	VersionID        string
	ConfigSnapshot   json.RawMessage // Omit to keep the book settings; {} starts from defaults.
	Pages            []PageDraft
	// Lock and check the upstream production source in this publication transaction.
	VerifyProduction func(context.Context, *sql.Tx) error
}

// Text edits create a new Version and pagination. These edits only redraw a page.
type PagePatch struct {
	DocumentRevision int64         `json:"-"` // Check an editor's pinned document while holding the same lock as the page write.
	Layout           *LayoutConfig `json:"layout,omitempty"`
	Background       *Background   `json:"background,omitempty"`
	ClearLayout      bool          `json:"clearLayout,omitempty"`
	ClearBackground  bool          `json:"clearBackground,omitempty"`
}

// PageStyle is the render-affecting projection of a book configuration.
// If highlight records are version-specific, leave HighlightRevision empty to
// scope reuse to the production version; otherwise supply their stable digest.
type PageStyle struct {
	Layout                LayoutConfig `json:"layout"`
	Background            Background   `json:"background"`
	HighlightReuseEnabled bool         `json:"highlightReuseEnabled"`
	HighlightRevision     string       `json:"highlightRevision,omitempty"`
}

func PageStyleFromSnapshot(snapshot json.RawMessage) (PageStyle, error) {
	var s PageStyle
	if len(snapshot) == 0 {
		return s, nil
	}
	// Earlier layout snapshots can contain the layout fields at the top level.
	// When both forms are present the nested layout takes precedence.
	var raw struct {
		PageStyle
		Layout *LayoutConfig `json:"layout"`
	}
	if err := json.Unmarshal(snapshot, &raw); err != nil {
		return s, err
	}
	s = raw.PageStyle
	if raw.Layout != nil {
		s.Layout = *raw.Layout
	} else if err := json.Unmarshal(snapshot, &s.Layout); err != nil {
		return s, err
	}
	if !s.HighlightReuseEnabled {
		s.HighlightRevision = ""
	}
	return s, nil
}

// ResolvePageStyle applies the independent, complete layout and background
// overrides. Omission keeps an existing override. Explicit clear restores its
// book setting, even when other page settings continue to override the book.
func ResolvePageStyle(book PageStyle, prior PageDraft, changes PagePatch) (PagePatch, PageStyle, error) {
	o := PagePatch{Layout: prior.Layout, Background: prior.Background}
	if changes.Layout != nil {
		if changes.ClearLayout || *changes.Layout == (LayoutConfig{}) {
			return PagePatch{}, PageStyle{}, ErrInvalid
		}
		o.Layout = changes.Layout
	} else if changes.ClearLayout {
		o.Layout = nil
	}
	if changes.Background != nil {
		if changes.ClearBackground || *changes.Background == (Background{}) {
			return PagePatch{}, PageStyle{}, ErrInvalid
		}
		o.Background = changes.Background
	} else if changes.ClearBackground {
		o.Background = nil
	}
	if o.Layout != nil {
		layout := o.Layout.canonicalSpacing()
		o.Layout = &layout
		book.Layout = layout
	} else {
		book.Layout = book.Layout.canonicalSpacing()
	}
	if o.Background != nil {
		book.Background = *o.Background
	}
	return o, book, nil
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
	Reuse             bool // Automation retries with the same ordered artifacts share one immutable collection.
	ExpectedRevision  int64
	VersionID         string
	PaginationVersion int64
	PageIDs           []string // Empty selects every page, otherwise uses book order.
	ConfigSnapshot    json.RawMessage
}
