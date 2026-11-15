package screenplay

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"strings"
	"testing"
)

func TestAssetLibraryCategoriesMemory(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryAssetLibraryRepository()
	state, e := s.LoadAssetLibraryCategories(ctx, "alice")
	if e != nil || state.Revision != 0 || state.CustomCategories == nil {
		t.Fatal(state, e)
	}
	entry := assetLibraryFixture()
	entry.Category = " 自创 "
	if _, e = s.SaveAssetLibraryEntry(ctx, entry, 0); e != nil {
		t.Fatal(e)
	}
	state, e = s.ChangeAssetLibraryCategory(ctx, "alice", "add", " 自创 ", 0)
	if e != nil || state.Revision != 1 || state.CustomCategories[0] != " 自创 " {
		t.Fatal(state, e)
	}
	state.CustomCategories[0] = "mutated"
	state, e = s.ChangeAssetLibraryCategory(ctx, "alice", "add", " 自创 ", 0)
	if e != nil || state.Revision != 1 || state.CustomCategories[0] != " 自创 " {
		t.Fatal(state, e)
	}
	if _, e = s.ChangeAssetLibraryCategory(ctx, "alice", "add", "different", 0); !errors.Is(e, ErrRevisionConflict) {
		t.Fatal(e)
	}
	state, e = s.ChangeAssetLibraryCategory(ctx, "alice", "add", "自创", 1)
	if e != nil || len(state.CustomCategories) != 2 {
		t.Fatal(state, e)
	}
	state, e = s.ChangeAssetLibraryCategory(ctx, "alice", "delete", " 自创 ", 2)
	if e != nil || state.Revision != 3 {
		t.Fatal(state, e)
	}
	state, e = s.ChangeAssetLibraryCategory(ctx, "alice", "delete", " 自创 ", 2)
	if e != nil || state.Revision != 3 {
		t.Fatal(state, e)
	}
	saved, e := s.GetAssetLibraryEntry(ctx, "alice", entry.ID)
	if e != nil || saved.Category != " 自创 " {
		t.Fatal(saved, e)
	}
	bob, e := s.LoadAssetLibraryCategories(ctx, "bob")
	if e != nil || bob.Revision != 0 || len(bob.CustomCategories) != 0 {
		t.Fatal(bob, e)
	}
	for _, name := range []string{"", "  ", "a\nb", "a\x00b", strings.Repeat("界", 192)} {
		if _, e = s.ChangeAssetLibraryCategory(ctx, "alice", "add", name, 3); !errors.Is(e, ErrInvalidArgument) {
			t.Fatalf("%q: %v", name, e)
		}
	}
	for _, name := range BuiltinAssetLibraryCategories() {
		if _, e = s.ChangeAssetLibraryCategory(ctx, "alice", "delete", name, 3); !errors.Is(e, ErrInvalidArgument) {
			t.Fatal(name, e)
		}
		state, e = s.ChangeAssetLibraryCategory(ctx, "alice", "add", name, 3)
		if e != nil || state.Revision != 3 || len(state.CustomCategories) != 1 {
			t.Fatal(state, e)
		}
	}
	if _, e = s.ChangeAssetLibraryCategory(ctx, "alice", "add", strings.Repeat("界", 191), 3); e != nil {
		t.Fatal(e)
	}
}
func TestAssetLibraryCategoriesMySQLTransactionAndFreshRead(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	ctx := context.Background()
	s := NewMySQLStore(db)
	m.ExpectBegin()
	m.ExpectExec("INSERT IGNORE INTO screenplay_asset_library_categories").WithArgs("alice").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectQuery("SELECT state_json FROM screenplay_asset_library_categories.*FOR UPDATE").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"state_json"}).AddRow(`{"custom_categories":[],"revision":0}`))
	raw, _ := json.Marshal(AssetLibraryCategoryState{CustomCategories: []string{" 自创 "}, Revision: 1})
	m.ExpectExec("UPDATE screenplay_asset_library_categories SET state_json").WithArgs(string(raw), "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
	state, e := s.ChangeAssetLibraryCategory(ctx, "alice", "add", " 自创 ", 0)
	if e != nil || state.Revision != 1 {
		t.Fatal(state, e)
	}
	m.ExpectQuery("SELECT state_json FROM screenplay_asset_library_categories WHERE owner_username").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"state_json"}).AddRow(raw))
	state, e = NewMySQLStore(db).LoadAssetLibraryCategories(ctx, "alice")
	if e != nil || state.CustomCategories[0] != " 自创 " {
		t.Fatal(state, e)
	}
	m.ExpectBegin()
	m.ExpectExec("INSERT IGNORE INTO screenplay_asset_library_categories").WithArgs("alice").WillReturnResult(sqlmock.NewResult(0, 0))
	m.ExpectQuery("SELECT state_json FROM screenplay_asset_library_categories.*FOR UPDATE").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"state_json"}).AddRow(raw))
	m.ExpectRollback()
	if _, e = s.ChangeAssetLibraryCategory(ctx, "alice", "add", "other", 0); !errors.Is(e, ErrRevisionConflict) {
		t.Fatal(e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}

func TestAssetLibraryCategoriesConcurrentDistinctChanges(t *testing.T) {
	s := NewMemoryAssetLibraryRepository()
	ctx := context.Background()
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, name := range []string{"one", "two"} {
		go func(name string) {
			<-start
			_, e := s.ChangeAssetLibraryCategory(ctx, "alice", "add", name, 0)
			results <- e
		}(name)
	}
	close(start)
	success, conflict := 0, 0
	for i := 0; i < 2; i++ {
		e := <-results
		if e == nil {
			success++
		} else if errors.Is(e, ErrRevisionConflict) {
			conflict++
		} else {
			t.Fatal(e)
		}
	}
	state, e := s.LoadAssetLibraryCategories(ctx, "alice")
	if e != nil || success != 1 || conflict != 1 || state.Revision != 1 || len(state.CustomCategories) != 1 {
		t.Fatal(state, e, success, conflict)
	}
}
func TestAssetLibraryCategoriesMySQLErrorDoesNotConfirmSave(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	s := NewMySQLStore(db)
	m.ExpectQuery("SELECT state_json FROM screenplay_asset_library_categories WHERE owner_username").WithArgs("alice").WillReturnError(errors.New("read unavailable"))
	if _, e = s.LoadAssetLibraryCategories(context.Background(), "alice"); e == nil {
		t.Fatal("read failure hidden")
	}
	m.ExpectBegin()
	m.ExpectExec("INSERT IGNORE INTO screenplay_asset_library_categories").WithArgs("alice").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectQuery("SELECT state_json FROM screenplay_asset_library_categories.*FOR UPDATE").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"state_json"}).AddRow(`{"custom_categories":[],"revision":0}`))
	m.ExpectExec("UPDATE screenplay_asset_library_categories SET state_json").WillReturnError(errors.New("write unavailable"))
	m.ExpectRollback()
	if state, e := s.ChangeAssetLibraryCategory(context.Background(), "alice", "add", "x", 0); e == nil || state.Revision != 0 {
		t.Fatal(state, e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
