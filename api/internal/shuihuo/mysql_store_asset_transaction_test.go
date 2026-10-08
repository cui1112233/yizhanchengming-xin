package shuihuo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func createAssetFixture() CreateAssetInput {
	return CreateAssetInput{
		BatchProjectID: 7,
		BookID:         8,
		SegmentID:      9,
		Type:           AssetImage,
		Bucket:         "private-assets",
		ObjectKey:      "staging/shuihuo/project-7/book-8/new.png",
		Metadata:       []byte(`{"filename":"new.png"}`),
		Status:         AssetReady,
	}
}

func expectCreateAssetInsert(mock sqlmock.Sqlmock, i CreateAssetInput) *sqlmock.ExpectedExec {
	return mock.ExpectExec(`(?s)INSERT INTO shuihuo_media_assets.*SELECT.*batch_projects.*books.*shuihuo_storyboard_segments`).
		WithArgs(i.BatchProjectID, i.BookID, i.SegmentID, i.Type, i.Bucket, i.ObjectKey, i.Metadata, i.Status,
			i.BatchProjectID, i.BookID, i.SegmentID, i.SegmentID, i.BatchProjectID, i.BookID)
}

func assetReadbackRows(i CreateAssetInput) *sqlmock.Rows {
	now := time.Now().UTC()
	return sqlmock.NewRows([]string{"id", "batch_project_id", "book_id", "segment_id", "asset_type", "bucket", "object_key", "metadata_json", "status", "created_at", "updated_at"}).
		AddRow(int64(31), i.BatchProjectID, i.BookID, i.SegmentID, i.Type, i.Bucket, i.ObjectKey, i.Metadata, i.Status, now, now)
}

func TestMySQLStoreCreateAssetRollsBackWhenReadbackFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	i := createAssetFixture()
	readErr := errors.New("readback failed")

	mock.ExpectBegin()
	expectCreateAssetInsert(mock, i).WillReturnResult(sqlmock.NewResult(31, 1))
	mock.ExpectQuery(`SELECT id,batch_project_id,book_id,segment_id,asset_type,bucket,object_key,metadata_json,status,created_at,updated_at FROM shuihuo_media_assets WHERE id=\?`).
		WithArgs(int64(31)).WillReturnError(readErr)
	mock.ExpectRollback()

	_, err = NewMySQLStore(db).CreateAsset(context.Background(), i)
	if !errors.Is(err, readErr) {
		t.Fatalf("err=%v, want readback failure", err)
	}
	if !isAssetDefinitelyNotPersisted(err) {
		t.Fatalf("err=%v, want rollback-confirmed not-persisted marker", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreCreateAssetCommitErrorIsNotMarkedSafeForObjectDeletion(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	i := createAssetFixture()
	commitErr := errors.New("commit outcome unknown")

	mock.ExpectBegin()
	expectCreateAssetInsert(mock, i).WillReturnResult(sqlmock.NewResult(31, 1))
	mock.ExpectQuery(`SELECT id,batch_project_id,book_id,segment_id,asset_type,bucket,object_key,metadata_json,status,created_at,updated_at FROM shuihuo_media_assets WHERE id=\?`).
		WithArgs(int64(31)).WillReturnRows(assetReadbackRows(i))
	mock.ExpectCommit().WillReturnError(commitErr)

	_, err = NewMySQLStore(db).CreateAsset(context.Background(), i)
	if !errors.Is(err, commitErr) {
		t.Fatalf("err=%v, want commit failure", err)
	}
	if isAssetDefinitelyNotPersisted(err) {
		t.Fatalf("commit error must preserve unknown outcome and must not authorize TOS deletion: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreCreateAssetCommitsReadbackFromSameTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	i := createAssetFixture()

	mock.ExpectBegin()
	expectCreateAssetInsert(mock, i).WillReturnResult(sqlmock.NewResult(31, 1))
	mock.ExpectQuery(`SELECT id,batch_project_id,book_id,segment_id,asset_type,bucket,object_key,metadata_json,status,created_at,updated_at FROM shuihuo_media_assets WHERE id=\?`).
		WithArgs(int64(31)).WillReturnRows(assetReadbackRows(i))
	mock.ExpectCommit()

	asset, err := NewMySQLStore(db).CreateAsset(context.Background(), i)
	if err != nil || asset.ID != 31 || asset.ObjectKey != i.ObjectKey {
		t.Fatalf("asset=%+v err=%v", asset, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
