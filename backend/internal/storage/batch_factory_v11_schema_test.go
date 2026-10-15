package storage

import (
	"strings"
	"testing"
)

func TestV11SliceOneSchemaRegistersSettingsSnapshotPromptAndIntakeTables(t *testing.T) {
	joined := strings.ToLower(strings.Join(V11SliceOneStatements(), "\n"))
	for _, table := range []string{
		"batch_factory_v11_intakes",
		"batch_factory_v11_settings_patches",
		"batch_factory_v11_config_versions",
		"batch_factory_v11_config_snapshots",
		"batch_factory_v11_prompt_definitions",
		"batch_factory_v11_prompt_versions",
		"batch_factory_v11_drafts",
	} {
		if !strings.Contains(joined, table) {
			t.Fatalf("missing %s", table)
		}
	}
}

func TestV11MigrationsAreOrderedFoundationThenSliceOneClosures(t *testing.T) {
	got := V11Migrations()
	if len(got) < 4 || got[0].Version != 1100001 || got[1].Version != 1100002 || got[2].Version != 1100003 || got[3].Version != 1100004 {
		t.Fatalf("migrations=%+v", got)
	}
}

func TestV11SliceOneSchemaHasNoTextDefault(t *testing.T) {
	for _, stmt := range V11SliceOneStatements() {
		u := strings.ToUpper(stmt)
		if strings.Contains(u, "TEXT NOT NULL DEFAULT") || strings.Contains(u, "MEDIUMTEXT NOT NULL DEFAULT") {
			t.Fatal(stmt)
		}
	}
}

func TestV11SliceOneSchemaLinksAggregateParents(t *testing.T) {
	joined := strings.ToLower(strings.Join(V11SliceOneStatements(), "\n"))
	for _, clause := range []string{
		"foreign key (batch_id) references batch_factory_v11_batches(id)",
		"foreign key (book_id) references batch_factory_v11_books(id)",
		"foreign key (video_id) references batch_factory_v11_videos(id)",
		"foreign key (prompt_id) references batch_factory_v11_prompt_definitions(id)",
	} {
		if !strings.Contains(joined, clause) {
			t.Fatalf("missing %s", clause)
		}
	}
}

func TestV11SliceOneKeepsFoundationChecksumStable(t *testing.T) {
	foundation := V11FoundationMigrations()
	all := V11Migrations()
	if len(foundation) != 1 || len(all) < 2 {
		t.Fatalf("foundation=%d all=%d", len(foundation), len(all))
	}
	if ChecksumFor(foundation[0]) != ChecksumFor(all[0]) {
		t.Fatal("Slice 1 rewrote the Foundation migration checksum")
	}
	if got := len(ChecksumFor(all[1])); got != 64 {
		t.Fatalf("Slice 1 checksum length=%d", got)
	}
}

func TestV11MigrationsSeparateVideoAndVisualPrompts(t *testing.T) {
	migrations := V11Migrations()
	if len(migrations) < 11 || migrations[10].Version != 1100011 || migrations[10].CallbackChecksum != "batch-factory-v11-separate-video-prompt-v1" {
		t.Fatalf("last migration=%+v", migrations)
	}
	if got := migrations[10].LegacyChecksums; len(got) != 1 || got[0] != "0aa0615fbe3d1ffa5d13faee0acc76ec39c0b1232c687630ade52e9f041ab329" {
		t.Fatalf("legacy prompt-split checksum=%#v", got)
	}
	if migrations[10].Reconcile == nil {
		t.Fatal("prompt-split migration must repair its verified historical ledger/schema split")
	}
	statements := V11SeparateVideoPromptStatements()
	if len(statements) != 3 || !strings.Contains(statements[0], "video_prompt") {
		t.Fatalf("prompt separation migration=%+v", statements)
	}
}

func TestV11ShuihuoConfigAcceptsVerifiedPublicChecksum(t *testing.T) {
	migrations := V11Migrations()
	var migration *Migration
	for i := range migrations {
		if migrations[i].Version == 1100023 {
			migration = &migrations[i]
			break
		}
	}
	if migration == nil {
		t.Fatal("shuihuo config migration missing")
	}
	if !checksumMatches(*migration, "27ba1a709f19049ce933a5be670e703c6bef140999611daa0757d64952fc5b20", ChecksumFor(*migration)) {
		t.Fatal("verified public checksum must be accepted")
	}
}
