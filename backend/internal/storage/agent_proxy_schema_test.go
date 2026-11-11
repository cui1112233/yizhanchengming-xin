package storage

import (
	"strings"
	"testing"
)

func TestAgentProxyMigrationIsAppendOnlyAndComplete(t *testing.T) {
	migrations := AgentProxyMigrations()
	if len(migrations) != 1 {
		t.Fatalf("AgentProxyMigrations length=%d, want 1", len(migrations))
	}
	migration := migrations[0]
	if migration.Version != 8800200 {
		t.Fatalf("version=%d, want 8800200", migration.Version)
	}
	if migration.CallbackChecksum != "v88-agent-proxy-skeleton-v1" {
		t.Fatalf("checksum marker=%q", migration.CallbackChecksum)
	}
	schema := strings.ToLower(strings.Join(migration.SQL, "\n"))
	for _, table := range []string{
		"agent_proxy_workspaces",
		"agent_proxy_sessions",
		"agent_proxy_messages",
		"agent_proxy_skills",
		"agent_proxy_scheduled_jobs",
	} {
		if !strings.Contains(schema, "create table if not exists "+table) {
			t.Errorf("missing idempotent table %s", table)
		}
	}
	if strings.Contains(schema, "agent_projects") || strings.Contains(schema, "agent_messages") {
		t.Error("agent proxy schema must not reuse or alter the existing agent workspace tables")
	}
}

func TestAgentProxyMigrationDoesNotCollideWithExistingVersions(t *testing.T) {
	seen := map[int]string{}
	for _, migration := range append(AppMigrations(), NovelFetchWorkshopMigrations()...) {
		seen[migration.Version] = migration.CallbackChecksum
	}
	for _, migration := range AgentProxyMigrations() {
		if existing, ok := seen[migration.Version]; ok {
			t.Fatalf("version %d already used by %s", migration.Version, existing)
		}
	}
}
