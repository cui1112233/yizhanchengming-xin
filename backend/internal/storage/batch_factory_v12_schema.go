package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// V12 keeps its upgrade lineage outside V11 tables. The immutable snapshot is
// the audit anchor for a legacy batch opened by the V12 workbench; source V11
// rows are never rewritten as part of that transition.
func V12UpgradeAuditStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v12_upgrade_audits (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  source_v11_batch_id VARCHAR(64) NOT NULL,
  target_v12_batch_id VARCHAR(64) NULL,
  status VARCHAR(32) NOT NULL,
  snapshot_json JSON NOT NULL,
  detail TEXT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_bfv12_upgrade_source (owner_username, source_v11_batch_id),
  KEY idx_bfv12_upgrade_target (owner_username, target_v12_batch_id, updated_at)
) ENGINE=InnoDB`,
	}
}

func V12H3AggregateStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v12_audio_measurements (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NOT NULL,
  audio_asset_id VARCHAR(64) NOT NULL,
  audio_content_hash VARCHAR(128) NOT NULL,
  audio_duration_ms BIGINT NOT NULL,
  video_source_revision VARCHAR(128) NOT NULL,
  video_source_hash VARCHAR(128) NOT NULL,
  probe_key VARCHAR(64) NOT NULL,
  probe_version VARCHAR(64) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_bfv12_audio_content (owner_username, book_id, audio_asset_id, audio_content_hash),
  KEY idx_bfv12_audio_latest (owner_username, batch_id, book_id, audio_asset_id, created_at),
  CONSTRAINT fk_bfv12_audio_batch FOREIGN KEY (batch_id) REFERENCES batch_factory_v11_batches(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bfv12_audio_book FOREIGN KEY (book_id) REFERENCES batch_factory_v11_books(id) ON DELETE RESTRICT
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v12_canonical_timelines (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NOT NULL,
  director_revision_id VARCHAR(64) NOT NULL,
  audio_asset_id VARCHAR(64) NOT NULL,
  audio_content_hash VARCHAR(128) NOT NULL,
  audio_duration_ms BIGINT NOT NULL,
  allocator_version VARCHAR(64) NOT NULL,
  timeline_json JSON NOT NULL,
  input_hash CHAR(64) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_bfv12_timeline_input (owner_username, book_id, input_hash),
  KEY idx_bfv12_timeline_owner_book (owner_username, batch_id, book_id, created_at),
  CONSTRAINT fk_bfv12_timeline_batch FOREIGN KEY (batch_id) REFERENCES batch_factory_v11_batches(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bfv12_timeline_book FOREIGN KEY (book_id) REFERENCES batch_factory_v11_books(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bfv12_timeline_director FOREIGN KEY (director_revision_id) REFERENCES batch_factory_v11_director_revisions(id) ON DELETE RESTRICT
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v12_video_compilations (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NOT NULL,
  director_revision_id VARCHAR(64) NOT NULL,
  canonical_timeline_id VARCHAR(64) NOT NULL,
  video_preset_key VARCHAR(128) NOT NULL,
  video_preset_revision BIGINT NOT NULL,
  video_preset_snapshot JSON NOT NULL,
  compiler_key VARCHAR(64) NOT NULL,
  compiler_version VARCHAR(64) NOT NULL,
  max_segment_ms BIGINT NOT NULL,
  compilation_json JSON NOT NULL,
  input_hash CHAR(64) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_bfv12_compilation_input (owner_username, book_id, input_hash),
  KEY idx_bfv12_compilation_owner_book (owner_username, batch_id, book_id, created_at),
  CONSTRAINT fk_bfv12_compilation_batch FOREIGN KEY (batch_id) REFERENCES batch_factory_v11_batches(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bfv12_compilation_book FOREIGN KEY (book_id) REFERENCES batch_factory_v11_books(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bfv12_compilation_director FOREIGN KEY (director_revision_id) REFERENCES batch_factory_v11_director_revisions(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bfv12_compilation_timeline FOREIGN KEY (canonical_timeline_id) REFERENCES batch_factory_v12_canonical_timelines(id) ON DELETE RESTRICT
) ENGINE=InnoDB`,
	}
}

func V12H3ProductionTraceStatements() []string {
	return []string{
		`ALTER TABLE batch_factory_v11_production_tasks ADD COLUMN compilation_id VARCHAR(64) NULL AFTER video_id`,
		`ALTER TABLE batch_factory_v11_production_tasks ADD COLUMN compilation_segment_key VARCHAR(64) NULL AFTER compilation_id`,
		`ALTER TABLE batch_factory_v11_production_tasks ADD COLUMN compile_trace_json JSON NULL AFTER compiled_prompt`,
		`ALTER TABLE batch_factory_v11_production_tasks ADD KEY idx_bfv11_production_compilation (owner_username, compilation_id, compilation_segment_key)`,
		`ALTER TABLE batch_factory_v11_production_tasks ADD CONSTRAINT fk_bfv11_production_compilation FOREIGN KEY (compilation_id) REFERENCES batch_factory_v12_video_compilations(id) ON DELETE RESTRICT`,
	}
}

func V12H3KernelStatements() []string {
	statements := append([]string(nil), V12H3AggregateStatements()...)
	return append(statements, V12H3ProductionTraceStatements()...)
}

// V12SlotDigestStatements adds the per-slot configuration hash used by the
// Director rotation guard. Each storyboard slot records the digest of the
// frozen inputs that produced it, so a later revision only orphans the slots
// whose inputs actually changed instead of rotating the whole book.
func V12SlotDigestStatements() []string {
	return []string{
		`ALTER TABLE batch_factory_v12_director_video_links ADD COLUMN slot_digest CHAR(64) NOT NULL DEFAULT '' AFTER ordinal`,
		`ALTER TABLE batch_factory_v12_director_video_links ADD KEY idx_bfv12_links_revision_ordinal (director_revision_id, ordinal)`,
	}
}

// V12AutomationPresetStatements versions the Batch Factory automation presets.
// Each update writes an immutable revision row and moves the head pointer, so
// a batch pinned to (preset_id, preset_revision) keeps an auditable config
// snapshot instead of silently following the latest overwrite.
func V12AutomationPresetStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v12_automation_presets (
			id VARCHAR(64) NOT NULL,
			owner_username VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			current_revision INT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			KEY idx_bfv12_presets_owner (owner_username)
		)`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v12_automation_preset_revisions (
			preset_id VARCHAR(64) NOT NULL,
			owner_username VARCHAR(255) NOT NULL,
			revision INT NOT NULL,
			name VARCHAR(255) NOT NULL,
			config_json JSON NOT NULL,
			created_at DATETIME NOT NULL,
			PRIMARY KEY (preset_id, revision),
			KEY idx_bfv12_preset_revisions_owner (owner_username)
		)`,
	}
}

func V12Migrations() []Migration {
	// 1200001-1200003 belong to Novel Fetch Workshop in the shared migration
	// ledger. V12 must have its own unused number or a Batch Factory startup
	// would reject that unrelated, already-applied migration by checksum.
	return []Migration{
		{Version: 1200101, SQL: V12UpgradeAuditStatements(), CallbackChecksum: "batch-factory-v12-upgrade-audit-v1"},
		{Version: 1200102, SQL: V12H3AggregateStatements(), CallbackChecksum: "batch-factory-v12-h3-aggregates-v1"},
		{Version: 1200103, SQL: V12H3ProductionTraceStatements(), CallbackChecksum: "batch-factory-v12-h3-production-trace-v1"},
		{Version: 1200104, SQL: []string{`ALTER TABLE batch_factory_v12_audio_measurements ADD COLUMN measurement_json JSON NULL`}, CallbackChecksum: "batch-factory-v12-line-audio-v1"},
		{Version: 1200105, SQL: V12CanonicalTableRenameStatements(), CallbackChecksum: "batch-factory-v12-canonical-tables-v1", Adopt: adoptV12CanonicalTableRename},
		{Version: 1200106, SQL: V12SlotDigestStatements(), CallbackChecksum: "batch-factory-v12-slot-digest-v1"},
		{Version: 1200107, SQL: V12AutomationPresetStatements(), CallbackChecksum: "batch-factory-v12-automation-presets-v1"},
	}
}

// V12CanonicalTableRenameStatements performs the one-way physical cutover of
// the old table names. The rows stay in the same tables; only the names used
// by the active V12 code change. Keeping this as a migration, instead of
// creating copies, avoids doubling the database and makes foreign keys follow
// the original rows.
func V12CanonicalTableRenameStatements() []string {
	tables := []string{
		"bridge_users", "batches", "books", "videos", "migration_audits", "capability_settings",
		"intakes", "batch_records", "book_records", "video_records", "settings_patches",
		"config_versions", "config_snapshots", "prompt_definitions", "prompt_versions", "drafts",
		"hook_revisions", "director_revisions", "director_video_links", "orphaned_overrides",
		"production_jobs", "production_tasks", "production_events", "hidden_production_tasks",
		"merge_jobs", "merge_sources", "external_credentials", "external_intents", "external_audits",
		"comic_pages", "comic_page_images", "comic_collections", "comic_collection_items", "comic_source_segments",
		"book_assets", "book_asset_images", "book_stage_runs", "local_artifact_purges", "shuihuo_configs",
	}
	statements := make([]string, 0, len(tables))
	for _, table := range tables {
		statements = append(statements, "RENAME TABLE batch_factory_v11_"+table+" TO batch_factory_v12_"+table)
	}
	return statements
}

// Adopt makes a previously completed physical cutover safe to restart after
// the migration ledger write was interrupted. It does not modify data.
func adoptV12CanonicalTableRename(ctx context.Context, tx *sql.Tx) (bool, error) {
	for _, statement := range V12CanonicalTableRenameStatements() {
		parts := strings.Fields(statement)
		if len(parts) != 5 {
			return false, fmt.Errorf("unexpected canonical rename statement: %s", statement)
		}
		var sourceCount, targetCount int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?`, parts[2]).Scan(&sourceCount); err != nil {
			return false, err
		}
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?`, parts[4]).Scan(&targetCount); err != nil {
			return false, err
		}
		if sourceCount > 0 {
			return false, nil
		}
		if targetCount == 0 {
			return false, fmt.Errorf("canonical table missing after partial cutover: %s", parts[4])
		}
	}
	return true, nil
}
