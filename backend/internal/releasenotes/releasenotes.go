// Package releasenotes serves the product-update notes that ship inside the Go
// binary. Notes are keyed by releaseId so one binary can never display another
// release's changelog.
package releasenotes

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed notes/*.json
var embeddedNotes embed.FS

// Item is one changelog entry. Link is optional; the frontend must not render
// the "体验" affordance when it is empty.
type Item struct {
	Title string   `json:"title"`
	Desc  string   `json:"desc"`
	Tags  []string `json:"tags,omitempty"`
	Link  string   `json:"link,omitempty"`
}

// Section groups entries by feature / improvement / fix.
type Section struct {
	Type  string `json:"type"`
	Label string `json:"label"`
	Items []Item `json:"items"`
}

// Note is the payload for GET /api/product-updates/latest.
type Note struct {
	ReleaseID  string    `json:"releaseId"`
	Version    string    `json:"version"`
	ReleasedAt string    `json:"releasedAt"`
	Title      string    `json:"title"`
	Summary    string    `json:"summary,omitempty"`
	Sections   []Section `json:"sections"`
}

// Catalog holds the notes embedded in this binary, keyed by releaseId.
type Catalog struct {
	notes map[string]Note
}

// Default returns the catalog compiled into the binary. A malformed notes
// directory degrades to an empty catalog so the endpoint answers 404 instead of
// failing the process.
func Default() *Catalog {
	catalog, err := Load(embeddedNotes, "notes")
	if err != nil {
		return &Catalog{notes: map[string]Note{}}
	}
	return catalog
}

// Load reads every *.json note under dir. Notes without a releaseId (drafts
// waiting for the next release) or without usable content are ignored.
func Load(fsys fs.FS, dir string) (*Catalog, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("read release notes dir: %w", err)
	}
	catalog := &Catalog{notes: map[string]Note{}}
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}
		raw, readErr := fs.ReadFile(fsys, filepath.ToSlash(filepath.Join(dir, entry.Name())))
		if readErr != nil {
			return nil, fmt.Errorf("read release note %s: %w", entry.Name(), readErr)
		}
		var note Note
		if unmarshalErr := json.Unmarshal(stripBOM(raw), &note); unmarshalErr != nil {
			return nil, fmt.Errorf("parse release note %s: %w", entry.Name(), unmarshalErr)
		}
		note = sanitize(note)
		if !usable(note) {
			continue
		}
		catalog.notes[note.ReleaseID] = note
	}
	return catalog, nil
}

// ForReleaseID returns the notes that belong to the requested release.
func (c *Catalog) ForReleaseID(releaseID string) (Note, bool) {
	if c == nil {
		return Note{}, false
	}
	note, ok := c.notes[strings.TrimSpace(releaseID)]
	return note, ok
}

type releaseManifest struct {
	ReleaseID string `json:"releaseId"`
}

// ReleaseIDFromManifest reads releaseId from a release manifest. Production
// uses runtime-release.json; local review uses backend/current-release.json.
func ReleaseIDFromManifest(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var manifest releaseManifest
	if err := json.Unmarshal(stripBOM(raw), &manifest); err != nil {
		return ""
	}
	return strings.TrimSpace(manifest.ReleaseID)
}

// CurrentReleaseID resolves the release this process belongs to. The runtime
// manifest wins; the tracked local release file is only a fallback for the
// review runtime, whose manifest has no releaseId field.
func CurrentReleaseID(runtimeManifestPath string) string {
	candidates := []string{
		runtimeManifestPath,
		"backend/current-release.json",
		filepath.Join("..", "backend", "current-release.json"),
	}
	for _, candidate := range candidates {
		if id := ReleaseIDFromManifest(candidate); id != "" {
			return id
		}
	}
	return ""
}

func sanitize(note Note) Note {
	note.ReleaseID = strings.TrimSpace(note.ReleaseID)
	note.Version = strings.TrimSpace(note.Version)
	note.ReleasedAt = strings.TrimSpace(note.ReleasedAt)
	note.Title = strings.TrimSpace(note.Title)
	note.Summary = strings.TrimSpace(note.Summary)
	sections := make([]Section, 0, len(note.Sections))
	for _, section := range note.Sections {
		section.Type = strings.TrimSpace(section.Type)
		section.Label = strings.TrimSpace(section.Label)
		if !knownSectionType(section.Type) || section.Label == "" {
			continue
		}
		items := make([]Item, 0, len(section.Items))
		for _, item := range section.Items {
			item.Title = strings.TrimSpace(item.Title)
			item.Desc = strings.TrimSpace(item.Desc)
			item.Link = strings.TrimSpace(item.Link)
			tags := make([]string, 0, len(item.Tags))
			for _, tag := range item.Tags {
				if trimmed := strings.TrimSpace(tag); trimmed != "" {
					tags = append(tags, trimmed)
				}
			}
			item.Tags = tags
			if item.Title == "" {
				continue
			}
			items = append(items, item)
		}
		section.Items = items
		if len(section.Items) == 0 {
			continue
		}
		sections = append(sections, section)
	}
	note.Sections = sections
	return note
}

func usable(note Note) bool {
	return note.ReleaseID != "" && note.Title != "" && len(note.Sections) > 0
}

func knownSectionType(sectionType string) bool {
	switch sectionType {
	case "feature", "improvement", "fix":
		return true
	default:
		return false
	}
}

func stripBOM(raw []byte) []byte {
	return []byte(strings.TrimPrefix(string(raw), "\ufeff"))
}
