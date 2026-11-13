package screenplay

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"
)

func TestMySQLDocumentLoadsExactOriginalAndEditable(t *testing.T) {
	db, m, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now()
	raw := "  {\"scenes\":[]}\r\n "
	editable := "# 第1集：测试\n## 1-1：日 内 家\n△ 门开了"
	rows := sqlmock.NewRows([]string{"id", "project_id", "owner_username", "source", "version", "status", "content", "raw_result", "content_sha256", "chapter_count", "episode_count", "segment_count", "revision", "created_at", "updated_at"}).AddRow("d", "p", "o", "b", 1, "active", editable, raw, contentDigest(editable), 0, 1, 1, 1, now, now)
	m.ExpectQuery(`SELECT id, project_id, owner_username, source, version, status, content, COALESCE\(raw_result, ''\), content_sha256`).WithArgs("p", "o").WillReturnRows(rows)
	doc, err := NewMySQLStore(db).ActiveDocument(context.Background(), "o", "p")
	if err != nil {
		t.Fatal(err)
	}
	if doc.RawResult != raw || doc.Content != editable {
		t.Fatal("raw and editable changed")
	}
	if err = m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
