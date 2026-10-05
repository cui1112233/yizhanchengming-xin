package unifiedsettings

import (
	"context"
	"regexp"
	"testing"
	"time"

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

func TestMySQLStorePersistsAndReloadsVersionProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)
	profile := VersionProfile{
		ProjectID: 9,
		Name: "女频短剧版",
		Version: "v3",
		Settings: Settings{
			Website121: map[string]any{"sources": []any{map[string]any{"source": "知乎", "platformId": "4"}}},
			StyleTypes: map[string]any{"styles": []any{"剧情"}},
			ProcessingRulePromptRef: "processing-v3",
			KnowledgePromptRef: "knowledge-v7",
		},
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO batch_version_config_profiles (batch_project_id, profile_name, version, settings_json) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE profile_name = VALUES(profile_name), version = VALUES(version), settings_json = VALUES(settings_json), updated_at = CURRENT_TIMESTAMP(6)")).
		WithArgs(int64(9), "女频短剧版", "v3", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "profile_name", "version", "settings_json", "created_at", "updated_at"}).AddRow(
		int64(3), "女频短剧版", "v3", `{"website121":{"sources":[{"source":"知乎","platformId":"4"}]},"styleTypes":{"styles":["剧情"]},"processingRulePromptRef":"processing-v3","knowledgePromptRef":"knowledge-v7"}`, now, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, profile_name, version, settings_json, created_at, updated_at FROM batch_version_config_profiles WHERE batch_project_id = ?")).WithArgs(int64(9)).WillReturnRows(rows)

	saved, err := store.SaveVersionProfile(context.Background(), profile)
	if err != nil { t.Fatal(err) }
	if saved.Name != "女频短剧版" || saved.Settings.ProcessingRulePromptRef != "processing-v3" || saved.Settings.KnowledgePromptRef != "knowledge-v7" {
		t.Fatalf("saved profile = %#v", saved)
	}
	if len(saved.Settings.Website121) == 0 || len(saved.Settings.StyleTypes) == 0 { t.Fatalf("sync snapshots missing after reload: %#v", saved.Settings) }
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
