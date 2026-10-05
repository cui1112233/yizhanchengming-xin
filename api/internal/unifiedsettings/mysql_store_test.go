package unifiedsettings

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStorePersistsProjectSettings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)
	settings := Settings{Production: map[string]any{"aiCopyEnabled": true, "aiCopyCount": float64(3)}, ProcessingRulePromptRef: "processing-v3"}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO batch_project_settings (batch_project_id, settings_json) VALUES (?, ?) ON DUPLICATE KEY UPDATE settings_json = VALUES(settings_json), updated_at = CURRENT_TIMESTAMP(6)")).WithArgs(int64(9), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	if _, err := store.SaveProjectSettings(context.Background(), 9, settings); err != nil { t.Fatal(err) }

	rows := sqlmock.NewRows([]string{"settings_json"}).AddRow(`{"production":{"aiCopyCount":3,"aiCopyEnabled":true},"processingRulePromptRef":"processing-v3"}`)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT settings_json FROM batch_project_settings WHERE batch_project_id = ?")).WithArgs(int64(9)).WillReturnRows(rows)
	loaded, err := store.GetProjectSettings(context.Background(), 9)
	if err != nil { t.Fatal(err) }
	if loaded.Production["aiCopyEnabled"] != true || loaded.ProcessingRulePromptRef != "processing-v3" { t.Fatalf("loaded = %#v", loaded) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestMySQLStoreReturnsEmptySettingsWhenProjectHasNoOverride(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT settings_json FROM batch_project_settings WHERE batch_project_id = ?")).WithArgs(int64(20)).WillReturnRows(sqlmock.NewRows([]string{"settings_json"}))
	loaded, err := store.GetProjectSettings(context.Background(), 20)
	if err != nil { t.Fatal(err) }
	if len(loaded.Production) != 0 || len(loaded.Publishing) != 0 { t.Fatalf("loaded = %#v", loaded) }
}
