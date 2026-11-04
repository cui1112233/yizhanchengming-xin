package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestLegacyDocumentSplitPrefersNormalizedOriginal(t *testing.T) {
	document := Document{
		BookID:      "123",
		Meta:        map[string]any{"bookName": "测试小说"},
		Original:    "规范原文",
		OriginalRaw: "原始原文",
		Versions: map[string]any{
			"ai1": "AI1正文",
			"ai3": "AI3正文",
		},
	}

	bodies := legacyBodyRecords(document)
	if len(bodies) != 4 {
		t.Fatalf("bodies=%+v", bodies)
	}
	contents := map[string]string{}
	for _, body := range bodies {
		contents[body.VersionID] = body.Content
	}
	if contents["original"] != "规范原文" || contents["ai1"] != "AI1正文" || contents["ai3"] != "AI3正文" {
		t.Fatalf("contents=%+v", contents)
	}
	if contents["original_raw"] != "原始原文" {
		t.Fatalf("full raw content was not preserved: %v", contents)
	}

	lightweight := lightweightLegacyDocument(document, bodies)
	if lightweight.Original != "" || lightweight.OriginalRaw != "" {
		t.Fatalf("full original text remains: %+v", lightweight)
	}
	for versionID, value := range lightweight.Versions {
		if _, isText := value.(string); isText {
			t.Fatalf("version %s still contains full text: %#v", versionID, value)
		}
	}
}

func TestMySQLStoreMigrateLegacyDocumentBodiesIsIdempotent(t *testing.T) {
	db, mock, store := newBodySQLMock(t)
	ctx := context.Background()
	document := Document{
		BookID:      "123",
		Meta:        map[string]any{"bookName": "测试小说"},
		Original:    "规范原文",
		OriginalRaw: "原始原文",
		Versions: map[string]any{
			"ai1": "AI1正文",
			"ai3": "AI3正文",
		},
		Logs: []any{},
	}
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	migrated := lightweightLegacyDocument(document, legacyBodyRecords(document))
	migratedRaw, err := json.Marshal(migrated)
	if err != nil {
		t.Fatal(err)
	}

	selectSQL := regexp.QuoteMeta(`SELECT document_json FROM novel_fetch_workshop_documents WHERE owner_username=? AND book_id=? FOR UPDATE`)
	insertSQL := regexp.QuoteMeta(`INSERT IGNORE INTO novel_fetch_workshop_bodies(owner_username,book_id,version_id,revision,content_encoding,content_blob,content_hash,char_count,state,created_at,updated_at,last_needed_at,expires_at) VALUES(?,?,?,1,'gzip',?,?,?,?,CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6),NULL)`)
	updateSQL := regexp.QuoteMeta(`UPDATE novel_fetch_workshop_documents SET document_json=CAST(? AS CHAR CHARACTER SET utf8mb4), updated_at=CURRENT_TIMESTAMP(6) WHERE owner_username=? AND book_id=?`)

	mock.ExpectBegin()
	mock.ExpectQuery(selectSQL).WithArgs("alice", "123").WillReturnRows(sqlmock.NewRows([]string{"document_json"}).AddRow(raw))
	mock.ExpectExec(insertSQL).WithArgs("alice", "123", "original_raw", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(4), "ready").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(insertSQL).WithArgs("alice", "123", "original", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(4), "ready").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(insertSQL).WithArgs("alice", "123", "ai1", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(5), "ready").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(insertSQL).WithArgs("alice", "123", "ai3", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(5), "ready").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(updateSQL).WithArgs(sqlmock.AnyArg(), "alice", "123").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := store.MigrateLegacyDocumentBodies(ctx, "alice", "123"); err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(selectSQL).WithArgs("alice", "123").WillReturnRows(sqlmock.NewRows([]string{"document_json"}).AddRow(migratedRaw))
	mock.ExpectCommit()
	if err := store.MigrateLegacyDocumentBodies(ctx, "alice", "123"); err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	_ = db
}

func TestMySQLStoreGetDocumentMigratesLegacyBodiesLazily(t *testing.T) {
	_, mock, store := newBodySQLMock(t)
	ctx := context.Background()
	updated := time.Date(2026, 9, 2, 14, 0, 0, 0, time.UTC)
	document := Document{
		BookID:      "123",
		Original:    "规范原文",
		OriginalRaw: "原始原文",
		Versions:    map[string]any{"ai1": "AI1正文"},
		Logs:        []any{},
	}
	raw, _ := json.Marshal(document)
	migrated := lightweightLegacyDocument(document, legacyBodyRecords(document))
	migratedRaw, _ := json.Marshal(migrated)

	getSQL := regexp.QuoteMeta(`SELECT document_json, updated_at FROM novel_fetch_workshop_documents WHERE owner_username=? AND book_id=?`)
	selectSQL := regexp.QuoteMeta(`SELECT document_json FROM novel_fetch_workshop_documents WHERE owner_username=? AND book_id=? FOR UPDATE`)
	insertSQL := regexp.QuoteMeta(`INSERT IGNORE INTO novel_fetch_workshop_bodies(owner_username,book_id,version_id,revision,content_encoding,content_blob,content_hash,char_count,state,created_at,updated_at,last_needed_at,expires_at) VALUES(?,?,?,1,'gzip',?,?,?,?,CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6),NULL)`)
	updateSQL := regexp.QuoteMeta(`UPDATE novel_fetch_workshop_documents SET document_json=CAST(? AS CHAR CHARACTER SET utf8mb4), updated_at=CURRENT_TIMESTAMP(6) WHERE owner_username=? AND book_id=?`)

	mock.ExpectQuery(getSQL).WithArgs("alice", "123").WillReturnRows(sqlmock.NewRows([]string{"document_json", "updated_at"}).AddRow(raw, updated))
	mock.ExpectBegin()
	mock.ExpectQuery(selectSQL).WithArgs("alice", "123").WillReturnRows(sqlmock.NewRows([]string{"document_json"}).AddRow(raw))
	mock.ExpectExec(insertSQL).WithArgs("alice", "123", "original_raw", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(4), "ready").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(insertSQL).WithArgs("alice", "123", "original", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(4), "ready").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(insertSQL).WithArgs("alice", "123", "ai1", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(5), "ready").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(updateSQL).WithArgs(sqlmock.AnyArg(), "alice", "123").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(getSQL).WithArgs("alice", "123").WillReturnRows(sqlmock.NewRows([]string{"document_json", "updated_at"}).AddRow(migratedRaw, updated.Add(time.Second)))

	got, err := store.GetDocument(ctx, "alice", "123")
	if err != nil {
		t.Fatal(err)
	}
	if got.Original != "" || got.OriginalRaw != "" {
		t.Fatalf("legacy text still returned: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreListDocumentsNeverReturnsLegacyFullText(t *testing.T) {
	_, mock, store := newBodySQLMock(t)
	updated := time.Date(2026, 9, 2, 14, 0, 0, 0, time.UTC)
	document := Document{
		BookID:      "123",
		Original:    "规范原文",
		OriginalRaw: "原始原文",
		Versions:    map[string]any{"ai1": "AI1正文"},
		Logs:        []any{},
	}
	raw, _ := json.Marshal(document)
	// ListDocuments 改为两步查询：先只取排好序的 book_id，再按 id 批量取正文，
	// 避免 document_json 参与 filesort（MySQL ERROR 1038）。
	idSQL := regexp.QuoteMeta(`SELECT book_id FROM novel_fetch_workshop_documents WHERE owner_username=? ORDER BY updated_at DESC, book_id ASC`)
	mock.ExpectQuery(idSQL).WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"book_id"}).AddRow("123"))
	listSQL := regexp.QuoteMeta(`SELECT book_id, document_json, updated_at FROM novel_fetch_workshop_documents WHERE owner_username=? AND book_id IN (?)`)
	mock.ExpectQuery(listSQL).WithArgs("alice", "123").WillReturnRows(sqlmock.NewRows([]string{"book_id", "document_json", "updated_at"}).AddRow("123", raw, updated))

	got, err := store.ListDocuments(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Original != "" || got[0].OriginalRaw != "" {
		t.Fatalf("legacy full text leaked from list: %+v", got)
	}
	for versionID, value := range got[0].Versions {
		if _, isText := value.(string); isText {
			t.Fatalf("version %s leaked full text: %#v", versionID, value)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
