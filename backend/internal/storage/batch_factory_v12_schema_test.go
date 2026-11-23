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
	if len(migrations) != 16 {
		t.Fatalf("V12 migration count=%d, want 16: %+v", len(migrations), migrations)
	}
	for i, migration := range migrations {
		if int(migration.Version) != 1200101+i {
			t.Fatalf("V12 migration %d has version %d", i, migration.Version)
		}
	}
}

func TestV12TextProductionMigrationCreatesDurablePageAndCollectionTables(t *testing.T) {
	migrations := V12Migrations()
	m := migrations[12]
	if m.Version != 1200113 || m.CallbackChecksum != "batch-factory-v12-text-production-v1" {
		t.Fatalf("missing V12 text production migration: version %d, marker %q", m.Version, m.CallbackChecksum)
	}
	tables := map[string][]string{
		"documents":        {"original_text mediumtext", "active_version_id", "pagination_version", "config_snapshot_json", "cover_artifact_id", "cover_locked", "(owner_username, book_id)"},
		"versions":         {"production_text mediumtext", "sensitive_mode", "change_trace_json", "content_digest"},
		"pages":            {"version_id", "ordinal_no", "start_offset", "end_offset", "page_text", "content_digest", "layout_json", "background_json", "source_kind", "locked", "render_version", "status", "last_error", "revision", "created_at", "updated_at", "unique key uq_bfv12_text_page_order (owner_username, book_id, pagination_version, ordinal_no)", "(owner_username, book_id, status)"},
		"page_artifacts":   {"page_id", "render_version", "storage_ref", "media_type", "sha256"},
		"backgrounds":      {"owner_username", "storage_ref", "media_type"},
		"collections":      {"book_id", "pagination_version", "cover_artifact_id", "status"},
		"collection_items": {"primary key (collection_id, ordinal_no)", "artifact_id", "storage_ref"},
	}
	for _, statement := range m.SQL {
		sql := strings.ToLower(strings.Join(strings.Fields(statement), " "))
		name, _, ok := strings.Cut(strings.TrimPrefix(sql, "create table if not exists batch_factory_v12_text_"), " (")
		required, exists := tables[name]
		if !ok || !exists {
			t.Fatalf("unexpected text migration statement: %s", sql)
		}
		for _, field := range required {
			if !strings.Contains(sql, field) {
				t.Errorf("table %s missing %q", name, field)
			}
		}
		delete(tables, name)
	}
	if len(tables) != 0 {
		t.Fatalf("missing text tables: %v", tables)
	}
}

func TestV12BatchRetentionMigrationAddsCollectionClock(t *testing.T) {
	joined := strings.ToLower(strings.Join(V12BatchRetentionStatements(), "\n"))
	for _, required := range []string{
		"alter table batch_factory_v12_batches",
		"completed_at",
		"retention_expires_at",
		"hidden_at",
		"purge_after",
		"idx_bfv12_batches_retention",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing %q in batch retention migration:\n%s", required, joined)
		}
	}
}

func TestV12BookRetentionMigrationAddsWholeBookClock(t *testing.T) {
	joined := strings.ToLower(strings.Join(V12BookRetentionStatements(), "\n"))
	for _, required := range []string{
		"alter table batch_factory_v12_books",
		"completed_at",
		"retention_expires_at",
		"hidden_at",
		"purge_after",
		"idx_bfv12_books_retention",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing %q in book retention migration:\n%s", required, joined)
		}
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
