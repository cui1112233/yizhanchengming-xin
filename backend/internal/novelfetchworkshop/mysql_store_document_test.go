package novelfetchworkshop

import (
	"context"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStorePutDocumentConvertsJSONParameterToUTF8(t *testing.T) {
	_, mock, store := newBodySQLMock(t)
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO novel_fetch_workshop_documents(owner_username,book_id,document_json,created_at,updated_at) VALUES(?,?,CAST(? AS CHAR CHARACTER SET utf8mb4),CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE document_json=VALUES(document_json), updated_at=CURRENT_TIMESTAMP(6)`)).
		WithArgs("alice", "book-1", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := store.PutDocument(context.Background(), "alice", Document{BookID: "book-1", Original: "中文模型输出"}); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
