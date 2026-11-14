package screenplay

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type MemoryAssetLibraryRepository struct {
	mu      sync.Mutex
	entries map[string]AssetLibraryEntry
}

func NewMemoryAssetLibraryRepository() *MemoryAssetLibraryRepository {
	return &MemoryAssetLibraryRepository{entries: map[string]AssetLibraryEntry{}}
}

var _ AssetLibraryRepository = (*MemoryAssetLibraryRepository)(nil)
var _ AssetLibraryRepository = (*MySQLStore)(nil)

func normalizeAssetLibraryEntry(e AssetLibraryEntry) (AssetLibraryEntry, error) {
	e.ID = strings.TrimSpace(e.ID)
	e.Owner = strings.TrimSpace(e.Owner)
	e.Name = strings.TrimSpace(e.Name)
	e.SourceProjectID = strings.TrimSpace(e.SourceProjectID)
	e.SourceAssetID = strings.TrimSpace(e.SourceAssetID)
	e.SourceMediaID = strings.TrimSpace(e.SourceMediaID)
	e.Image.AssetID = strings.TrimSpace(e.Image.AssetID)
	e.Image.URL = ""
	if e.Image.Status == "completed" {
		e.Image.Status = "succeeded"
	}
	if e.ID == "" || e.Owner == "" || e.Name == "" || !validAssetType(e.Type) || e.Image.Kind != "image" || e.Image.Status != "succeeded" || e.Image.AssetID == "" || utf8.RuneCountInString(e.Name) > 191 || utf8.RuneCountInString(e.Category) > 191 || len(e.ID) > 64 || len(e.Owner) > 191 || len(e.SourceProjectID) > 64 || len(e.SourceAssetID) > 64 || len(e.SourceMediaID) > 64 {
		return AssetLibraryEntry{}, ErrInvalidArgument
	}
	return cloneAssetLibraryEntry(e), nil
}
func cloneAssetLibraryEntry(e AssetLibraryEntry) AssetLibraryEntry {
	extra := map[string]string(nil)
	if e.Snapshot.Extra != nil {
		extra = map[string]string{}
		for k, v := range e.Snapshot.Extra {
			extra[k] = v
		}
	}
	e.Snapshot.Extra = extra
	return e
}
func sameAssetLibraryPayload(a, b AssetLibraryEntry) bool {
	a.Revision = 0
	a.CreatedAt = time.Time{}
	a.UpdatedAt = time.Time{}
	b.Revision = 0
	b.CreatedAt = time.Time{}
	b.UpdatedAt = time.Time{}
	rawA, _ := json.Marshal(a)
	rawB, _ := json.Marshal(b)
	return a.Owner == b.Owner && bytes.Equal(rawA, rawB)
}
func validAssetLibrarySearch(owner string, kind AssetType) bool {
	return strings.TrimSpace(owner) != "" && (kind == "" || validAssetType(kind))
}
func (m *MemoryAssetLibraryRepository) ListAssetLibrary(_ context.Context, owner string, kind AssetType, query, category string) ([]AssetLibraryEntry, error) {
	if !validAssetLibrarySearch(owner, kind) {
		return nil, ErrInvalidArgument
	}
	query = strings.ToLower(strings.TrimSpace(query))
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []AssetLibraryEntry{}
	for _, e := range m.entries {
		if e.Owner != owner || (kind != "" && e.Type != kind) || (category != "" && e.Category != category) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(e.Name), query) && !strings.Contains(strings.ToLower(e.Description), query) {
			continue
		}
		out = append(out, cloneAssetLibraryEntry(e))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}
func (m *MemoryAssetLibraryRepository) GetAssetLibraryEntry(_ context.Context, owner, id string) (AssetLibraryEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.entries[id]
	if !ok || e.Owner != owner {
		return AssetLibraryEntry{}, ErrNotFound
	}
	return cloneAssetLibraryEntry(e), nil
}
func (m *MemoryAssetLibraryRepository) SaveAssetLibraryEntry(_ context.Context, input AssetLibraryEntry, expected int64) (AssetLibraryEntry, error) {
	e, err := normalizeAssetLibraryEntry(input)
	if err != nil || expected < 0 {
		if err == nil {
			err = ErrInvalidArgument
		}
		return AssetLibraryEntry{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	old, exists := m.entries[e.ID]
	now := time.Now().UTC()
	if exists {
		if old.Owner != e.Owner {
			return AssetLibraryEntry{}, ErrNotFound
		}
		if sameAssetLibraryPayload(old, e) {
			return cloneAssetLibraryEntry(old), nil
		}
		if expected != old.Revision {
			return AssetLibraryEntry{}, ErrRevisionConflict
		}
		e.CreatedAt = old.CreatedAt
		e.Revision = old.Revision + 1
	} else {
		if expected != 0 {
			return AssetLibraryEntry{}, ErrRevisionConflict
		}
		e.CreatedAt = now
		e.Revision = 1
	}
	e.UpdatedAt = now
	if m.entries == nil {
		m.entries = map[string]AssetLibraryEntry{}
	}
	m.entries[e.ID] = cloneAssetLibraryEntry(e)
	return cloneAssetLibraryEntry(e), nil
}

const assetLibraryColumns = `id,owner_username,asset_type,name,description,category,image_json,snapshot_json,source_project_id,source_asset_id,source_media_id,image_object_json,revision,created_at,updated_at`

func scanAssetLibrary(scanner interface{ Scan(...any) error }) (AssetLibraryEntry, error) {
	var e AssetLibraryEntry
	var image, snapshot, imageObject []byte
	err := scanner.Scan(&e.ID, &e.Owner, &e.Type, &e.Name, &e.Description, &e.Category, &image, &snapshot, &e.SourceProjectID, &e.SourceAssetID, &e.SourceMediaID, &imageObject, &e.Revision, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return e, ErrNotFound
	}
	if err != nil {
		return e, err
	}
	if err = json.Unmarshal(image, &e.Image); err != nil {
		return e, err
	}
	if len(snapshot) > 0 {
		if err = json.Unmarshal(snapshot, &e.Snapshot); err != nil {
			return e, err
		}
	}
	if len(imageObject) > 0 {
		if err = json.Unmarshal(imageObject, &e.ImageObject); err != nil {
			return e, err
		}
	}
	e.Image.URL = ""
	if e.Image.Status == "completed" {
		e.Image.Status = "succeeded"
	}
	return e, nil
}
func (s *MySQLStore) GetAssetLibraryEntry(ctx context.Context, owner, id string) (AssetLibraryEntry, error) {
	return scanAssetLibrary(s.db.QueryRowContext(ctx, `SELECT `+assetLibraryColumns+` FROM screenplay_asset_library WHERE owner_username=? AND id=?`, owner, id))
}
func (s *MySQLStore) ListAssetLibrary(ctx context.Context, owner string, kind AssetType, query, category string) ([]AssetLibraryEntry, error) {
	if !validAssetLibrarySearch(owner, kind) {
		return nil, ErrInvalidArgument
	}
	q := `SELECT ` + assetLibraryColumns + ` FROM screenplay_asset_library WHERE owner_username=?`
	args := []any{owner}
	if kind != "" {
		q += ` AND asset_type=?`
		args = append(args, string(kind))
	}
	if category != "" {
		q += ` AND category=?`
		args = append(args, category)
	}
	if text := strings.ToLower(strings.TrimSpace(query)); text != "" {
		text = strings.NewReplacer("=", "==", "%", "=%", "_", "=_").Replace(text)
		like := "%" + text + "%"
		q += ` AND (LOWER(name) LIKE ? ESCAPE '=' OR LOWER(description) LIKE ? ESCAPE '=')`
		args = append(args, like, like)
	}
	q += ` ORDER BY updated_at DESC,id DESC`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AssetLibraryEntry{}
	for rows.Next() {
		e, err := scanAssetLibrary(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
func (s *MySQLStore) SaveAssetLibraryEntry(ctx context.Context, input AssetLibraryEntry, expected int64) (AssetLibraryEntry, error) {
	e, err := normalizeAssetLibraryEntry(input)
	if err != nil || expected < 0 {
		if err == nil {
			err = ErrInvalidArgument
		}
		return AssetLibraryEntry{}, err
	}
	image, err := json.Marshal(e.Image)
	if err != nil {
		return e, err
	}
	snapshot, err := json.Marshal(e.Snapshot)
	if err != nil {
		return e, err
	}
	imageObject, err := json.Marshal(e.ImageObject)
	if err != nil {
		return e, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return e, err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	// Reserve a new ID without overwriting an existing record. Concurrent equal
	// creates serialize on the primary key and return the same revision.
	if expected == 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO screenplay_asset_library (`+assetLibraryColumns+`) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,1,?,?) ON DUPLICATE KEY UPDATE id=id`, e.ID, e.Owner, string(e.Type), e.Name, e.Description, e.Category, string(image), string(snapshot), e.SourceProjectID, e.SourceAssetID, e.SourceMediaID, string(imageObject), now, now)
		if err != nil {
			return AssetLibraryEntry{}, err
		}
	}
	old, err := scanAssetLibrary(tx.QueryRowContext(ctx, `SELECT `+assetLibraryColumns+` FROM screenplay_asset_library WHERE id=? FOR UPDATE`, e.ID))
	if errors.Is(err, ErrNotFound) {
		return AssetLibraryEntry{}, ErrRevisionConflict
	}
	if err != nil {
		return AssetLibraryEntry{}, err
	}
	if old.Owner != e.Owner {
		return AssetLibraryEntry{}, ErrNotFound
	}
	if sameAssetLibraryPayload(old, e) {
		if err = tx.Commit(); err != nil {
			return AssetLibraryEntry{}, err
		}
		return old, nil
	}
	if old.Revision != expected {
		return AssetLibraryEntry{}, ErrRevisionConflict
	}
	e.CreatedAt = old.CreatedAt
	e.UpdatedAt = now
	e.Revision = old.Revision + 1
	result, err := tx.ExecContext(ctx, `UPDATE screenplay_asset_library SET asset_type=?,name=?,description=?,category=?,image_json=?,snapshot_json=?,source_project_id=?,source_asset_id=?,source_media_id=?,image_object_json=?,revision=?,updated_at=? WHERE id=? AND owner_username=? AND revision=?`, string(e.Type), e.Name, e.Description, e.Category, string(image), string(snapshot), e.SourceProjectID, e.SourceAssetID, e.SourceMediaID, string(imageObject), e.Revision, e.UpdatedAt, e.ID, e.Owner, old.Revision)
	if err != nil {
		return AssetLibraryEntry{}, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return AssetLibraryEntry{}, err
	}
	if rows != 1 {
		return AssetLibraryEntry{}, ErrRevisionConflict
	}
	if err = tx.Commit(); err != nil {
		return AssetLibraryEntry{}, err
	}
	return e, nil
}
