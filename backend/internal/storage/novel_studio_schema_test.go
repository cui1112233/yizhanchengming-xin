package storage

import (
	"strings"
	"testing"
)

func TestNovelStudioMigrationIsSeparateOwnerScopedAndHasNoProjectRequirement(t *testing.T) {
	ms := NovelCreationMigrations()
	m := ms[len(ms)-1]
	if m.Version != 8800316 || m.CallbackChecksum != "v88-novel-studio-sessions-v1" {
		t.Fatal("studio migration missing")
	}
	sql := strings.Join(m.SQL, "\n")
	for _, part := range []string{"CREATE TABLE IF NOT EXISTS novel_studio_sessions", "PRIMARY KEY (owner_username, studio_id)", "UNIQUE KEY uq_novel_studio_creation (owner_username, create_key)", "record_json JSON NOT NULL"} {
		if !strings.Contains(sql, part) {
			t.Fatal("missing studio constraint", part)
		}
	}
	if strings.Contains(sql, "FOREIGN KEY") || strings.Contains(sql, "ALTER TABLE novel_sessions") {
		t.Fatal("studio must not require or weaken old project bindings")
	}
}
