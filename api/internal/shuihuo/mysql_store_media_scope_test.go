package shuihuo

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func nullableID(value int64) driver.Value {
	if value == 0 {
		return nil
	}
	return value
}

func TestMySQLStoreCreateMediaTaskRejectsEveryCrossScopeReference(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input CreateMediaTaskInput
	}{
		{name: "book outside project", input: CreateMediaTaskInput{BatchProjectID: 7, BookID: 88, Kind: MediaImage, Provider: "image", Model: "model", RequestID: "req-book"}},
		{name: "segment outside project book", input: CreateMediaTaskInput{BatchProjectID: 7, BookID: 8, SegmentID: 41, Kind: MediaImage, Provider: "image", Model: "model", RequestID: "req-segment"}},
		{name: "source asset outside project book", input: CreateMediaTaskInput{BatchProjectID: 7, BookID: 8, SourceAssetID: 51, Kind: MediaAudio, Provider: "tts", Model: "model", RequestID: "req-asset"}},
		{name: "production task job outside project book", input: CreateMediaTaskInput{BatchProjectID: 7, BookID: 8, ProductionTaskID: 31, Kind: MediaVideo, Provider: "volc", Model: "model", RequestID: "req-video"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			i := tc.input
			args := []driver.Value{
				i.BatchProjectID, i.BookID, nullableID(i.SegmentID), nullableID(i.SourceAssetID), nullableID(i.ProductionTaskID), i.Kind, i.Provider, i.Model, i.RequestID,
				i.BatchProjectID, i.BookID,
				i.SegmentID, i.SegmentID, i.BatchProjectID, i.BookID,
				i.SourceAssetID, i.SourceAssetID, i.BatchProjectID, i.BookID,
				i.ProductionTaskID, i.ProductionTaskID, i.BatchProjectID, i.BookID,
			}
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id=? FOR UPDATE")).WithArgs(i.BatchProjectID).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(nil))
			mock.ExpectExec(`(?s)INSERT INTO shuihuo_media_tasks.*SELECT.*batch_projects.*books.*shuihuo_storyboard_segments.*shuihuo_media_assets.*video_production_tasks.*video_production_jobs`).
				WithArgs(args...).
				WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectRollback()

			_, err = NewMySQLStore(db).CreateMediaTask(context.Background(), i)
			if !errors.Is(err, ErrNotFound) {
				t.Fatalf("err=%v, want ErrNotFound", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMySQLStoreCreateMediaTaskRejectsArchivedProjectUnderLock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	archivedAt := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id=? FOR UPDATE")).WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(archivedAt))
	mock.ExpectRollback()
	_, err = NewMySQLStore(db).CreateMediaTask(context.Background(), CreateMediaTaskInput{BatchProjectID: 7, BookID: 8, Kind: MediaImage})
	if !errors.Is(err, ErrProjectArchived) {
		t.Fatalf("err=%v, want ErrProjectArchived", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreRetryMediaTaskRejectsArchivedProjectUnderLock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	archivedAt := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id=? FOR UPDATE")).WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(archivedAt))
	mock.ExpectRollback()
	_, err = NewMySQLStore(db).RetryMediaTask(context.Background(), 7, 8, 9)
	if !errors.Is(err, ErrProjectArchived) {
		t.Fatalf("err=%v, want ErrProjectArchived", err)
	}
}

func TestMySQLStoreValidateAssetScopeRejectsForeignBookOrSegment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`(?s)SELECT EXISTS.*batch_projects.*books.*shuihuo_storyboard_segments`).
		WithArgs(int64(7), int64(8), int64(41), int64(41), int64(7), int64(8)).
		WillReturnRows(sqlmock.NewRows([]string{"valid"}).AddRow(false))

	err = NewMySQLStore(db).ValidateAssetScope(context.Background(), 7, 8, 41)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err=%v, want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
