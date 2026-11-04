package batchfactoryv11

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListBatchRecoveryIndexReadsOnlyGiantRecoveryFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	// 先只取排好序的 book id，避免 source_text/草稿正文参与 filesort（MySQL 1038）。
	idQuery := `SELECT b.id,b.batch_id FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id JOIN batch_factory_v12_batches ba ON ba.id=b.batch_id AND ba.owner_username=b.owner_username WHERE b.owner_username=? AND JSON_UNQUOTE(JSON_EXTRACT(r.source_metadata_json,'$.sourceMode'))='giant_material' ORDER BY ba.created_at DESC,ba.id DESC,b.ordinal,b.id`
	mock.ExpectQuery(regexp.QuoteMeta(idQuery)).
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"book_id", "batch_id"}).AddRow("book-1", "batch-1"))
	query := `SELECT ba.id,ba.revision,ba.created_at,ba.updated_at,COALESCE(bsp.patch_json,JSON_OBJECT()),b.id,b.revision,COALESCE(r.source_text,''),r.source_metadata_json,COALESCE(d.content,''),COALESCE(ksp.patch_json,JSON_OBJECT()) FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id JOIN batch_factory_v12_batches ba ON ba.id=b.batch_id AND ba.owner_username=b.owner_username LEFT JOIN batch_factory_v12_settings_patches bsp ON bsp.scope_type='batch' AND bsp.scope_id=ba.id AND bsp.owner_username=ba.owner_username LEFT JOIN batch_factory_v12_settings_patches ksp ON ksp.scope_type='book' AND ksp.scope_id=b.id AND ksp.owner_username=b.owner_username LEFT JOIN batch_factory_v12_drafts d ON d.owner_username=b.owner_username AND d.draft_key=CONCAT('working-front:',b.id) AND d.kind='working-front-content' AND d.scope=ba.id WHERE b.owner_username=? AND b.id IN (?)`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("alice", "book-1").
		WillReturnRows(sqlmock.NewRows([]string{"batch_id", "batch_revision", "created_at", "updated_at", "batch_patch", "book_id", "book_revision", "source_text", "source_metadata", "working_content", "book_patch"}).
			AddRow("batch-1", int64(7), now, now, []byte(`{"textModelId":"model-a"}`), "book-1", int64(3), "", []byte(`{"sourceMode":"giant_material","executorJobId":"job-1"}`), "OCR 正文", []byte(`{"openingEnabled":true}`)))

	got, err := NewMySQLStore(db).ListBatchRecoveryIndex(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "batch-1" || len(got[0].Books) != 1 {
		t.Fatalf("unexpected recovery index: %#v", got)
	}
	if got[0].Books[0].WorkingFrontContent != "OCR 正文" || got[0].Books[0].SourceMetadata["executorJobId"] != "job-1" {
		t.Fatalf("missing recovery fields: %#v", got[0].Books[0])
	}
	if string(got[0].SettingsState.Patch["textModelId"]) != `"model-a"` || string(got[0].Books[0].SettingsState.Patch["openingEnabled"]) != "true" {
		t.Fatalf("missing settings patches: batch=%v book=%v", got[0].SettingsState.Patch, got[0].Books[0].SettingsState.Patch)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
