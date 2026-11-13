package screenplay

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
)

func TestMySQLAuthoringPersistsAcrossRepositoriesAndRejectsConflict(t *testing.T) {
	ctx := context.Background()
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	repo := NewMySQLStore(db).AuthoringRepository()
	st := AuthoringState{Owner: "alice", ProjectID: "p", Chapters: []Chapter{{ID: "c", Title: "章", ScriptText: "持久结果"}}}
	m.ExpectExec("INSERT IGNORE INTO screenplay_authoring").WithArgs("p", "alice", int64(1), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	st, e = repo.Save(ctx, st, 0)
	if e != nil {
		t.Fatal(e)
	}
	b, _ := json.Marshal(st)
	m.ExpectQuery("SELECT body,revision FROM screenplay_authoring").WithArgs("alice", "p").WillReturnRows(sqlmock.NewRows([]string{"body", "revision"}).AddRow(b, 1))
	other := NewMySQLStore(db).AuthoringRepository()
	got, e := other.Load(ctx, "alice", "p")
	if e != nil || got.Chapters[0].ScriptText != "持久结果" || got.Revision != 1 {
		t.Fatal(got, e)
	}
	m.ExpectExec("UPDATE screenplay_authoring").WithArgs(sqlmock.AnyArg(), int64(2), "p", "alice", int64(1)).WillReturnResult(sqlmock.NewResult(0, 0))
	if _, e = other.Save(ctx, got, 1); !errors.Is(e, ErrRevisionConflict) {
		t.Fatal(e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestMySQLAuthoringStoryboardCASRejectsBeforeShotDelete(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	m.ExpectBegin()
	m.ExpectQuery("SELECT id, revision FROM screenplay_storyboards").WithArgs("p", "alice", 1, string(EngineXiaoluoV7)).WillReturnRows(sqlmock.NewRows([]string{"id", "revision"}).AddRow("b", 4))
	m.ExpectRollback()
	_, e = NewMySQLStore(db).SaveStoryboardRevision(context.Background(), Storyboard{ProjectID: "p", OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoluoV7}, 3)
	if !errors.Is(e, ErrRevisionConflict) {
		t.Fatal(e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
