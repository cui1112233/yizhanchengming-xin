package storage

import (
	"strings"
	"testing"
)

func TestV12MigrationsAddAnIndependentUpgradeAudit(t *testing.T) {
	migrations := V12Migrations()
	if len(migrations) < 1 || migrations[0].Version != 1200101 {
		t.Fatalf("V12 migrations=%+v", migrations)
	}
	joined := strings.ToLower(strings.Join(migrations[0].SQL, "\n"))
	for _, required := range []string{"batch_factory_v12_upgrade_audits", "source_v11_batch_id", "snapshot_json", "unique key"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing %q in V12 migration: %s", required, joined)
		}
	}
}

func TestV12H3KernelSchemaPersistsTimelineCompilationAndProductionTrace(t *testing.T) {
	joined := strings.ToLower(strings.Join(V12H3KernelStatements(), "\n"))
	for _, required := range []string{
		"batch_factory_v12_audio_measurements",
		"video_source_revision",
		"video_source_hash",
		"probe_key",
		"batch_factory_v12_canonical_timelines",
		"director_revision_id",
		"audio_asset_id",
		"audio_content_hash",
		"audio_duration_ms",
		"allocator_version",
		"timeline_json",
		"input_hash",
		"batch_factory_v12_video_compilations",
		"canonical_timeline_id",
		"video_preset_key",
		"video_preset_revision",
		"video_preset_snapshot",
		"compiler_version",
		"compilation_json",
		"compilation_id",
		"compilation_segment_key",
		"compile_trace_json",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing %q in V12 H3 migrations:\n%s", required, joined)
		}
	}
}

func TestV12MigrationsRegisterH3KernelAfterUpgradeAudit(t *testing.T) {
	migrations := V12Migrations()
	if len(migrations) != 5 {
		t.Fatalf("V12 migration count=%d, want 4: %+v", len(migrations), migrations)
	}
	if migrations[0].Version != 1200101 || migrations[1].Version != 1200102 || migrations[2].Version != 1200103 || migrations[3].Version != 1200104 || migrations[4].Version != 1200105 {
		t.Fatalf("V12 migration order=%+v", migrations)
	}
}
