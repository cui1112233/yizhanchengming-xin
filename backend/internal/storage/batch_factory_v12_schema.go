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

// V12ProductionMediaStatements adds the durable hand-off ledger between the
// local production cache and TOS. A local file is never cleanup-eligible until
// remote_verified_at is populated after byte-size and checksum verification.
func V12ProductionMediaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v12_media_objects (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  media_kind VARCHAR(32) NOT NULL,
  media_type VARCHAR(96) NOT NULL DEFAULT 'application/octet-stream',
  local_storage_ref VARCHAR(512) NOT NULL,
  object_key VARCHAR(768) NOT NULL,
  byte_size BIGINT NULL,
  sha256 CHAR(64) NULL,
  state VARCHAR(32) NOT NULL DEFAULT 'pending',
  attempt_count INT NOT NULL DEFAULT 0,
  retention_days INT NOT NULL DEFAULT 7,
  last_error VARCHAR(512) NULL,
  remote_verified_at DATETIME(6) NULL,
  purge_after DATETIME(6) NULL,
  local_deleted_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_bfv12_media_owner_ref (owner_username, local_storage_ref),
  UNIQUE KEY uq_bfv12_media_object_key (object_key),
  KEY idx_bfv12_media_upload (state, updated_at, id),
  KEY idx_bfv12_media_purge (state, purge_after, id),
  KEY idx_bfv12_media_owner_kind (owner_username, media_kind, created_at)
) ENGINE=InnoDB`,
	}
}

// V12ImageGenerationCacheStatements stores only successful artifact references
// and short-lived generation leases. It never stores credentials or full
// prompts; the SHA-256 fingerprint is calculated before persistence.
func V12ImageGenerationCacheStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v12_image_generation_cache (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  fingerprint CHAR(64) NOT NULL,
  model_id VARCHAR(191) NOT NULL,
  state VARCHAR(24) NOT NULL,
  storage_ref VARCHAR(512) NULL,
  media_type VARCHAR(96) NULL,
  lease_until DATETIME(6) NULL,
  queue_ms BIGINT NOT NULL DEFAULT 0,
  header_ms BIGINT NOT NULL DEFAULT 0,
  body_ms BIGINT NOT NULL DEFAULT 0,
  download_ms BIGINT NOT NULL DEFAULT 0,
  save_ms BIGINT NOT NULL DEFAULT 0,
  total_ms BIGINT NOT NULL DEFAULT 0,
  last_error VARCHAR(512) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_bfv12_image_cache_fingerprint (owner_username, fingerprint),
  KEY idx_bfv12_image_cache_lease (state, lease_until, updated_at),
  KEY idx_bfv12_image_cache_storage_ref (storage_ref)
) ENGINE=InnoDB`,
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
		{Version: 1200106, SQL: V12ProductionMediaStatements(), CallbackChecksum: "batch-factory-v12-production-media-v1"},
		{Version: 1200107, SQL: V12ImageGenerationCacheStatements(), CallbackChecksum: "batch-factory-v12-image-generation-cache-v1"},
		{Version: 1200108, SQL: V12SlotDigestStatements(), CallbackChecksum: "batch-factory-v12-slot-digest-v1"},
		{Version: 1200109, SQL: V12AutomationPresetStatements(), CallbackChecksum: "batch-factory-v12-automation-presets-v1"},
		{Version: 1200110, SQL: V12ComicPageLastErrorStatements(), CallbackChecksum: "batch-factory-v12-comic-page-last-error-v1"},
		{Version: 1200111, SQL: V12BookRetentionStatements(), CallbackChecksum: "batch-factory-v12-book-retention-v1"},
	}
}

// V12BookRetentionStatements adds the whole-book retention clock. A book is the
// cleanup unit: once its last production activity is older than the configured
// retention window it is hidden first, then permanently removed together with
// its TOS objects and local cache after a 24-hour grace period.
func V12BookRetentionStatements() []string {
	return []string{
		`ALTER TABLE batch_factory_v12_books ADD COLUMN completed_at DATETIME(6) NULL AFTER updated_at`,
		`ALTER TABLE batch_factory_v12_books ADD COLUMN retention_expires_at DATETIME(6) NULL AFTER completed_at`,
		`ALTER TABLE batch_factory_v12_books ADD COLUMN hidden_at DATETIME(6) NULL AFTER retention_expires_at`,
		`ALTER TABLE batch_factory_v12_books ADD COLUMN purge_after DATETIME(6) NULL AFTER hidden_at`,
		`ALTER TABLE batch_factory_v12_books ADD KEY idx_bfv12_books_retention (owner_username, hidden_at, retention_expires_at)`,
	}
}

// V12ComicPageLastErrorStatements 只给漫画页补一列可展示的失败原因。上游原文
// 可能带较长的网关片段，用 TEXT 避免被 VARCHAR 截断；页表其它字段不变。
func V12ComicPageLastErrorStatements() []string {
	return []string{`ALTER TABLE batch_factory_v12_comic_pages ADD COLUMN last_error TEXT NULL AFTER status`}
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

// Recover a partially completed physical cutover without replaying already
// renamed tables. MySQL RENAME is DDL, so sql.Tx cannot roll back the previous
// per-table statements. Validate the complete plan before making one atomic
// rename of the remaining tables; never overwrite an existing target.
func adoptV12CanonicalTableRename(ctx context.Context, tx *sql.Tx) (bool, error) {
	pending := []string{}
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
		if sourceCount > 0 && targetCount > 0 {
			return false, fmt.Errorf("canonical table conflict: both %s and %s exist; neither table was changed", parts[2], parts[4])
		}
		if sourceCount == 0 && targetCount == 0 {
			return false, fmt.Errorf("canonical table missing after partial cutover: %s", parts[4])
		}
		if sourceCount > 0 {
			pending = append(pending, parts[2]+" TO "+parts[4])
		}
	}
	if len(pending) > 0 {
		if _, err := tx.ExecContext(ctx, "RENAME TABLE "+strings.Join(pending, ", ")); err != nil {
			return false, err
		}
	}
	return true, nil
}
