package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"
)

func TestContentRecordSQLScopedRoundTripAndManualRollback(t *testing.T) {
	st, m := newNovelCreationMockStore(t)
	ctx := context.Background()
	data := json.RawMessage(`{"Job":{"ID":"d","Status":"failed"},"Object":{"Key":"novel-documents/owned/d"}}`)
	r := ContentRecord{Owner: "alice", ProjectID: "p", Kind: "document", ID: "d", Status: "failed", Data: data}
	m.ExpectExec("INSERT INTO novel_content_records.*ON DUPLICATE KEY UPDATE").WithArgs("alice", "p", "document", "d", "failed", nil, []byte(data)).WillReturnResult(sqlmock.NewResult(0, 1))
	if e := st.PutContentRecord(ctx, r); e != nil {
		t.Fatal(e)
	}
	m.ExpectQuery("SELECT project_id, title").WithArgs("alice", "p").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "genre", "world", "protagonist", "conflict", "model", "words", "status", "created", "updated", "deleted"}).AddRow("p", "book", "fantasy", nil, nil, nil, "fake", 2000, "draft", time.Now(), time.Now(), nil))
	m.ExpectQuery("SELECT status,lease_until,snapshot_json FROM novel_content_records WHERE owner_username=\\? AND project_id=\\? AND kind=\\? AND record_id=\\?").WithArgs("alice", "p", "document", "d").WillReturnRows(sqlmock.NewRows([]string{"status", "lease_until", "snapshot_json"}).AddRow("failed", nil, data))
	got, ok, e := st.GetContentRecord(ctx, "alice", "p", "document", "d")
	if e != nil || !ok || string(got.Data) != string(data) {
		t.Fatalf("%+v %v", got, e)
	}
	m.ExpectBegin()
	m.ExpectQuery("SELECT 1 FROM novel_projects.*status NOT IN.*archived.*purging.*FOR UPDATE").WithArgs("alice", "p").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	m.ExpectExec("UPDATE novel_chapters.*AND version_no=\\?").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("INSERT INTO novel_chapter_versions.*NULL").WillReturnError(errors.New("storage unavailable"))
	m.ExpectRollback()
	e = st.Atomic(ctx, "alice", "p", func(tx Store) error {
		return tx.WriteManualChapter(ctx, "alice", Chapter{ID: "c", ProjectID: "p", Version: 2, Content: "new", UpdatedAt: time.Now()}, ChapterVersion{ID: "v", ChapterID: "c", Version: 2, Content: "new", Snapshot: json.RawMessage(`{"source":"manual_edit"}`), CreatedAt: time.Now()})
	})
	if e == nil {
		t.Fatal("version failure ignored")
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
