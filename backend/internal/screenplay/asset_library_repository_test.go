package screenplay

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"strings"
	"testing"
	"time"
)

func assetLibraryFixture() AssetLibraryEntry {
	return AssetLibraryEntry{ID: "library-a", Owner: "alice", Type: AssetCharacter, Name: " Alice ", Description: "Hero DESCRIPTION", Category: "Hero", Image: MediaItem{ID: "media-1", AssetID: "image-1", Kind: "image", Status: "succeeded", URL: "https://temporary.invalid/signed?secret=token"}, Snapshot: Asset{Name: "source-name", Prompt: "full original", Voice: "voice", Extra: map[string]string{"custom": "retained"}}, SourceProjectID: "source-project", SourceAssetID: "source-asset", SourceMediaID: "source-media", ImageObject: AssetLibraryImageObject{ObjectKey: "asset-library/alice/copy.png", ContentType: "image/png", ByteSize: 123}}
}
func TestAssetLibraryMemorySaveCASNamesAndIsolation(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryAssetLibraryRepository()
	input := assetLibraryFixture()
	saved, err := s.SaveAssetLibraryEntry(ctx, input, 0)
	if err != nil || saved.Revision != 1 || saved.Name != "Alice" || saved.Category != "Hero" || saved.Image.URL != "" {
		t.Fatal(saved, err)
	}
	input.Image.URL = "different short-lived URL"
	repeat, err := s.SaveAssetLibraryEntry(ctx, input, 0)
	if err != nil || repeat.Revision != 1 || !repeat.UpdatedAt.Equal(saved.UpdatedAt) {
		t.Fatal(repeat, err)
	}
	input.Description = "changed"
	if _, err = s.SaveAssetLibraryEntry(ctx, input, 0); !errors.Is(err, ErrRevisionConflict) {
		t.Fatal(err)
	}
	updated, err := s.SaveAssetLibraryEntry(ctx, input, 1)
	if err != nil || updated.Revision != 2 || !updated.CreatedAt.Equal(saved.CreatedAt) {
		t.Fatal(updated, err)
	}
	if _, err = s.SaveAssetLibraryEntry(ctx, input, 0); err != nil {
		t.Fatal("identical stale replay must return saved result", err)
	}
	second := input
	second.ID = "library-b"
	if _, err = s.SaveAssetLibraryEntry(ctx, second, 0); err != nil {
		t.Fatal("same name must allow a separate record", err)
	}
	all, _ := s.ListAssetLibrary(ctx, "alice", AssetCharacter, "", "")
	if len(all) != 2 {
		t.Fatal(all)
	}
	if _, err = s.GetAssetLibraryEntry(ctx, "bob", saved.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	foreign := input
	foreign.Owner = "bob"
	if _, err = s.SaveAssetLibraryEntry(ctx, foreign, 2); !errors.Is(err, ErrNotFound) {
		t.Fatal("foreign id overwrite", err)
	}
	read, _ := s.GetAssetLibraryEntry(ctx, "alice", saved.ID)
	read.Snapshot.Extra["custom"] = "mutated"
	again, _ := s.GetAssetLibraryEntry(ctx, "alice", saved.ID)
	if again.Snapshot.Extra["custom"] != "retained" {
		t.Fatal("stored snapshot alias")
	}
}
func TestAssetLibraryMemoryFiltersAndValidation(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryAssetLibraryRepository()
	for i, kind := range []AssetType{AssetCharacter, AssetScene, AssetProp} {
		e := assetLibraryFixture()
		e.ID = fmt.Sprint(i)
		e.Type = kind
		if _, err := s.SaveAssetLibraryEntry(ctx, e, 0); err != nil {
			t.Fatal(err)
		}
	}
	cases := []struct {
		owner           string
		kind            AssetType
		query, category string
		count           int
	}{{"alice", AssetCharacter, " descRIPtion ", "Hero", 1}, {"alice", "", "ALICE", "Hero", 3}, {"alice", AssetCharacter, "", "hero", 0}, {"alice", AssetScene, "missing", "", 0}, {"bob", "", "", "", 0}, {"alice", "", "%", "", 0}}
	for _, test := range cases {
		rows, err := s.ListAssetLibrary(ctx, test.owner, test.kind, test.query, test.category)
		if err != nil || len(rows) != test.count {
			t.Fatal(test, rows, err)
		}
	}
	for _, mutate := range []func(*AssetLibraryEntry){func(e *AssetLibraryEntry) { e.Name = " " }, func(e *AssetLibraryEntry) { e.Type = "avatar" }, func(e *AssetLibraryEntry) { e.Image.Kind = "video" }, func(e *AssetLibraryEntry) { e.Image.Status = "running" }, func(e *AssetLibraryEntry) { e.Image.AssetID = "" }, func(e *AssetLibraryEntry) { e.Owner = "" }} {
		e := assetLibraryFixture()
		mutate(&e)
		if _, err := s.SaveAssetLibraryEntry(ctx, e, 0); !errors.Is(err, ErrInvalidArgument) {
			t.Fatal(e, err)
		}
	}
	if _, err := s.ListAssetLibrary(ctx, "alice", "avatar", "", ""); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
	if _, err := s.SaveAssetLibraryEntry(ctx, assetLibraryFixture(), 5); !errors.Is(err, ErrRevisionConflict) {
		t.Fatal(err)
	}
}
func libraryEntryRows(e AssetLibraryEntry) *sqlmock.Rows {
	image, _ := json.Marshal(e.Image)
	snapshot, _ := json.Marshal(e.Snapshot)
	imageObject, _ := json.Marshal(e.ImageObject)
	return sqlmock.NewRows(strings.Split(assetLibraryColumns, ",")).AddRow(e.ID, e.Owner, string(e.Type), e.Name, e.Description, e.Category, string(image), string(snapshot), e.SourceProjectID, e.SourceAssetID, e.SourceMediaID, string(imageObject), e.Revision, e.CreatedAt, e.UpdatedAt, e.DeletedAt)
}

type libraryImageMatcher struct{}

func (libraryImageMatcher) Match(v driver.Value) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	var image MediaItem
	return json.Unmarshal([]byte(s), &image) == nil && image.URL == "" && image.AssetID == "image-1" && image.Status == "succeeded"
}
func normalizedLibraryFixture() AssetLibraryEntry {
	e, _ := normalizeAssetLibraryEntry(assetLibraryFixture())
	e.Revision = 1
	e.CreatedAt = time.Now().UTC()
	e.UpdatedAt = e.CreatedAt
	return e
}
func TestAssetLibraryMySQLIdempotentCreateAndRead(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	s := NewMySQLStore(db)
	e := normalizedLibraryFixture()
	m.ExpectBegin()
	m.ExpectExec("INSERT INTO screenplay_asset_library.*ON DUPLICATE KEY UPDATE id=id").WithArgs(e.ID, e.Owner, string(e.Type), e.Name, e.Description, e.Category, libraryImageMatcher{}, sqlmock.AnyArg(), e.SourceProjectID, e.SourceAssetID, e.SourceMediaID, libraryImageObjectMatcher{}, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectQuery("SELECT .* FROM screenplay_asset_library WHERE id=\\? FOR UPDATE").WithArgs(e.ID).WillReturnRows(libraryEntryRows(e))
	m.ExpectCommit()
	got, err := s.SaveAssetLibraryEntry(context.Background(), assetLibraryFixture(), 0)
	if err != nil || got.Revision != 1 || got.Image.URL != "" {
		t.Fatal(got, err)
	}
	m.ExpectQuery("SELECT .* FROM screenplay_asset_library WHERE owner_username=\\? AND id=\\?").WithArgs("alice", e.ID).WillReturnRows(libraryEntryRows(e))
	got, err = s.GetAssetLibraryEntry(context.Background(), "alice", e.ID)
	if err != nil || got.Snapshot.Extra["custom"] != "retained" || got.SourceMediaID != e.SourceMediaID || got.ImageObject != e.ImageObject {
		t.Fatal(got, err)
	}
	m.ExpectQuery("SELECT .* FROM screenplay_asset_library WHERE owner_username=\\? AND id=\\?").WithArgs("bob", e.ID).WillReturnError(sql.ErrNoRows)
	if _, err = s.GetAssetLibraryEntry(context.Background(), "bob", e.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err = m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestAssetLibraryMySQLCASUpdateAndRollback(t *testing.T) {
	for _, outcome := range []string{"success", "stale", "foreign", "write-failure", "lost-cas"} {
		t.Run(outcome, func(t *testing.T) {
			db, m, _ := sqlmock.New()
			defer db.Close()
			e := normalizedLibraryFixture()
			input := e
			input.Description = "changed"
			expected := int64(1)
			if outcome == "stale" {
				expected = 5
			}
			if outcome == "foreign" {
				input.Owner = "bob"
			}
			m.ExpectBegin()
			m.ExpectQuery("SELECT .* FROM screenplay_asset_library WHERE id=\\? FOR UPDATE").WithArgs(e.ID).WillReturnRows(libraryEntryRows(e))
			if outcome != "stale" && outcome != "foreign" {
				q := m.ExpectExec("UPDATE screenplay_asset_library SET").WithArgs(string(e.Type), e.Name, "changed", e.Category, libraryImageMatcher{}, sqlmock.AnyArg(), e.SourceProjectID, e.SourceAssetID, e.SourceMediaID, libraryImageObjectMatcher{}, int64(2), sqlmock.AnyArg(), e.ID, e.Owner, int64(1))
				switch outcome {
				case "write-failure":
					q.WillReturnError(fmt.Errorf("disk failed"))
				case "lost-cas":
					q.WillReturnResult(sqlmock.NewResult(0, 0))
				default:
					q.WillReturnResult(sqlmock.NewResult(0, 1))
				}
			}
			if outcome == "success" {
				m.ExpectCommit()
			} else {
				m.ExpectRollback()
			}
			got, err := NewMySQLStore(db).SaveAssetLibraryEntry(context.Background(), input, expected)
			switch outcome {
			case "success":
				if err != nil || got.Revision != 2 || !got.CreatedAt.Equal(e.CreatedAt) {
					t.Fatal(got, err)
				}
			case "foreign":
				if !errors.Is(err, ErrNotFound) {
					t.Fatal(err)
				}
			case "stale", "lost-cas":
				if !errors.Is(err, ErrRevisionConflict) {
					t.Fatal(err)
				}
			case "write-failure":
				if err == nil {
					t.Fatal("write error swallowed")
				}
			}
			if err = m.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestAssetLibraryMySQLSearchAndInvalidInput(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	s := NewMySQLStore(db)
	e := normalizedLibraryFixture()
	m.ExpectQuery("SELECT .* FROM screenplay_asset_library WHERE owner_username=\\? AND asset_type=\\? AND category=\\? AND \\(LOWER\\(name\\) LIKE").WithArgs("alice", "character", "Hero", "%a=%=_==%", "%a=%=_==%").WillReturnRows(libraryEntryRows(e))
	rows, err := s.ListAssetLibrary(context.Background(), "alice", AssetCharacter, " A%_= ", "Hero")
	if err != nil || len(rows) != 1 {
		t.Fatal(rows, err)
	}
	invalid := assetLibraryFixture()
	invalid.Image.Status = "running"
	if _, err = s.SaveAssetLibraryEntry(context.Background(), invalid, 0); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
	if err = m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAssetLibraryAcceptsCMGeneratedAndUploadedSuccessfulMedia(t *testing.T) {
	// saveBytes and the image-generation completion path create exactly this
	// succeeded MediaItem shape, not a synthetic library-only success state.
	item := MediaItem{ID: "sp-media-real", AssetID: "sp-media-real", Kind: "image", Status: "succeeded", TargetID: "source-asset", CreatedAt: time.Now().UTC()}
	ctx := context.Background()
	repo := NewMemoryAssetLibraryRepository()
	entry := assetLibraryFixture()
	entry.Image = item
	saved, err := repo.SaveAssetLibraryEntry(ctx, entry, 0)
	if err != nil || saved.Image.Status != "succeeded" || saved.Image.ID != item.ID || saved.Image.AssetID != item.AssetID {
		t.Fatal(saved, err)
	}
	legacy := entry
	legacy.Image.Status = "completed"
	again, err := repo.SaveAssetLibraryEntry(ctx, legacy, 0)
	if err != nil || again.Revision != saved.Revision || again.Image.Status != "succeeded" {
		t.Fatal("compatible completed state must normalize without a new revision", again, err)
	}
}

type libraryImageObjectMatcher struct{}

func (libraryImageObjectMatcher) Match(v driver.Value) bool {
	raw, ok := v.(string)
	if !ok {
		return false
	}
	var obj AssetLibraryImageObject
	return json.Unmarshal([]byte(raw), &obj) == nil && obj.ObjectKey == "asset-library/alice/copy.png" && obj.ContentType == "image/png" && obj.ByteSize == 123
}

func TestAssetLibraryDeleteOwnerCASAndSQLRollback(t *testing.T) {
	ctx := context.Background()
	memory := NewMemoryAssetLibraryRepository()
	saved, err := memory.SaveAssetLibraryEntry(ctx, assetLibraryFixture(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if err = memory.DeleteAssetLibraryEntry(ctx, "bob", saved.ID, saved.Revision); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err = memory.DeleteAssetLibraryEntry(ctx, "alice", saved.ID, saved.Revision+1); !errors.Is(err, ErrRevisionConflict) {
		t.Fatal(err)
	}
	if err = memory.DeleteAssetLibraryEntry(ctx, "alice", saved.ID, saved.Revision); err != nil {
		t.Fatal(err)
	}
	if _, err = memory.GetAssetLibraryEntry(ctx, "alice", saved.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	for _, outcome := range []string{"success", "stale", "write-error", "foreign"} {
		t.Run(outcome, func(t *testing.T) {
			db, m, _ := sqlmock.New()
			defer db.Close()
			owner := "alice"
			if outcome == "foreign" {
				owner = "bob"
			}
			m.ExpectBegin()
			q := m.ExpectQuery("SELECT revision FROM screenplay_asset_library.*FOR UPDATE").WithArgs(owner, saved.ID)
			if outcome == "foreign" {
				q.WillReturnError(sql.ErrNoRows)
			} else {
				q.WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(1))
			}
			expected := int64(1)
			if outcome == "stale" {
				expected = 2
			}
			if outcome == "success" || outcome == "write-error" {
				exec := m.ExpectExec("UPDATE screenplay_asset_library SET deleted_at").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), owner, saved.ID, expected)
				if outcome == "write-error" {
					exec.WillReturnError(fmt.Errorf("write failed"))
				} else {
					exec.WillReturnResult(sqlmock.NewResult(0, 1))
				}
			}
			if outcome == "success" {
				m.ExpectCommit()
			} else {
				m.ExpectRollback()
			}
			err := NewMySQLStore(db).DeleteAssetLibraryEntry(ctx, owner, saved.ID, expected)
			if outcome == "success" && err != nil || outcome != "success" && err == nil {
				t.Fatal(outcome, err)
			}
			if err = m.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAssetLibraryTombstonePreventsDeletedIDAndRevisionReuse(t *testing.T) {
	ctx := context.Background()
	memory := NewMemoryAssetLibraryRepository()
	input := assetLibraryFixture()
	saved, err := memory.SaveAssetLibraryEntry(ctx, input, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err = memory.DeleteAssetLibraryEntry(ctx, "alice", saved.ID, saved.Revision); err != nil {
		t.Fatal(err)
	}
	if _, err = memory.GetAssetLibraryEntry(ctx, "alice", saved.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	rows, _ := memory.ListAssetLibrary(ctx, "alice", "", "", "")
	if len(rows) != 0 {
		t.Fatal("deleted entry is listed", rows)
	}
	for _, revision := range []int64{0, saved.Revision} {
		if _, err = memory.SaveAssetLibraryEntry(ctx, input, revision); !errors.Is(err, ErrInvalidState) {
			t.Fatal("deleted ID reused", revision, err)
		}
	}
	input.Image.AssetID = "replacement-image"
	if _, err = memory.SaveAssetLibraryEntry(ctx, input, 0); !errors.Is(err, ErrInvalidState) {
		t.Fatal("new image reused deleted revision", err)
	}
	tombstone := normalizedLibraryFixture()
	now := time.Now().UTC()
	tombstone.DeletedAt = &now
	tombstone.Revision = 2
	db, m, _ := sqlmock.New()
	defer db.Close()
	m.ExpectBegin()
	m.ExpectExec("INSERT INTO screenplay_asset_library.*ON DUPLICATE KEY UPDATE id=id").WillReturnResult(sqlmock.NewResult(0, 0))
	m.ExpectQuery("SELECT .* FROM screenplay_asset_library WHERE id=.*FOR UPDATE").WithArgs(tombstone.ID).WillReturnRows(libraryEntryRows(tombstone))
	m.ExpectRollback()
	if _, err = NewMySQLStore(db).SaveAssetLibraryEntry(ctx, assetLibraryFixture(), 0); !errors.Is(err, ErrInvalidState) {
		t.Fatal("SQL resurrected tombstone", err)
	}
	m.ExpectQuery("SELECT .* FROM screenplay_asset_library WHERE owner_username=.*AND id=.*AND deleted_at IS NULL").WithArgs("alice", tombstone.ID).WillReturnError(sql.ErrNoRows)
	if _, err = NewMySQLStore(db).GetAssetLibraryEntry(ctx, "alice", tombstone.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err = m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
