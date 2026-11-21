package novelfetchworkshop

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestNovelFetchTaskRetentionExpiryUsesLocalMidnight(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)
	last := time.Date(2026, 11, 14, 11, 7, 0, 0, loc)
	got := novelFetchTaskRetentionExpiry(last, 7, loc)
	want := time.Date(2026, 11, 21, 0, 0, 0, 0, loc).UTC()
	if !got.Equal(want) {
		t.Fatalf("expiry=%s want=%s", got.Format(time.RFC3339Nano), want.Format(time.RFC3339Nano))
	}
}

func TestMySQLStoreCleanupExpiredTasksDeletesOnlyExpiredTerminalTasks(t *testing.T) {
	_, mock, store := newBodySQLMock(t)
	ctx := context.Background()
	now := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	expiredUpdated := time.Date(2026, 11, 14, 3, 0, 0, 0, time.UTC)
	notExpiredUpdated := time.Date(2026, 11, 15, 9, 0, 0, 0, time.UTC)

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM novel_fetch_workshop_runs WHERE owner_username=? AND updated_at<=? AND status NOT IN ('queued','running')`)).
		WithArgs("alice", now.Add(-7*24*time.Hour)).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT book_id, updated_at FROM novel_fetch_workshop_documents WHERE owner_username=? AND updated_at<=? AND `+novelFetchTerminalStatusSQL+` ORDER BY updated_at ASC LIMIT ?`)).
		WithArgs("alice", now.Add(-6*24*time.Hour), 200).
		WillReturnRows(sqlmock.NewRows([]string{"book_id", "updated_at"}).
			AddRow("book-old", expiredUpdated).
			AddRow("book-fresh", notExpiredUpdated))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE novel_fetch_workshop_bodies SET state='releasable', releasable_at=?, expires_at=?, updated_at=CURRENT_TIMESTAMP(6) WHERE owner_username=? AND book_id=? AND state<>'releasable'`)).
		WithArgs(now, now.Add(DefaultBodyRetentionDays*24*time.Hour), "alice", "book-old").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM novel_fetch_workshop_documents WHERE owner_username=? AND book_id=?`)).
		WithArgs("alice", "book-old").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := store.CleanupExpiredTasks(ctx, "alice", now, 7, 200)
	if err != nil {
		t.Fatal(err)
	}
	if result.Deleted != 1 || result.RunsDeleted != 3 || len(result.Results) != 1 || result.Results[0].BookID != "book-old" {
		t.Fatalf("result=%+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreListDocumentOwners(t *testing.T) {
	_, mock, store := newBodySQLMock(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT owner_username FROM novel_fetch_workshop_documents ORDER BY owner_username ASC`)).
		WillReturnRows(sqlmock.NewRows([]string{"owner_username"}).AddRow("alice").AddRow("bob"))
	owners, err := store.ListDocumentOwners(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(owners) != 2 || owners[0] != "alice" || owners[1] != "bob" {
		t.Fatalf("owners=%v", owners)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
