package storage

import (
	"context"
	"strings"
	"testing"
)

type memoryLedger map[int]string

func (m memoryLedger) RecordedChecksum(_ context.Context, v int) (string, bool, error) {
	s, ok := m[v]
	return s, ok, nil
}
func (m memoryLedger) Record(_ context.Context, v int, checksum string) error {
	m[v] = checksum
	return nil
}

func TestRunMigrationPlanRejectsRecordedChecksumMismatch(t *testing.T) {
	m := Migration{Version: 26, SQL: []string{"CREATE TABLE x (id BIGINT)"}}
	err := RunMigrationPlan(context.Background(), memoryLedger{26: "wrong"}, []Migration{m}, func(context.Context, Migration) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("err=%v", err)
	}
}

func TestRunMigrationPlanAcceptsOnlyExplicitLegacyChecksum(t *testing.T) {
	m := Migration{
		Version:         26,
		SQL:             []string{"CREATE TABLE x (id BIGINT)"},
		LegacyChecksums: []string{"legacy-verified"},
	}
	called := false
	err := RunMigrationPlan(context.Background(), memoryLedger{26: "legacy-verified"}, []Migration{m}, func(context.Context, Migration) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("an accepted legacy migration must not run again")
	}
}

func TestV11FoundationSchemaHasNoTextDefault(t *testing.T) {
	for _, statement := range V11FoundationStatements() {
		upper := strings.ToUpper(statement)
		if strings.Contains(upper, "TEXT NOT NULL DEFAULT") || strings.Contains(upper, "MEDIUMTEXT NOT NULL DEFAULT") {
			t.Fatal(statement)
		}
	}
}

func TestChecksumStableAcrossWhitespace(t *testing.T) {
	a := checksumFor(Migration{Version: 1, SQL: []string{"CREATE  TABLE x ( id BIGINT );"}})
	b := checksumFor(Migration{Version: 1, SQL: []string{"CREATE TABLE x ( id BIGINT );"}})
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func TestAppMigrationsIncludesV11AndV12Migrations(t *testing.T) {
	migrations := AppMigrations()
	want := append(append(append(append(append(append(append([]Migration(nil), V11Migrations()...), V12Migrations()...), LocalExecutorMigrations()...), GiantMaterialExecutorMigrations()...), AuthMigrations()...), AgentWorkspaceMigrations()...), AutomationMigrations()...)
	if len(migrations) != len(want) {
		t.Fatalf("AppMigrations length=%d, want %d", len(migrations), len(want))
	}
	for i, migration := range migrations {
		if migration.Version != want[i].Version {
			t.Fatalf("migration[%d]=%d, want %d", i, migration.Version, want[i].Version)
		}
	}
}

func TestMergePollerMigrationUsesMySQLCompatibleAddColumnSyntax(t *testing.T) {
	statements := V11MergePollerStatements()
	if len(statements) != 1 {
		t.Fatalf("got %d statements, want 1", len(statements))
	}
	upper := strings.ToUpper(statements[0])
	if strings.Contains(upper, "ADD COLUMN IF NOT EXISTS") {
		t.Fatalf("MySQL compatibility regression: %s", statements[0])
	}
	if !strings.Contains(upper, "ADD COLUMN PROVIDER_TASK_ID") {
		t.Fatalf("provider_task_id migration missing: %s", statements[0])
	}
}

func TestLocalExecutorJobSourceTaskIDSupportsBatchFactoryIdentity(t *testing.T) {
	migrations := LocalExecutorMigrations()
	for _, migration := range migrations {
		for _, statement := range migration.SQL {
			upper := strings.ToUpper(statement)
			if strings.Contains(upper, "MODIFY COLUMN SOURCE_TASK_ID VARCHAR(255)") {
				return
			}
		}
	}
	t.Fatal("local executor migrations must widen source_task_id for batch, book and video identities")
}
