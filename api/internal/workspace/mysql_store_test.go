package workspace

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRecentProjectionSQLKeepsOwnershipJoinsAndStableOrder(t *testing.T) {
	for _, fragment := range []string{
		"JOIN auth_batch_project_ownership o ON o.batch_project_id=bp.id",
		"o.owner_user_id=? OR (? > 0 AND o.team_id IS NOT NULL AND o.team_id=?)",
		"JOIN books b ON b.id=br.book_id AND b.intake_id=vp.intake_id",
		"JOIN video_production_jobs linked_vpj ON linked_vpj.id=linked_vpt.production_job_id",
		"linked_vpj.batch_project_id=smt.batch_project_id",
		"linked_vpj.book_id=smt.book_id",
		"COALESCE(CASE WHEN linked_vpj.id IS NOT NULL THEN linked_vpt.status END,smt.status)",
		"GREATEST(smt.updated_at,COALESCE(CASE WHEN linked_vpj.id IS NOT NULL THEN linked_vpt.updated_at END,smt.updated_at))",
		"smt.batch_project_id=rv.batch_project_id",
		"smt.book_id=rv.book_id",
		"ORDER BY updated_at DESC,kind_rank ASC,project_id DESC,book_id DESC,source_id DESC",
	} {
		if !strings.Contains(scopedRecentQuery, fragment) {
			t.Fatalf("scoped recent query lost required contract %q", fragment)
		}
	}
	if strings.Contains(recentProjectionQuery, "SELECT 'intake'") {
		t.Fatal("intake has no direct ownership fact and must not be exposed as recent activity")
	}
	if strings.Contains(elevatedVisibleProjects, "auth_batch_project_ownership") || strings.Contains(elevatedVisibleProjects, "owner_user_id") {
		t.Fatal("elevated projection must remain a separate fixed query without user-controlled ownership bypass")
	}
}

func TestMySQLStoreListRecentScopesOwnerAndNonZeroTeamInOneQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Date(2026, 10, 9, 4, 5, 6, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("WITH visible_projects AS (")).
		WithArgs(int64(7), int64(3), int64(3), 6).
		WillReturnRows(sqlmock.NewRows([]string{"kind", "origin", "source_id", "title", "status", "updated_at", "project_id"}).
			AddRow("script", "stage", 81, "真实小说", "completed", now, 12))

	items, err := NewMySQLStore(db).ListRecent(context.Background(), 7, 3, 6)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items=%#v, want one item", items)
	}
	got := items[0]
	if got.Kind != "script" || got.ID != "script:stage:81" || got.Title != "真实小说" || got.Status != "completed" || !got.UpdatedAt.Equal(now) {
		t.Fatalf("item=%#v", got)
	}
	if got.Href != "/batch-factory?projectId=12" {
		t.Fatalf("href=%q", got.Href)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreListRecentDoesNotTreatTeamZeroAsSharedOwnership(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("WITH visible_projects AS (")).
		WithArgs(int64(7), int64(0), int64(0), 6).
		WillReturnRows(sqlmock.NewRows([]string{"kind", "origin", "source_id", "title", "status", "updated_at", "project_id"}))

	items, err := NewMySQLStore(db).ListRecent(context.Background(), 7, 0, 6)
	if err != nil {
		t.Fatal(err)
	}
	if items == nil || len(items) != 0 {
		t.Fatalf("items=%#v, want non-nil empty slice", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreListRecentElevatedUsesDedicatedQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("WITH visible_projects AS (")).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows([]string{"kind", "origin", "source_id", "title", "status", "updated_at", "project_id"}))

	items, err := NewMySQLStore(db).ListRecentElevated(context.Background(), 20)
	if err != nil {
		t.Fatal(err)
	}
	if items == nil {
		t.Fatal("elevated empty result must serialize as []")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreListRecentNamespacesSameNumericIDFromDistinctTTSSources(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Date(2026, 10, 9, 4, 5, 6, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("WITH visible_projects AS (")).
		WithArgs(int64(7), int64(3), int64(3), 6).
		WillReturnRows(sqlmock.NewRows([]string{"kind", "origin", "source_id", "title", "status", "updated_at", "project_id"}).
			AddRow("tts", "measurement", 44, "测量", "measured", now, 12).
			AddRow("tts", "media", 44, "水货音频", "completed", now.Add(-time.Minute), 12))

	items, err := NewMySQLStore(db).ListRecent(context.Background(), 7, 3, 6)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].ID != "tts:measurement:44" || items[1].ID != "tts:media:44" || items[0].ID == items[1].ID {
		t.Fatalf("items=%#v, want source-namespaced TTS identities", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreListRecentReturnsSafeWrappedReadError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("WITH visible_projects AS (")).
		WithArgs(int64(7), int64(3), int64(3), 6).
		WillReturnError(errors.New("mysql password=do-not-leak"))

	_, err = NewMySQLStore(db).ListRecent(context.Background(), 7, 3, 6)
	if err == nil || err.Error() == "mysql password=do-not-leak" {
		t.Fatalf("err=%v, want wrapped internal error", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
