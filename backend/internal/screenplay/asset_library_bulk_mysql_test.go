package screenplay

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestLibraryBulkMySQLRollbackWholeBatchAndClearCAS(t *testing.T) {
	for _, clear := range []bool{false, true} {
		t.Run(map[bool]string{false: "import-second-row-fails", true: "clear-stale"}[clear], func(t *testing.T) {
			db, m, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			store := &MySQLStore{db: db}
			ctx := context.Background()
			m.ExpectBegin()
			m.ExpectExec("INSERT INTO screenplay_asset_library_bulk_owners").WithArgs("alice").WillReturnResult(sqlmock.NewResult(1, 1))
			m.ExpectExec("INSERT INTO screenplay_asset_library_bulk_operations").WithArgs("alice", "request-key", "digest").WillReturnResult(sqlmock.NewResult(1, 1))
			m.ExpectQuery("SELECT digest,result_json.*FOR UPDATE").WithArgs("alice", "request-key").WillReturnRows(sqlmock.NewRows([]string{"digest", "result_json"}).AddRow("digest", nil))
			a := normalizedLibraryFixture()
			b := a
			b.ID = "second-row"
			query := "SELECT .*FROM screenplay_asset_library FORCE INDEX .*owner_username=\\? AND asset_type=\\? FOR UPDATE"
			if clear {
				m.ExpectQuery(query).WithArgs("alice", "character").WillReturnRows(libraryEntryRows(a))
			} else {
				m.ExpectQuery(query).WithArgs("alice", "character").WillReturnRows(sqlmock.NewRows(strings.Split(assetLibraryColumns, ",")))
			}
			action := "import"
			entries := []AssetLibraryEntry{a, b}
			expected := []AssetLibraryExpected(nil)
			if clear {
				action = "clear"
				entries = nil
				expected = []AssetLibraryExpected{{ID: a.ID, Revision: a.Revision + 1}}
			} else {
				m.ExpectExec("INSERT INTO screenplay_asset_library ").WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec("INSERT INTO screenplay_asset_library ").WillReturnError(errors.New("second write fails"))
			}
			m.ExpectRollback()
			_, err = store.ApplyAssetLibraryBulk(ctx, "alice", "request-key", "digest", action, entries, expected)
			if err == nil {
				t.Fatal("expected rollback")
			}
			if clear && !errors.Is(err, ErrRevisionConflict) {
				t.Fatal(err)
			}
			if err = m.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLibraryBulkMySQLPersistResultBeforeCommitAndReplay(t *testing.T) {
	db, m, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := &MySQLStore{db: db}
	ctx := context.Background()
	m.ExpectBegin()
	m.ExpectExec("INSERT INTO screenplay_asset_library_bulk_owners").WithArgs("alice").WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectExec("INSERT INTO screenplay_asset_library_bulk_operations").WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectQuery("SELECT digest,result_json.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"digest", "result_json"}).AddRow("digest", nil))
	m.ExpectQuery("SELECT .*FOR UPDATE").WillReturnRows(sqlmock.NewRows(strings.Split(assetLibraryColumns, ",")))
	m.ExpectExec("UPDATE screenplay_asset_library SET deleted_at").WillReturnResult(sqlmock.NewResult(0, 0))
	m.ExpectExec("UPDATE screenplay_asset_library_bulk_operations SET result_json").WithArgs(`{"entries":[],"cleared_count":0}`, "alice", "clear-key").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
	result, err := store.ApplyAssetLibraryBulk(ctx, "alice", "clear-key", "digest", "clear", nil, []AssetLibraryExpected{})
	if err != nil || result.ClearedCount != 0 {
		t.Fatal(result, err)
	}
	m.ExpectBegin()
	m.ExpectExec("INSERT INTO screenplay_asset_library_bulk_owners").WithArgs("alice").WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectExec("INSERT INTO screenplay_asset_library_bulk_operations").WillReturnResult(sqlmock.NewResult(0, 0))
	m.ExpectQuery("SELECT digest,result_json.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"digest", "result_json"}).AddRow("digest", bulkJSON(result)))
	m.ExpectCommit()
	_, err = store.ApplyAssetLibraryBulk(ctx, "alice", "clear-key", "digest", "clear", nil, []AssetLibraryExpected{})
	if err != nil {
		t.Fatal(err)
	}
	if err = m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
