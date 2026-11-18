package storage

import (
	"reflect"
	"strings"
	"testing"
)

func TestNovelCreationMigrationCreatesOwnerScopedLongFormDomain(t *testing.T) {
	migrations := NovelCreationMigrations()
	if len(migrations) != 10 {
		t.Fatalf("migrations=%d, want 10", len(migrations))
	}
	migration := migrations[0]
	if migration.Version != 8800300 {
		t.Fatalf("version=%d, want 8800300", migration.Version)
	}
	if migration.CallbackChecksum != "v88-novel-creation-long-form-v1" {
		t.Fatalf("checksum=%q", migration.CallbackChecksum)
	}
	joined := strings.Join(migration.SQL, "\n")
	for _, table := range []string{
		"novel_projects", "novel_sessions", "novel_messages", "novel_volumes",
		"novel_chapters", "novel_chapter_versions", "novel_runs", "novel_run_steps",
		"novel_run_events", "novel_foundations", "novel_memories", "novel_audits", "novel_revisions",
	} {
		if !strings.Contains(joined, "CREATE TABLE IF NOT EXISTS "+table) {
			t.Fatalf("missing table %s", table)
		}
		if !strings.Contains(joined, "idx_"+table+"_owner") {
			t.Fatalf("%s missing owner_username index", table)
		}
	}
	for _, required := range []string{"LONGTEXT", " JSON ", "DATETIME(6)", "VARCHAR(96)"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing required storage type %q", required)
		}
	}
	settings := migrations[1]
	if settings.Version != 8800301 || settings.CallbackChecksum != "v88-novel-creation-project-settings-v1" {
		t.Fatalf("settings migration=%+v", settings)
	}
	settingsSQL := strings.Join(settings.SQL, "\n")
	if !strings.Contains(settingsSQL, "CREATE TABLE IF NOT EXISTS novel_project_settings") {
		t.Fatalf("missing novel_project_settings table")
	}
	if !strings.Contains(settingsSQL, "PRIMARY KEY (owner_username, project_id)") {
		t.Fatalf("novel_project_settings must be owner scoped")
	}
	p2 := migrations[2]
	if p2.Version != 8800302 || p2.CallbackChecksum != "v88-novel-creation-materials-notes-schedules-v1" {
		t.Fatalf("p2 migration=%+v", p2)
	}
	p2SQL := strings.Join(p2.SQL, "\n")
	for _, table := range []string{"novel_materials", "novel_project_notes", "novel_schedules"} {
		if !strings.Contains(p2SQL, "CREATE TABLE IF NOT EXISTS "+table) {
			t.Fatalf("missing table %s", table)
		}
	}
	if !strings.Contains(p2SQL, "ADD COLUMN agent_json JSON NULL") {
		t.Fatalf("missing novel_project_settings.agent_json")
	}
	p3 := migrations[3]
	if p3.Version != 8800303 || p3.CallbackChecksum != "v88-novel-creation-batches-v1" {
		t.Fatalf("p3 migration=%+v", p3)
	}
	p3SQL := strings.Join(p3.SQL, "\n")
	if !strings.Contains(p3SQL, "CREATE TABLE IF NOT EXISTS novel_batches") ||
		!strings.Contains(p3SQL, "completed_chapters INT NOT NULL") ||
		!strings.Contains(p3SQL, "KEY idx_novel_batches_active (status, updated_at)") {
		t.Fatalf("novel_batches schema missing progress columns")
	}
}

func TestNovelCreationMigrationsAreStableAcrossCalls(t *testing.T) {
	if !reflect.DeepEqual(NovelCreationMigrations(), NovelCreationMigrations()) {
		t.Fatal("migration result changed between calls")
	}
}

func TestAppMigrationsIncludeNovelCreationExactlyOnce(t *testing.T) {
	count := 0
	for _, migration := range AppMigrations() {
		if migration.Version == 8800300 {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("novel creation migration count=%d, want 1", count)
	}
}

func TestNovelCreationSchemaKeepsProjectBindingsConsistent(t *testing.T) {
	statementFor := func(table string) string {
		for _, statement := range NovelCreationMigrations()[0].SQL {
			if strings.Contains(statement, "CREATE TABLE IF NOT EXISTS "+table) {
				return statement
			}
		}
		t.Fatalf("missing table %s", table)
		return ""
	}
	for _, constraint := range []string{
		"UNIQUE KEY uq_novel_sessions_project (owner_username, session_id, project_id)",
	} {
		if !strings.Contains(statementFor("novel_sessions"), constraint) {
			t.Fatalf("novel_sessions missing composite project binding: %s", constraint)
		}
	}
	for _, constraint := range []string{
		"FOREIGN KEY (owner_username, session_id, project_id) REFERENCES novel_sessions(owner_username, session_id, project_id)",
	} {
		if !strings.Contains(statementFor("novel_messages"), constraint) {
			t.Fatalf("novel_messages missing composite project binding: %s", constraint)
		}
		if !strings.Contains(statementFor("novel_runs"), constraint) {
			t.Fatalf("novel_runs missing composite project binding: %s", constraint)
		}
	}
	for _, constraint := range []string{
		"UNIQUE KEY uq_novel_runs_session_project (owner_username, run_id, session_id, project_id)",
		"FOREIGN KEY (owner_username, run_id, session_id, project_id) REFERENCES novel_runs(owner_username, run_id, session_id, project_id)",
	} {
		if !strings.Contains(statementFor("novel_runs"), constraint) && !strings.Contains(statementFor("novel_run_events"), constraint) {
			t.Fatalf("missing composite project binding: %s", constraint)
		}
	}
}

func TestNovelCreationAutomationInputMigration(t *testing.T) {
	migrations := NovelCreationMigrations()
	m := migrations[4]
	if m.Version != 8800311 {
		t.Fatalf("migration=%+v", m)
	}
	sql := strings.Join(m.SQL, "\n")
	for _, part := range []string{"ALTER TABLE novel_schedules ADD COLUMN input_snapshot_json JSON NULL", "ALTER TABLE novel_batches ADD COLUMN input_snapshot_json JSON NULL", "CREATE TABLE IF NOT EXISTS novel_style_profiles", "PRIMARY KEY (owner_username, style_profile_id)"} {
		if !strings.Contains(sql, part) {
			t.Fatalf("missing %s", part)
		}
	}
	seen := map[int]bool{}
	for _, m := range AppMigrations() {
		if seen[m.Version] {
			t.Fatalf("duplicate migration %d", m.Version)
		}
		seen[m.Version] = true
	}
}

func TestNovelWorkerMigrationUsesDurableDispatchAndProjectSlot(t *testing.T) {
	migrations := NovelCreationMigrations()
	m := migrations[5]
	if m.Version != 8800312 {
		t.Fatalf("worker migration=%d", m.Version)
	}
	sql := strings.Join(m.SQL, "\n")
	for _, part := range []string{"novel_run_dispatch", "execution_json JSON NOT NULL", "PRIMARY KEY (owner_username, run_id)", "novel_project_slots", "fence BIGINT NOT NULL", "lease_until DATETIME(6)", "PRIMARY KEY (owner_username, project_id)"} {
		if !strings.Contains(sql, part) {
			t.Fatalf("missing %s", part)
		}
	}
}

func TestNovelAutomationRoundMigrationUniquePlanAndRun(t *testing.T) {
	m := NovelCreationMigrations()[6]
	if m.Version != 8800313 {
		t.Fatal(m.Version)
	}
	sql := strings.Join(m.SQL, "\n")
	for _, part := range []string{"PRIMARY KEY (owner_username, kind, plan_id, round_key)", "UNIQUE KEY uq_novel_automation_run (owner_username, run_id)", "planned_at DATETIME(6)", "REFERENCES novel_runs(owner_username, run_id)"} {
		if !strings.Contains(sql, part) {
			t.Fatal("missing " + part)
		}
	}
}
