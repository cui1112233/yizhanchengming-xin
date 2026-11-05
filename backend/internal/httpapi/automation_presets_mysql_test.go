package httpapi

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLAutomationPresetStoreUpdateConflictKeepsHistoryUntouched(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	store, err := NewMySQLAutomationPresetStore(db, "", func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT name,current_revision,created_at FROM batch_factory_v12_automation_presets`).
		WithArgs("alice", "preset-1").
		WillReturnRows(sqlmock.NewRows([]string{"name", "current_revision", "created_at"}).AddRow("预设", 3, now))
	mock.ExpectRollback()

	if _, err := store.Update("alice", "preset-1", 2, "预设", map[string]any{"productionMode": "comic"}); !errors.Is(err, ErrAutomationPresetConflict) {
		t.Fatalf("Update err=%v, want version conflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLAutomationPresetStoreUpdateWritesImmutableRevision(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	store, err := NewMySQLAutomationPresetStore(db, "", func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT name,current_revision,created_at FROM batch_factory_v12_automation_presets`).
		WithArgs("alice", "preset-1").
		WillReturnRows(sqlmock.NewRows([]string{"name", "current_revision", "created_at"}).AddRow("预设", 1, now))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO batch_factory_v12_automation_preset_revisions(preset_id,owner_username,revision,name,config_json,created_at) VALUES(?,?,?,?,?,?)`)).
		WithArgs("preset-1", "alice", 2, "新名字", sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE batch_factory_v12_automation_presets SET name=?,current_revision=?,updated_at=? WHERE owner_username=? AND id=?`)).
		WithArgs("新名字", 2, now, "alice", "preset-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	updated, err := store.Update("alice", "preset-1", 1, "新名字", map[string]any{"productionMode": "comic"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 2 || updated.Name != "新名字" || updated.Config["productionMode"] != "comic" {
		t.Fatalf("updated=%+v, want revision 2 with the new config", updated)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLAutomationPresetStoreImportsLegacyJSONOnce(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	legacyPath := filepath.Join(t.TempDir(), "presets.json")
	legacy := automationPresetFile{Version: 1, Presets: map[string][]AutomationPreset{
		"alice": {{ID: "preset-1", Name: "旧预设", Version: 4, Config: map[string]any{"productionMode": "comic"}, CreatedAt: now.Format(time.RFC3339Nano), UpdatedAt: now.Format(time.RFC3339Nano)}},
	}}
	raw, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(`SELECT 1 FROM batch_factory_v12_automation_presets`).
		WithArgs("preset-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"1"}))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO batch_factory_v12_automation_presets`).
		WithArgs("preset-1", "alice", "旧预设", 4, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO batch_factory_v12_automation_preset_revisions`).
		WithArgs("preset-1", "alice", 4, "旧预设", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	store, err := NewMySQLAutomationPresetStore(db, legacyPath, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	if store == nil {
		t.Fatal("store is nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
