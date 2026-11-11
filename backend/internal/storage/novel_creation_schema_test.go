package storage

import (
	"reflect"
	"strings"
	"testing"
)

func TestNovelCreationMigrationCreatesOwnerScopedLongFormDomain(t *testing.T) {
	migrations := NovelCreationMigrations()
	if len(migrations) != 1 {
		t.Fatalf("migrations=%d, want 1", len(migrations))
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
