package novelfetchworkshop

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreMarkBodyReleasableStartsRetentionFromReleaseTime(t *testing.T) {
	_, mock, store := newBodySQLMock(t)
	ctx := context.Background()
	released := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)
	expires := released.Add(7 * 24 * time.Hour)
	updated := time.Date(2026, 9, 2, 10, 1, 0, 0, time.UTC)
	lastNeeded := time.Date(2026, 9, 2, 9, 59, 0, 0, time.UTC)

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE novel_fetch_workshop_bodies SET state='releasable', releasable_at=?, expires_at=? WHERE owner_username=? AND book_id=? AND version_id=?`)).
		WithArgs(released, expires, "alice", "123", "ai3").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT revision, content_hash, char_count, state, updated_at, last_needed_at, releasable_at, expires_at FROM novel_fetch_workshop_bodies WHERE owner_username=? AND book_id=? AND version_id=?`)).
		WithArgs("alice", "123", "ai3").
		WillReturnRows(sqlmock.NewRows([]string{"revision", "content_hash", "char_count", "state", "updated_at", "last_needed_at", "releasable_at", "expires_at"}).
			AddRow(uint64(2), "hash", int64(12), "releasable", updated, lastNeeded, released, expires))

	ref, err := store.MarkBodyReleasable(ctx, "alice", "123", "ai3", released, 0)
	if err != nil {
		t.Fatal(err)
	}
	if ref.State != "releasable" || ref.ReleasableAt != released.Format(time.RFC3339Nano) || ref.ExpiresAt != expires.Format(time.RFC3339Nano) {
		t.Fatalf("ref=%+v", ref)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreExpiredCleanupSelectsOnlyExpiredReleasableBodies(t *testing.T) {
	_, mock, store := newBodySQLMock(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 9, 10, 0, 0, 0, time.UTC)
	released := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT book_id, version_id, content_encoding, content_blob, releasable_at FROM novel_fetch_workshop_bodies WHERE owner_username=? AND state='releasable' AND releasable_at IS NOT NULL AND expires_at IS NOT NULL AND expires_at<=? ORDER BY releasable_at ASC, book_id ASC, version_id ASC LIMIT ?`)).
		WithArgs("alice", now, 10).
		WillReturnRows(sqlmock.NewRows([]string{"book_id", "version_id", "content_encoding", "content_blob", "releasable_at"}).AddRow("123", "ai3", "inline", nil, released))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM novel_fetch_workshop_bodies WHERE owner_username=? AND book_id=? AND version_id=? AND state='releasable' AND releasable_at=?`)).
		WithArgs("alice", "123", "ai3", released).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE novel_fetch_workshop_documents SET document_json=JSON_REMOVE(document_json, ?), updated_at=CURRENT_TIMESTAMP(6) WHERE owner_username=? AND book_id=?`)).
		WithArgs(`$.bodyRefs."ai3"`, "alice", "123").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := store.CleanupBodies(ctx, "alice", BodyCleanupRequest{Now: now, Reason: BodyCleanupReasonExpired, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if result.Deleted != 1 || len(result.Results) != 1 || result.Results[0].VersionID != "ai3" || result.Results[0].Reason != BodyCleanupReasonExpired {
		t.Fatalf("result=%+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreCapacityCleanupUsesOldestReleasableBodiesOnly(t *testing.T) {
	_, mock, store := newBodySQLMock(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	oldest := time.Date(2026, 9, 2, 8, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT book_id, version_id, content_encoding, content_blob, releasable_at FROM novel_fetch_workshop_bodies WHERE owner_username=? AND state='releasable' AND releasable_at IS NOT NULL ORDER BY releasable_at ASC, book_id ASC, version_id ASC LIMIT ?`)).
		WithArgs("alice", 1).
		WillReturnRows(sqlmock.NewRows([]string{"book_id", "version_id", "content_encoding", "content_blob", "releasable_at"}).AddRow("456", "ai1", "inline", nil, oldest))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM novel_fetch_workshop_bodies WHERE owner_username=? AND book_id=? AND version_id=? AND state='releasable' AND releasable_at=?`)).
		WithArgs("alice", "456", "ai1", oldest).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE novel_fetch_workshop_documents SET document_json=JSON_REMOVE(document_json, ?), updated_at=CURRENT_TIMESTAMP(6) WHERE owner_username=? AND book_id=?`)).
		WithArgs(`$.bodyRefs."ai1"`, "alice", "456").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := store.CleanupBodies(ctx, "alice", BodyCleanupRequest{Now: now, Reason: BodyCleanupReasonCapacity, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if result.Deleted != 1 || result.Results[0].VersionID != "ai1" || result.Results[0].Reason != BodyCleanupReasonCapacity {
		t.Fatalf("result=%+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
