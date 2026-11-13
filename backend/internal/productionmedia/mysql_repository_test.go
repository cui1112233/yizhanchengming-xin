package productionmedia

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLRepositoryEnqueueDefaultsRetentionAndIsIdempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	query := `INSERT INTO batch_factory_v12_media_objects(owner_username,media_kind,media_type,local_storage_ref,object_key,state,attempt_count,retention_days,created_at,updated_at) VALUES(?,?,?,?,?,'pending',0,?,CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE media_type=IF(media_type<>VALUES(media_type),VALUES(media_type),media_type),retention_days=IF(retention_days<>VALUES(retention_days),VALUES(retention_days),retention_days)`
	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs("alice", KindComicImage, "image/png", "image_123.png", "production-media/comic/images/image_123.png", 7).WillReturnResult(sqlmock.NewResult(1, 1))
	repo := NewMySQLRepository(db)
	err = repo.Enqueue(context.Background(), Object{Owner: "alice", MediaKind: KindComicImage, MediaType: "image/png", LocalStorageRef: "image_123.png", ObjectKey: "production-media/comic/images/image_123.png"})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLRepositoryRejectsUnsafeStorageReferenceBeforeSQL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewMySQLRepository(db)
	if err := repo.Enqueue(context.Background(), Object{Owner: "alice", MediaKind: KindVideo, LocalStorageRef: "../secret", ObjectKey: "production-media/video/secret"}); err == nil {
		t.Fatal("expected unsafe storage reference error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
