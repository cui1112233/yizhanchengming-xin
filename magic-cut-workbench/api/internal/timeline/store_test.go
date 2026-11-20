package timeline

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPutRejectsOutdatedRevisionWithCurrentValue(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(selectLatestRevisionForUpdateSQL)).WithArgs("alice", "mcp_1").WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(3))
	mock.ExpectRollback()
	_, err = NewStore(db, time.Now).Put(context.Background(), "alice", "mcp_1", Timeline{Revision: 2, Tracks: []Track{}}, 2)
	var conflict *RevisionConflictError
	if !errors.As(err, &conflict) || conflict.CurrentRevision != 3 {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPutCreatesTheNextTimelineRevision(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(selectLatestRevisionForUpdateSQL)).WithArgs("alice", "mcp_1").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta(insertTimelineSQL)).WithArgs("mcp_1", "alice", 1, sqlmock.AnyArg(), now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	saved, err := NewStore(db, func() time.Time { return now }).Put(context.Background(), "alice", "mcp_1", Timeline{Tracks: []Track{}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Revision != 1 {
		t.Fatalf("revision=%d", saved.Revision)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateTimelineRejectsInvalidClipBounds(t *testing.T) {
	err := Validate(Timeline{Tracks: []Track{{ID: "v1", Kind: "video", Clips: []Clip{{ID: "c1", MaterialID: "m1", StartMs: -1, DurationMs: 0, Speed: 1}}}}})
	if err == nil {
		t.Fatal("expected invalid timeline")
	}
}
