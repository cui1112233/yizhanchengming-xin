package storage

import (
	"strings"
	"testing"
)

func TestAICutNamedTimelineMigrationPreservesLegacyTables(t *testing.T) {
	found := false
	for _, m := range AICutMigrations() {
		if m.Version != 8800030 {
			continue
		}
		found = true
		if len(m.SQL) != 2 {
			t.Fatalf("statements=%d", len(m.SQL))
		}
		joined := strings.Join(m.SQL, "\n")
		for _, want := range []string{"CREATE TABLE IF NOT EXISTS aicut_timelines", "CREATE TABLE IF NOT EXISTS aicut_named_timeline_history", "REFERENCES aicut_projects(id, owner_username)", "REFERENCES aicut_timelines(project_id, owner_username, id)"} {
			if !strings.Contains(joined, want) {
				t.Fatalf("missing %q", want)
			}
		}
		for _, forbidden := range []string{"ALTER TABLE", "UPDATE aicut_projects", "DELETE FROM", "DROP TABLE"} {
			if strings.Contains(joined, forbidden) {
				t.Fatalf("unexpected legacy mutation %q", forbidden)
			}
		}
	}
	if !found {
		t.Fatal("missing named timeline migration")
	}
}

func TestAICutASRTimelineMigrationKeepsLegacyDefault(t *testing.T) {
	for _, m := range AICutMigrations() {
		if m.Version == 8800031 {
			sql := strings.Join(m.SQL, "\n")
			if !strings.Contains(sql, "ALTER TABLE aicut_asr_jobs ADD COLUMN timeline_id VARCHAR(64) NOT NULL DEFAULT ''") {
				t.Fatalf("sql=%s", sql)
			}
			return
		}
	}
	t.Fatal("missing ASR target migration")
}

func TestAICutAgentTimelineMigrationAddsStableTarget(t *testing.T) {
	for _, m := range AICutMigrations() {
		if m.Version == 8800032 {
			if len(m.SQL) != 1 || !strings.Contains(m.SQL[0], "ALTER TABLE aicut_agent_runs ADD COLUMN timeline_id VARCHAR(64) NOT NULL DEFAULT ''") {
				t.Fatalf("sql=%v", m.SQL)
			}
			return
		}
	}
	t.Fatal("missing Agent target migration")
}
