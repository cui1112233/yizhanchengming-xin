package agentstudio

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreGetAttachmentChecksProjectOwnershipBeforeObjectMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	actor := Actor{UserID: 7, TeamID: 2}
	now := time.Now().UTC()
	projectQuery := regexp.QuoteMeta(`SELECT id,owner_user_id,team_id,title,batch_project_id,book_id,created_at,updated_at FROM agent_projects WHERE id=? AND (owner_user_id=? OR (team_id IS NOT NULL AND team_id=?))`)
	mock.ExpectQuery(projectQuery).WithArgs(int64(9), int64(7), int64(2)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "owner_user_id", "team_id", "title", "batch_project_id", "book_id", "created_at", "updated_at"}).AddRow(9, 7, nil, "Draft", nil, nil, now, now),
	)
	attachmentQuery := regexp.QuoteMeta(`SELECT id,project_id,owner_user_id,bucket,object_key,filename,content_type,byte_size,sha256,created_at FROM agent_attachments WHERE project_id=? AND id=?`)
	mock.ExpectQuery(attachmentQuery).WithArgs(int64(9), int64(3)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "project_id", "owner_user_id", "bucket", "object_key", "filename", "content_type", "byte_size", "sha256", "created_at"}).AddRow(3, 9, 7, "bucket", "agent/project-9/key", "draft.txt", "text/plain", 6, "abc", now),
	)

	got, err := store.GetAttachment(context.Background(), actor, 9, 3)
	if err != nil {
		t.Fatal(err)
	}
	if got.ObjectKey != "agent/project-9/key" || got.OwnerUserID != 7 || got.Filename != "draft.txt" {
		t.Fatalf("attachment=%+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreGetAttachmentDoesNotQueryMetadataForUnauthorizedProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	projectQuery := regexp.QuoteMeta(`SELECT id,owner_user_id,team_id,title,batch_project_id,book_id,created_at,updated_at FROM agent_projects WHERE id=? AND (owner_user_id=? OR (team_id IS NOT NULL AND team_id=?))`)
	mock.ExpectQuery(projectQuery).WithArgs(int64(9), int64(8), int64(0)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "owner_user_id", "team_id", "title", "batch_project_id", "book_id", "created_at", "updated_at"}),
	)
	_, err = store.GetAttachment(context.Background(), Actor{UserID: 8}, 9, 3)
	if err != ErrNotFound {
		t.Fatalf("error=%v want %v", err, ErrNotFound)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
