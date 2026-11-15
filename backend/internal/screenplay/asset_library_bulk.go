package screenplay

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const LibraryArchiveJSON = "library_archive_json"
const LibraryArchiveID = "library_archive_id"
const LibraryMetadataOnly = "library_archive_metadata_only"

// ValidateAssetLibraryBulkEntries validates the entire archive before media writes.
func ValidateAssetLibraryBulkEntries(entries []AssetLibraryEntry) error {
	for _, entry := range entries {
		if entry.Type != AssetCharacter {
			return ErrInvalidArgument
		}
		if _, err := normalizeAssetLibraryEntry(entry); err != nil {
			return err
		}
	}
	return nil
}

type AssetLibraryExpected struct {
	ID       string `json:"id"`
	Revision int64  `json:"revision"`
}
type AssetLibraryBulkResult struct {
	Entries      []AssetLibraryEntry `json:"entries"`
	ClearedCount int                 `json:"cleared_count"`
}
type assetLibraryBulkRecord struct {
	Digest    string
	Result    AssetLibraryBulkResult
	Completed bool
}

// Optional capability preserves existing repository substitutes.
type AssetLibraryBulkRepository interface {
	ReserveAssetLibraryBulk(context.Context, string, string, string) error
	LookupAssetLibraryBulk(context.Context, string, string, string) (AssetLibraryBulkResult, error)
	ApplyAssetLibraryBulk(context.Context, string, string, string, string, []AssetLibraryEntry, []AssetLibraryExpected) (AssetLibraryBulkResult, error)
}

func (m *MemoryAssetLibraryRepository) ReserveAssetLibraryBulk(ctx context.Context, owner, key, digest string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if owner == "" || key == "" || len(key) > 191 || digest == "" {
		return ErrInvalidArgument
	}
	if m.bulkOperations == nil {
		m.bulkOperations = map[string]assetLibraryBulkRecord{}
	}
	opKey := owner + "\x00" + key
	if old, ok := m.bulkOperations[opKey]; ok {
		if old.Digest != digest {
			return ErrRevisionConflict
		}
		return nil
	}
	m.bulkOperations[opKey] = assetLibraryBulkRecord{Digest: digest}
	return nil
}

func (m *MemoryAssetLibraryRepository) LookupAssetLibraryBulk(ctx context.Context, owner, key, digest string) (AssetLibraryBulkResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return AssetLibraryBulkResult{}, err
	}
	old, ok := m.bulkOperations[owner+"\x00"+key]
	if !ok {
		return AssetLibraryBulkResult{}, ErrNotFound
	}
	if old.Digest != digest {
		return AssetLibraryBulkResult{}, ErrRevisionConflict
	}
	if !old.Completed {
		return AssetLibraryBulkResult{}, ErrNotFound
	}
	return cloneBulkResult(old.Result), nil
}

func cloneBulkResult(r AssetLibraryBulkResult) AssetLibraryBulkResult {
	out := AssetLibraryBulkResult{ClearedCount: r.ClearedCount, Entries: []AssetLibraryEntry{}}
	for _, e := range r.Entries {
		out.Entries = append(out.Entries, cloneAssetLibraryEntry(e))
	}
	return out
}

func prepareLibraryBulk(owner, key, digest, action string, entries, current []AssetLibraryEntry, expected []AssetLibraryExpected) ([]AssetLibraryEntry, error) {
	if strings.TrimSpace(owner) == "" || key == "" || len(key) > 191 || digest == "" {
		return nil, ErrInvalidArgument
	}
	if action == "clear" {
		if len(current) != len(expected) {
			return nil, ErrRevisionConflict
		}
		versions := map[string]int64{}
		for _, e := range expected {
			if e.ID == "" || e.Revision < 1 || versions[e.ID] != 0 {
				return nil, ErrInvalidArgument
			}
			versions[e.ID] = e.Revision
		}
		for _, e := range current {
			if versions[e.ID] != e.Revision {
				return nil, ErrRevisionConflict
			}
		}
		return nil, nil
	}
	if action != "import" || len(entries) == 0 {
		return nil, ErrInvalidArgument
	}
	used := map[string]bool{}
	ids := map[string]bool{}
	for _, e := range current {
		logical := e.Snapshot.Extra[LibraryArchiveID]
		if logical == "" {
			logical = e.ID
		}
		used[logical] = true
		ids[e.ID] = true
	}
	out := make([]AssetLibraryEntry, 0, len(entries))
	for i, e := range entries {
		if e.Owner != owner || e.Type != AssetCharacter || ids[e.ID] {
			return nil, ErrRevisionConflict
		}
		ids[e.ID] = true
		e = cloneAssetLibraryEntry(e)
		if e.Snapshot.Extra == nil {
			e.Snapshot.Extra = map[string]string{}
		}
		logical := e.Snapshot.Extra[LibraryArchiveID]
		if logical == "" || used[logical] {
			sum := sha256.Sum256([]byte(owner + "\x00" + key + "\x00" + fmt.Sprint(i)))
			logical = fmt.Sprintf("cm-%x", sum[:16])
			for used[logical] {
				logical += "x"
			}
		}
		used[logical] = true
		e.Snapshot.Extra[LibraryArchiveID] = logical
		normalized, err := normalizeAssetLibraryEntry(e)
		if err != nil {
			return nil, err
		}
		out = append(out, normalized)
	}
	return out, nil
}
func (m *MemoryAssetLibraryRepository) ApplyAssetLibraryBulk(ctx context.Context, owner, key, digest, action string, entries []AssetLibraryEntry, expected []AssetLibraryExpected) (AssetLibraryBulkResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	empty := AssetLibraryBulkResult{}
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	opKey := owner + "\x00" + key
	if old, ok := m.bulkOperations[opKey]; ok {
		if old.Digest != digest {
			return empty, ErrRevisionConflict
		}
		if old.Completed {
			return cloneBulkResult(old.Result), nil
		}
	}
	current := []AssetLibraryEntry{}
	for _, e := range m.entries {
		if e.Owner == owner && e.Type == AssetCharacter && e.DeletedAt == nil {
			current = append(current, e)
		}
	}
	prepared, err := prepareLibraryBulk(owner, key, digest, action, entries, current, expected)
	if err != nil {
		return empty, err
	}
	for _, e := range prepared {
		if _, exists := m.entries[e.ID]; exists {
			return empty, ErrInvalidState
		}
	}
	now := time.Now().UTC()
	result := AssetLibraryBulkResult{Entries: []AssetLibraryEntry{}}
	if action == "clear" {
		for _, e := range current {
			e.DeletedAt = &now
			e.UpdatedAt = now
			e.Revision++
			m.entries[e.ID] = cloneAssetLibraryEntry(e)
		}
		result.ClearedCount = len(current)
	} else {
		for _, e := range prepared {
			e.Revision = 1
			e.CreatedAt = now
			e.UpdatedAt = now
			m.entries[e.ID] = cloneAssetLibraryEntry(e)
			result.Entries = append(result.Entries, e)
		}
	}
	if m.bulkOperations == nil {
		m.bulkOperations = map[string]assetLibraryBulkRecord{}
	}
	m.bulkOperations[opKey] = assetLibraryBulkRecord{Digest: digest, Result: cloneBulkResult(result), Completed: true}
	return cloneBulkResult(result), nil
}
func bulkJSON(v any) string { raw, _ := json.Marshal(v); return string(raw) }
