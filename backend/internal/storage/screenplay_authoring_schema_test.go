package storage

import (
	"strings"
	"testing"
)

func TestScreenplayAuthoringSchema(t *testing.T) {
	ms := ScreenplayAuthoringMigrations()
	if len(ms) != 1 || ms[0].Version != 8800107 {
		t.Fatal(ms)
	}
	sql := strings.Join(ms[0].SQL, "\n")
	for _, v := range []string{"screenplay_authoring", "owner_username", "project_id", "revision", "body JSON", "PRIMARY KEY(project_id,owner_username)"} {
		if !strings.Contains(sql, v) {
			t.Fatal(v)
		}
	}
}
