package batchfactoryv11

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// jsonTextArg deliberately accepts only text. go-sql-driver/mysql maps []byte
// to _binary, which MySQL JSON columns reject.
type jsonTextArg struct{}

func (jsonTextArg) Match(value driver.Value) bool {
	_, ok := value.(string)
	return ok
}

func TestMySQLCreateIntakeKeepsSameBookIDAcrossPlatforms(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO batch_factory_v11_intakes").
		WithArgs(sqlmock.AnyArg(), "alice", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	store := NewMySQLStore(db)
	intake, err := store.CreateIntake(context.Background(), "alice", NovelFetchIntakeInput{Books: []CreateBookInput{
		{ID: "737092", Title: "甲", Platform: "3"},
		{ID: "737092", Title: "乙", Platform: "15"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	var got NovelFetchIntakeInput
	if err := json.Unmarshal(intake.Payload, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Books) != 2 {
		t.Fatalf("cross-platform same ID must keep both books, got %d: %+v", len(got.Books), got.Books)
	}
	if got.Books[0].Platform != "3" || got.Books[1].Platform != "15" {
		t.Fatalf("books=%+v", got.Books)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLCreateIntakeDeduplicatesSamePlatformBookID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO batch_factory_v11_intakes").
		WithArgs(sqlmock.AnyArg(), "alice", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	store := NewMySQLStore(db)
	intake, err := store.CreateIntake(context.Background(), "alice", NovelFetchIntakeInput{Books: []CreateBookInput{
		{ID: "737092", Title: "甲", Platform: "3"},
		{ID: "737092", Title: "甲重复行", Platform: "3"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	var got NovelFetchIntakeInput
	if err := json.Unmarshal(intake.Payload, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Books) != 1 || got.Books[0].Title != "甲" {
		t.Fatalf("same-platform duplicate must collapse to the first row, got %+v", got.Books)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLCreateIntakeBindsJSONPayloadAsText(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// MySQL JSON columns reject a driver-bound []byte value because the driver
	// labels it as _binary. The persisted value must instead be bound as UTF-8
	// text so MySQL can parse it as JSON.
	mock.ExpectExec("INSERT INTO batch_factory_v11_intakes").
		WithArgs(sqlmock.AnyArg(), "alice", `{"books":[]}`, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	store := NewMySQLStore(db)
	if _, err := store.CreateIntake(context.Background(), "alice", NovelFetchIntakeInput{Books: []CreateBookInput{}}); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateBatchTxBindsGiantMetadataAndSettingsAsText(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("INSERT INTO batch_factory_v11_batches").
		WithArgs(sqlmock.AnyArg(), "alice", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO batch_factory_v11_batch_records").
		WithArgs(sqlmock.AnyArg(), "giant", nil).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO batch_factory_v11_books").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "alice", 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO batch_factory_v11_book_records").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, nil, "book", nil, nil, nil, jsonTextArg{}).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE batch_factory_v11_batches").
		WithArgs(int64(2), sqlmock.AnyArg(), sqlmock.AnyArg(), "alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO batch_factory_v11_settings_patches").
		WithArgs(string(ScopeBatch), sqlmock.AnyArg(), "alice", sqlmock.AnyArg(), nil, nil, jsonTextArg{}, int64(2), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	_, err = createBatchTx(context.Background(), tx, "alice", CreateBatchInput{
		Title: "giant",
		Books: []CreateBookInput{{
			Title:          "book",
			SourceMetadata: map[string]any{"sourceMode": "giant_material"},
		}},
		GiantAutomationPlan:  map[string]any{"presetId": "preset"},
		InitialBatchSettings: SettingsPatch{"automationMode": json.RawMessage(`"full_auto"`)},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLAppendBooksFromIntakeBindsSourceMetadataAsText(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT payload_json,consumed_at FROM batch_factory_v11_intakes").
		WithArgs("intake-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"payload_json", "consumed_at"}).AddRow(
			`{"books":[{"bookId":"book-1","title":"书","platform":"15","sourceMetadata":{"platformId":"15"}}]}`,
			nil,
		))
	mock.ExpectQuery("SELECT revision FROM batch_factory_v11_batches").
		WithArgs("batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT COALESCE\\(source_book_id").
		WithArgs("batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"source_book_id", "platform"}))
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(ordinal\\)\\+1,0\\)").
		WithArgs("batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"nextOrdinal"}).AddRow(0))
	mock.ExpectExec("INSERT INTO batch_factory_v11_books").
		WithArgs(sqlmock.AnyArg(), "batch-1", "alice", 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO batch_factory_v11_book_records").
		WithArgs(sqlmock.AnyArg(), "book-1", nil, "15", "书", nil, nil, "book-1.txt", jsonTextArg{}).
		WillReturnError(errors.New("stop after append metadata write"))
	mock.ExpectRollback()

	store := NewMySQLStore(db)
	_, err = store.AppendBooksFromIntake(context.Background(), "alice", "batch-1", "intake-1", false)
	if err == nil || err.Error() != "stop after append metadata write" {
		t.Fatalf("AppendBooksFromIntake error = %v, want metadata write sentinel", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLUpdateBookMetadataBindsJSONAsText(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT revision FROM batch_factory_v11_books").
		WithArgs("book-1", "batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(int64(1)))
	// The deliberately returned error stops before hydration. It proves the
	// JSON-column argument matched as text rather than driver-bound _binary.
	mock.ExpectExec("UPDATE batch_factory_v11_book_records").
		WithArgs(jsonTextArg{}, "book-1").
		WillReturnError(errors.New("stop after metadata write"))
	mock.ExpectRollback()

	store := NewMySQLStore(db)
	_, err = store.UpdateBookMetadata(context.Background(), "alice", "batch-1", "book-1", UpdateBookMetadataInput{
		Metadata:         map[string]any{"executorJobId": "giant-job-1"},
		ExpectedRevision: 1,
	})
	if err == nil || err.Error() != "stop after metadata write" {
		t.Fatalf("UpdateBookMetadata error = %v, want metadata write sentinel", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLCaptureBookSourceBindsJSONAsText(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT revision FROM batch_factory_v11_books").
		WithArgs("book-1", "batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT source_text,title,source_book_id,source_metadata_json").
		WithArgs("book-1").
		WillReturnRows(sqlmock.NewRows([]string{"source_text", "title", "source_book_id", "source_metadata_json"}).AddRow("", "小说 1", "1", []byte(`{}`)))
	mock.ExpectExec("UPDATE batch_factory_v11_book_records").
		WithArgs("小说 1", "正文", "正文", jsonTextArg{}, "book-1").
		WillReturnError(errors.New("stop after source write"))
	mock.ExpectRollback()

	store := NewMySQLStore(db)
	_, err = store.CaptureBookSource(context.Background(), "alice", "batch-1", "book-1", CaptureBookSourceInput{
		SourceText:       "正文",
		SourceMetadata:   map[string]any{"originalReadVia": "bookstore"},
		ExpectedRevision: 1,
	})
	if err == nil || err.Error() != "stop after source write" {
		t.Fatalf("CaptureBookSource error = %v, want source write sentinel", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
