package novelfetchworkshop

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
)

func TestAutomationStatusInputsIsOwnerScopedAndKeepsLegacyFallback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT COALESCE").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"settings"}).AddRow(`{"novel_fetch_automation":{"enabled":true},"novel_fetch_current_batch_id":"batch-1"}`))
	mock.ExpectQuery("SELECT CASE WHEN").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"document"}).AddRow(`{"meta":{"status":"waiting_retry"}}`).AddRow(`{"original":"旧任务正文","versions":{"ai1":"已生成"},"meta":{}}`))
	s := &MySQLStore{DB: db}
	settings, docs, err := s.AutomationStatusInputs(context.Background(), "alice")
	if err != nil || len(docs) != 2 {
		t.Fatalf("inputs=%v err=%v", docs, err)
	}
	if settings["novel_fetch_automation"].(map[string]any)["enabled"] != true || settings["novel_fetch_current_batch_id"] != "batch-1" || docs[0].Meta["status"] != "waiting_retry" || docs[1].Versions["ai1"] != "已生成" {
		t.Fatal("status, batch scope or legacy body lost")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
