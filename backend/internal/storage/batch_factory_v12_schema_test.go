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
	if len(migrations) != 10 {
		t.Fatalf("V12 migration count=%d, want 10: %+v", len(migrations), migrations)
	}
	if migrations[0].Version != 1200101 || migrations[1].Version != 1200102 || migrations[2].Version != 1200103 || migrations[3].Version != 1200104 || migrations[4].Version != 1200105 || migrations[5].Version != 1200106 || migrations[6].Version != 1200107 || migrations[7].Version != 1200108 || migrations[8].Version != 1200109 || migrations[9].Version != 1200110 {
		t.Fatalf("V12 migration order=%+v", migrations)
	}
}

func TestV12ComicPageLastErrorMigrationAddsTextColumn(t *testing.T) {
	joined := strings.ToLower(strings.Join(V12ComicPageLastErrorStatements(), "\n"))
	for _, required := range []string{"alter table batch_factory_v12_comic_pages", "add column last_error text null", "after status"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing %q in comic page last_error migration:\n%s", required, joined)
		}
	}
}

func TestV12SlotDigestMigrationAddsPerSlotHash(t *testing.T) {
	joined := strings.ToLower(strings.Join(V12SlotDigestStatements(), "\n"))
	for _, required := range []string{"batch_factory_v12_director_video_links", "slot_digest", "char(64)"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing %q in slot digest migration:\n%s", required, joined)
		}
	}
}

func TestV12ImageGenerationCacheSchemaPreventsDuplicatePaidRequests(t *testing.T) {
	joined := strings.ToLower(strings.Join(V12ImageGenerationCacheStatements(), "\n"))
	for _, required := range []string{
		"batch_factory_v12_image_generation_cache",
		"owner_username",
		"fingerprint",
		"storage_ref",
		"lease_until",
		"queue_ms",
		"header_ms",
		"body_ms",
		"download_ms",
		"save_ms",
		"total_ms",
		"unique key uq_bfv12_image_cache_fingerprint",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing %q in image generation cache schema:\n%s", required, joined)
		}
	}
}

func TestV12ProductionMediaSchemaProtectsVerifiedRemoteCopies(t *testing.T) {
	joined := strings.ToLower(strings.Join(V12ProductionMediaStatements(), "\n"))
	for _, required := range []string{
		"batch_factory_v12_media_objects",
		"owner_username",
		"local_storage_ref",
		"object_key",
		"sha256",
		"remote_verified_at",
		"purge_after",
		"unique key",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing %q in production media schema:\n%s", required, joined)
		}
	}
}
