package unifiedsettings

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

type integrationSnapshotSource struct{}

func (integrationSnapshotSource) Get121Snapshot(context.Context, int64) (map[string]any, error) {
	return map[string]any{"sources": []map[string]string{{"source": "知乎", "platformId": "4"}}}, nil
}

func (integrationSnapshotSource) GetStyleTypeSnapshot(context.Context, int64) (map[string]any, error) {
	return map[string]any{"styles": []string{"剧情"}, "genres": []string{"都市"}, "genders": []string{"女频"}}, nil
}

func TestMySQLPersistenceSurvivesStoreReload(t *testing.T) {
	dsn := os.Getenv("TASK10_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TASK10_MYSQL_DSN not configured")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil { t.Fatal(err) }
	defer db.Close()
	if err := db.Ping(); err != nil { t.Fatal(err) }

	ctx := context.Background()
	intakeResult, err := db.ExecContext(ctx, `INSERT INTO intakes (name, status) VALUES (?, 'completed')`, "task10-persistence")
	if err != nil { t.Fatal(err) }
	intakeID, err := intakeResult.LastInsertId()
	if err != nil { t.Fatal(err) }
	defer db.ExecContext(ctx, `DELETE FROM intakes WHERE id = ?`, intakeID)

	projectResult, err := db.ExecContext(ctx, `INSERT INTO batch_projects (intake_id, name) VALUES (?, ?)`, intakeID, "task10-project")
	if err != nil { t.Fatal(err) }
	projectID, err := projectResult.LastInsertId()
	if err != nil { t.Fatal(err) }

	store := NewMySQLStore(db)
	projectSettings := Settings{
		Production: map[string]any{"aiCopyEnabled": true, "aiCopyCount": float64(4)},
		Publishing: map[string]any{"uploadVideoType": "individual"},
	}
	if _, err := store.SaveProjectSettings(ctx, projectID, projectSettings); err != nil { t.Fatal(err) }
	profile := VersionProfile{
		ProjectID: projectID,
		Name: "v88 对应档",
		Version: "v88",
		Settings: Settings{
			ProcessingRulePromptRef: "processing-v3",
			KnowledgePromptRef: "knowledge-v7",
		},
	}
	if _, err := store.SaveVersionProfile(ctx, profile); err != nil { t.Fatal(err) }

	service := NewService(store, StaticDefaults{}, integrationSnapshotSource{})
	if _, err := service.Sync121(ctx, projectID); err != nil { t.Fatalf("sync 121: %v", err) }
	if _, err := service.SyncStyleTypes(ctx, projectID); err != nil { t.Fatalf("sync style types: %v", err) }

	// Simulate a page refresh/new request by constructing a fresh store and reading from MySQL again.
	reloaded := NewMySQLStore(db)
	gotProject, err := reloaded.GetProjectSettings(ctx, projectID)
	if err != nil { t.Fatal(err) }
	if !reflect.DeepEqual(gotProject.Production, projectSettings.Production) || !reflect.DeepEqual(gotProject.Publishing, projectSettings.Publishing) {
		t.Fatalf("reloaded project settings = %#v, want %#v", gotProject, projectSettings)
	}
	gotProfile, err := reloaded.GetVersionProfile(ctx, projectID)
	if err != nil { t.Fatal(err) }
	if gotProfile.Name != "v88 对应档" || gotProfile.Version != "v88" {
		t.Fatalf("reloaded profile = %#v", gotProfile)
	}
	if gotProfile.Settings.ProcessingRulePromptRef != "processing-v3" || gotProfile.Settings.KnowledgePromptRef != "knowledge-v7" {
		t.Fatalf("reloaded prompt refs = %#v", gotProfile.Settings)
	}
	if len(gotProfile.Settings.Website121) == 0 {
		t.Fatalf("121 snapshot did not survive MySQL reload: %#v", gotProfile.Settings.Website121)
	}
	if len(gotProfile.Settings.StyleTypes) == 0 {
		t.Fatalf("style/type snapshot did not survive MySQL reload: %#v", gotProfile.Settings.StyleTypes)
	}
}
