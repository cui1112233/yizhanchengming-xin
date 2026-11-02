package batchfactoryv11

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
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
	mock.ExpectExec("INSERT INTO batch_factory_v12_intakes").
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
	mock.ExpectExec("INSERT INTO batch_factory_v12_intakes").
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
	mock.ExpectExec("INSERT INTO batch_factory_v12_intakes").
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
	mock.ExpectExec("INSERT INTO batch_factory_v12_batches").
		WithArgs(sqlmock.AnyArg(), "alice", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO batch_factory_v12_batch_records").
		WithArgs(sqlmock.AnyArg(), "giant", nil).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO batch_factory_v12_books").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "alice", 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO batch_factory_v12_book_records").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, nil, "book", nil, nil, nil, jsonTextArg{}).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE batch_factory_v12_batches").
		WithArgs(int64(2), sqlmock.AnyArg(), sqlmock.AnyArg(), "alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO batch_factory_v12_settings_patches").
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
	mock.ExpectQuery("SELECT payload_json,consumed_at FROM batch_factory_v12_intakes").
		WithArgs("intake-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"payload_json", "consumed_at"}).AddRow(
			`{"books":[{"bookId":"book-1","title":"书","platform":"15","sourceMetadata":{"platformId":"15"}}]}`,
			nil,
		))
	mock.ExpectQuery("SELECT revision FROM batch_factory_v12_batches").
		WithArgs("batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT COALESCE\\(source_book_id").
		WithArgs("batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"source_book_id", "platform"}))
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(ordinal\\)\\+1,0\\)").
		WithArgs("batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"nextOrdinal"}).AddRow(0))
	mock.ExpectExec("INSERT INTO batch_factory_v12_books").
		WithArgs(sqlmock.AnyArg(), "batch-1", "alice", 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO batch_factory_v12_book_records").
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
	mock.ExpectQuery("SELECT revision FROM batch_factory_v12_books").
		WithArgs("book-1", "batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(int64(1)))
	// The deliberately returned error stops before hydration. It proves the
	// JSON-column argument matched as text rather than driver-bound _binary.
	mock.ExpectExec("UPDATE batch_factory_v12_book_records").
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

func TestMySQLCreateProductionJobPropagatesTaskAttemptConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM batch_factory_v12_production_jobs").WithArgs("alice", "batch-1", "book-1", "request-1").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(attempt\\),0\\) FROM batch_factory_v12_production_tasks").WithArgs("alice", "video-1", 0).WillReturnRows(sqlmock.NewRows([]string{"attempt"}).AddRow(1))
	mock.ExpectExec("INSERT INTO batch_factory_v12_production_jobs").WillReturnResult(sqlmock.NewResult(1, 1))
	conflict := &mysql.MySQLError{Number: 1062, Message: "Duplicate entry for key 'uq_bfv11_production_task_attempt'"}
	mock.ExpectExec("INSERT INTO batch_factory_v12_production_tasks").WillReturnError(conflict)
	mock.ExpectRollback()

	store := NewMySQLStore(db)
	_, err = store.CreateProductionJob(context.Background(), ProductionJob{Owner: "alice", BatchID: "batch-1", BookID: "book-1", RequestID: "request-1", DirectorRevisionID: "director-1", Tasks: []ProductionTask{{VideoID: "video-1", Attempt: 1, FinalPromptHash: "hash", CompiledPrompt: "prompt"}}})
	if !errors.Is(err, conflict) {
		t.Fatalf("CreateProductionJob() error = %v, want task attempt conflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLCreateProductionJobAllocatesNextAttemptForDifferentRequest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM batch_factory_v12_production_jobs").
		WithArgs("alice", "batch-1", "book-1", "request-2").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(attempt\\),0\\) FROM batch_factory_v12_production_tasks").
		WithArgs("alice", "video-1", 0).
		WillReturnRows(sqlmock.NewRows([]string{"attempt"}).AddRow(1))
	mock.ExpectExec("INSERT INTO batch_factory_v12_production_jobs").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO batch_factory_v12_production_tasks").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "alice", "video-1", 0, nil, nil, nil, 2, "hash", "prompt", nil, "[]", "[]", 0.0, 0.0, 0.0, nil, nil, "queued", nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO batch_factory_v12_production_events").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO batch_factory_v12_production_events").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT id,batch_id,book_id,request_id,director_revision_id,status,created_at,updated_at FROM batch_factory_v12_production_jobs").
		WithArgs(sqlmock.AnyArg(), "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "batch_id", "book_id", "request_id", "director_revision_id", "status", "created_at", "updated_at"}).AddRow("new-job", "batch-1", "book-1", "request-2", "director-1", "queued", time.Now(), time.Now()))
	mock.ExpectQuery("SELECT id,video_id").
		WithArgs(sqlmock.AnyArg(), "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "opening_variant_index", "compilation_id", "compilation_segment_key", "provider", "status", "attempt", "final_prompt_hash", "compiled_prompt", "compile_trace_json", "reference_image_urls", "downgraded_asset_ids", "target_duration_seconds", "requested_duration_seconds", "actual_duration_seconds", "provider_task_id", "media_url", "error_message", "created_at", "updated_at"}).AddRow("new-task", "video-1", 0, "", "", "personal_api", "queued", 2, "hash", "prompt", "null", "[]", "[]", 0, 0, 0, "", "", "", time.Now(), time.Now()))

	store := NewMySQLStore(db)
	job, err := store.CreateProductionJob(context.Background(), ProductionJob{Owner: "alice", BatchID: "batch-1", BookID: "book-1", RequestID: "request-2", DirectorRevisionID: "director-1", Tasks: []ProductionTask{{VideoID: "video-1", OpeningVariantIndex: 0, Status: ProductionQueued, Attempt: 1, FinalPromptHash: "hash", CompiledPrompt: "prompt", ReferenceImageURLs: []string{}, DowngradedAssetIDs: []string{}}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(job.Tasks) != 1 || job.Tasks[0].Attempt != 2 {
		t.Fatalf("CreateProductionJob() task attempts = %+v, want attempt 2", job.Tasks)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLCreateProductionJobAllocatesUniqueAttemptsWithinSameJob(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM batch_factory_v12_production_jobs").WithArgs("alice", "batch-1", "book-1", "request-same-key").WillReturnError(sql.ErrNoRows)
	for range 2 {
		mock.ExpectQuery("SELECT COALESCE\\(MAX\\(attempt\\),0\\) FROM batch_factory_v12_production_tasks").WithArgs("alice", "video-1", 0).WillReturnRows(sqlmock.NewRows([]string{"attempt"}).AddRow(0))
	}
	mock.ExpectExec("INSERT INTO batch_factory_v12_production_jobs").WillReturnResult(sqlmock.NewResult(1, 1))
	for range 2 {
		mock.ExpectExec("INSERT INTO batch_factory_v12_production_tasks").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO batch_factory_v12_production_events").WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectExec("INSERT INTO batch_factory_v12_production_events").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT id,batch_id,book_id,request_id,director_revision_id,status,created_at,updated_at").WithArgs(sqlmock.AnyArg(), "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "batch_id", "book_id", "request_id", "director_revision_id", "status", "created_at", "updated_at"}).AddRow("job-1", "batch-1", "book-1", "request-same-key", "director-1", "queued", time.Now(), time.Now()))
	mock.ExpectQuery("SELECT id,video_id").WithArgs(sqlmock.AnyArg(), "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "opening_variant_index", "compilation_id", "compilation_segment_key", "provider", "status", "attempt", "final_prompt_hash", "compiled_prompt", "compile_trace_json", "reference_image_urls", "downgraded_asset_ids", "target_duration_seconds", "requested_duration_seconds", "actual_duration_seconds", "provider_task_id", "media_url", "error_message", "created_at", "updated_at"}).AddRow("task-1", "video-1", 0, "", "", "personal_api", "queued", 1, "hash", "prompt", "null", "[]", "[]", 0, 0, 0, "", "", "", time.Now(), time.Now()).AddRow("task-2", "video-1", 0, "", "", "personal_api", "queued", 2, "hash", "prompt", "null", "[]", "[]", 0, 0, 0, "", "", "", time.Now(), time.Now()))

	store := NewMySQLStore(db)
	job, err := store.CreateProductionJob(context.Background(), ProductionJob{Owner: "alice", BatchID: "batch-1", BookID: "book-1", RequestID: "request-same-key", DirectorRevisionID: "director-1", Tasks: []ProductionTask{{VideoID: "video-1", OpeningVariantIndex: 0, FinalPromptHash: "hash", CompiledPrompt: "prompt"}, {VideoID: "video-1", OpeningVariantIndex: 0, FinalPromptHash: "hash", CompiledPrompt: "prompt"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(job.Tasks) != 2 || job.Tasks[0].Attempt != 1 || job.Tasks[1].Attempt != 2 {
		t.Fatalf("same-key attempts = %+v, want 1 and 2", job.Tasks)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLCreateProductionJobAllowsOpeningVariantsAndIncrementsVariantAttempt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM batch_factory_v12_production_jobs").WithArgs("alice", "batch-1", "book-1", "request-variants").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(attempt\\),0\\) FROM batch_factory_v12_production_tasks").WithArgs("alice", "video-1", 0).WillReturnRows(sqlmock.NewRows([]string{"attempt"}).AddRow(0))
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(attempt\\),0\\) FROM batch_factory_v12_production_tasks").WithArgs("alice", "video-1", 1).WillReturnRows(sqlmock.NewRows([]string{"attempt"}).AddRow(0))
	mock.ExpectExec("INSERT INTO batch_factory_v12_production_jobs").WillReturnResult(sqlmock.NewResult(1, 1))
	for range 2 {
		mock.ExpectExec("INSERT INTO batch_factory_v12_production_tasks").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO batch_factory_v12_production_events").WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectExec("INSERT INTO batch_factory_v12_production_events").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT id,batch_id,book_id,request_id,director_revision_id,status,created_at,updated_at FROM batch_factory_v12_production_jobs").WithArgs(sqlmock.AnyArg(), "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "batch_id", "book_id", "request_id", "director_revision_id", "status", "created_at", "updated_at"}).AddRow("job-1", "batch-1", "book-1", "request-variants", "director-1", "queued", time.Now(), time.Now()))
	mock.ExpectQuery("SELECT id,video_id").WithArgs(sqlmock.AnyArg(), "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "opening_variant_index", "compilation_id", "compilation_segment_key", "provider", "status", "attempt", "final_prompt_hash", "compiled_prompt", "compile_trace_json", "reference_image_urls", "downgraded_asset_ids", "target_duration_seconds", "requested_duration_seconds", "actual_duration_seconds", "provider_task_id", "media_url", "error_message", "created_at", "updated_at"}).AddRow("task-0", "video-1", 0, "", "", "personal_api", "queued", 1, "hash", "prompt", "null", "[]", "[]", 0, 0, 0, "", "", "", time.Now(), time.Now()).AddRow("task-1", "video-1", 1, "", "", "personal_api", "queued", 1, "hash", "prompt", "null", "[]", "[]", 0, 0, 0, "", "", "", time.Now(), time.Now()))

	store := NewMySQLStore(db)
	job, err := store.CreateProductionJob(context.Background(), ProductionJob{Owner: "alice", BatchID: "batch-1", BookID: "book-1", RequestID: "request-variants", DirectorRevisionID: "director-1", Tasks: []ProductionTask{{VideoID: "video-1", OpeningVariantIndex: 0, FinalPromptHash: "hash", CompiledPrompt: "prompt"}, {VideoID: "video-1", OpeningVariantIndex: 1, FinalPromptHash: "hash", CompiledPrompt: "prompt"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(job.Tasks) != 2 || job.Tasks[0].Attempt != 1 || job.Tasks[1].Attempt != 1 {
		t.Fatalf("variant attempts = %+v, want both attempt 1", job.Tasks)
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
	mock.ExpectQuery("SELECT revision FROM batch_factory_v12_books").
		WithArgs("book-1", "batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT source_text,title,source_book_id,source_metadata_json").
		WithArgs("book-1").
		WillReturnRows(sqlmock.NewRows([]string{"source_text", "title", "source_book_id", "source_metadata_json"}).AddRow("", "小说 1", "1", []byte(`{}`)))
	mock.ExpectExec("UPDATE batch_factory_v12_book_records").
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

func TestMySQLCaptureBookSourceRejectsObjectPlaceholderBeforeDatabaseWrite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewMySQLStore(db)
	_, err = store.CaptureBookSource(context.Background(), "alice", "batch-1", "book-1", CaptureBookSourceInput{
		SourceText:       " [object Object] ",
		ReplaceSource:    true,
		ExpectedRevision: 1,
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("CaptureBookSource error = %v, want ErrInvalid", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
