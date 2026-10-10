package storage

import (
	"context"
	"database/sql"
)

func V11FoundationStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_bridge_users (
  username VARCHAR(191) NOT NULL PRIMARY KEY,
  is_owner BOOLEAN NOT NULL DEFAULT FALSE,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_batches (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  revision BIGINT NOT NULL DEFAULT 1,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  KEY idx_bfv11_batches_owner (owner_username)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_books (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  batch_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  ordinal INT NOT NULL,
  revision BIGINT NOT NULL DEFAULT 1,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_bfv11_book_ordinal (batch_id, ordinal),
  KEY idx_bfv11_books_owner_batch (owner_username, batch_id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_videos (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  ordinal INT NOT NULL,
  compatibility_state VARCHAR(32) NOT NULL DEFAULT 'active',
  revision BIGINT NOT NULL DEFAULT 1,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  KEY idx_bfv11_videos_owner_book (owner_username, book_id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_migration_audits (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  migration_version BIGINT NOT NULL,
  event_type VARCHAR(64) NOT NULL,
  expected_checksum CHAR(64) NULL,
  recorded_checksum CHAR(64) NULL,
  detail TEXT NULL,
  created_at DATETIME(6) NOT NULL
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_capability_settings (
  capability_key VARCHAR(96) NOT NULL PRIMARY KEY,
  available BOOLEAN NOT NULL DEFAULT FALSE,
  reason VARCHAR(255) NULL,
  updated_at DATETIME(6) NOT NULL
) ENGINE=InnoDB`,
	}
}

func V11FoundationMigrations() []Migration {
	return []Migration{{Version: 1100001, SQL: V11FoundationStatements(), CallbackChecksum: "batch-factory-v11-foundation-v1"}}
}

func V11SliceOneStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_intakes (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  payload_json JSON NOT NULL,
  consumed_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL,
  KEY idx_bfv11_intakes_owner_created (owner_username, created_at)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_batch_records (
  batch_id VARCHAR(64) NOT NULL PRIMARY KEY,
  title VARCHAR(255) NOT NULL DEFAULT '',
  source_intake_id VARCHAR(64) NULL
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_book_records (
  book_id VARCHAR(64) NOT NULL PRIMARY KEY,
  title VARCHAR(255) NOT NULL DEFAULT '',
  source_text MEDIUMTEXT NULL
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_video_records (
  video_id VARCHAR(64) NOT NULL PRIMARY KEY,
  label VARCHAR(255) NOT NULL DEFAULT '',
  visual_prompt MEDIUMTEXT NULL,
  duration_seconds DECIMAL(10,3) NULL
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_settings_patches (
  scope_type VARCHAR(16) NOT NULL,
  scope_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NULL,
  video_id VARCHAR(64) NULL,
  patch_json JSON NOT NULL,
  revision BIGINT NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (scope_type, scope_id),
  KEY idx_bfv11_settings_owner_batch (owner_username, batch_id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_config_versions (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  config_json JSON NOT NULL,
  created_at DATETIME(6) NOT NULL
) ENGINE=InnoDB`,
		`INSERT IGNORE INTO batch_factory_v11_config_versions(id,name,config_json,created_at) VALUES ('system-default-v1','System Default',JSON_OBJECT('source','system'),CURRENT_TIMESTAMP(6))`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_config_snapshots (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NULL,
  video_id VARCHAR(64) NULL,
  effective_json JSON NOT NULL,
  created_at DATETIME(6) NOT NULL,
  KEY idx_bfv11_snapshots_owner_batch (owner_username, batch_id, created_at)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_prompt_definitions (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  name VARCHAR(255) NOT NULL,
  kind VARCHAR(64) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  KEY idx_bfv11_prompts_owner_kind (owner_username, kind)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_prompt_versions (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  prompt_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  revision BIGINT NOT NULL,
  content MEDIUMTEXT NULL,
  created_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_bfv11_prompt_revision (prompt_id, revision),
  KEY idx_bfv11_prompt_versions_owner (owner_username, prompt_id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_drafts (
  owner_username VARCHAR(191) NOT NULL,
  draft_key VARCHAR(191) NOT NULL,
  kind VARCHAR(64) NOT NULL,
  scope VARCHAR(64) NOT NULL,
  content MEDIUMTEXT NULL,
  revision BIGINT NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (owner_username, draft_key, kind, scope)
) ENGINE=InnoDB`,
		`ALTER TABLE batch_factory_v11_books ADD CONSTRAINT fk_bfv11_books_batch FOREIGN KEY (batch_id) REFERENCES batch_factory_v11_batches(id) ON DELETE RESTRICT`,
		`ALTER TABLE batch_factory_v11_videos ADD CONSTRAINT fk_bfv11_videos_book FOREIGN KEY (book_id) REFERENCES batch_factory_v11_books(id) ON DELETE RESTRICT`,
		`ALTER TABLE batch_factory_v11_batch_records ADD CONSTRAINT fk_bfv11_batch_records_batch FOREIGN KEY (batch_id) REFERENCES batch_factory_v11_batches(id) ON DELETE RESTRICT`,
		`ALTER TABLE batch_factory_v11_book_records ADD CONSTRAINT fk_bfv11_book_records_book FOREIGN KEY (book_id) REFERENCES batch_factory_v11_books(id) ON DELETE RESTRICT`,
		`ALTER TABLE batch_factory_v11_video_records ADD CONSTRAINT fk_bfv11_video_records_video FOREIGN KEY (video_id) REFERENCES batch_factory_v11_videos(id) ON DELETE RESTRICT`,
		`ALTER TABLE batch_factory_v11_config_snapshots ADD CONSTRAINT fk_bfv11_snapshots_batch FOREIGN KEY (batch_id) REFERENCES batch_factory_v11_batches(id) ON DELETE RESTRICT`,
		`ALTER TABLE batch_factory_v11_prompt_versions ADD CONSTRAINT fk_bfv11_prompt_versions_prompt FOREIGN KEY (prompt_id) REFERENCES batch_factory_v11_prompt_definitions(id) ON DELETE RESTRICT`,
	}
}

func V11ConfigVersionOwnershipStatements() []string {
	return []string{
		`ALTER TABLE batch_factory_v11_config_versions
  ADD COLUMN owner_username VARCHAR(191) NULL AFTER id,
  ADD KEY idx_bfv11_config_versions_owner_created (owner_username, created_at, id)`,
	}
}

func V11DirectorStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_hook_revisions (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  revision BIGINT NOT NULL,
  status VARCHAR(32) NOT NULL,
  hook_text MEDIUMTEXT NOT NULL,
  source_digest CHAR(64) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  approved_at DATETIME(6) NULL,
  UNIQUE KEY uq_bfv11_hook_book_revision (book_id, revision),
  KEY idx_bfv11_hook_owner_book (owner_username, batch_id, book_id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_director_revisions (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  revision BIGINT NOT NULL,
  snapshot_id VARCHAR(64) NOT NULL,
  mode VARCHAR(32) NOT NULL,
  source_digest CHAR(64) NOT NULL,
  hook_revision_id VARCHAR(64) NULL,
  output_json JSON NOT NULL,
  created_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_bfv11_director_book_revision (book_id, revision),
  KEY idx_bfv11_director_owner_book (owner_username, batch_id, book_id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_director_video_links (
  director_revision_id VARCHAR(64) NOT NULL,
  video_id VARCHAR(64) NOT NULL,
  ordinal INT NOT NULL,
  PRIMARY KEY (director_revision_id, video_id),
  UNIQUE KEY uq_bfv11_director_video_ordinal (director_revision_id, ordinal)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_orphaned_overrides (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  director_revision_id VARCHAR(64) NOT NULL,
  video_id VARCHAR(64) NOT NULL,
  patch_json JSON NOT NULL,
  state VARCHAR(32) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  KEY idx_bfv11_orphan_director (director_revision_id)
) ENGINE=InnoDB`,
		`ALTER TABLE batch_factory_v11_hook_revisions ADD CONSTRAINT fk_bfv11_hook_book FOREIGN KEY (book_id) REFERENCES batch_factory_v11_books(id) ON DELETE RESTRICT`,
		`ALTER TABLE batch_factory_v11_director_revisions ADD CONSTRAINT fk_bfv11_director_book FOREIGN KEY (book_id) REFERENCES batch_factory_v11_books(id) ON DELETE RESTRICT`,
		`ALTER TABLE batch_factory_v11_director_revisions ADD CONSTRAINT fk_bfv11_director_snapshot FOREIGN KEY (snapshot_id) REFERENCES batch_factory_v11_config_snapshots(id) ON DELETE RESTRICT`,
		`ALTER TABLE batch_factory_v11_director_video_links ADD CONSTRAINT fk_bfv11_director_link_revision FOREIGN KEY (director_revision_id) REFERENCES batch_factory_v11_director_revisions(id) ON DELETE RESTRICT`,
		`ALTER TABLE batch_factory_v11_director_video_links ADD CONSTRAINT fk_bfv11_director_link_video FOREIGN KEY (video_id) REFERENCES batch_factory_v11_videos(id) ON DELETE RESTRICT`,
		`ALTER TABLE batch_factory_v11_orphaned_overrides ADD CONSTRAINT fk_bfv11_orphan_revision FOREIGN KEY (director_revision_id) REFERENCES batch_factory_v11_director_revisions(id) ON DELETE RESTRICT`,
	}
}

// V11ProductionStatements is deliberately append-only.  A production request
// must survive a process restart before any external video provider is called,
// so the job, every VIDEO task and the state-transition audit trail live in
// separate durable tables.
func V11ProductionStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_production_jobs (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NOT NULL,
  request_id VARCHAR(128) NOT NULL,
  director_revision_id VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_bfv11_production_request (owner_username, batch_id, book_id, request_id),
  KEY idx_bfv11_production_jobs_owner_batch_created (owner_username, batch_id, created_at),
  CONSTRAINT fk_bfv11_production_job_batch FOREIGN KEY (batch_id) REFERENCES batch_factory_v11_batches(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bfv11_production_job_book FOREIGN KEY (book_id) REFERENCES batch_factory_v11_books(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bfv11_production_job_director FOREIGN KEY (director_revision_id) REFERENCES batch_factory_v11_director_revisions(id) ON DELETE RESTRICT
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_production_tasks (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  job_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  video_id VARCHAR(64) NOT NULL,
  attempt INT NOT NULL,
  final_prompt_hash CHAR(64) NOT NULL,
  compiled_prompt MEDIUMTEXT NOT NULL,
  provider_task_id VARCHAR(255) NULL,
  media_url MEDIUMTEXT NULL,
  status VARCHAR(32) NOT NULL,
  error_message VARCHAR(255) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_bfv11_production_task_attempt (job_id, video_id, attempt),
  KEY idx_bfv11_production_tasks_owner_status (owner_username, status, updated_at),
  CONSTRAINT fk_bfv11_production_task_job FOREIGN KEY (job_id) REFERENCES batch_factory_v11_production_jobs(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bfv11_production_task_video FOREIGN KEY (video_id) REFERENCES batch_factory_v11_videos(id) ON DELETE RESTRICT
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_production_events (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  job_id VARCHAR(64) NOT NULL,
  task_id VARCHAR(64) NULL,
  owner_username VARCHAR(191) NOT NULL,
  event_type VARCHAR(64) NOT NULL,
  from_state VARCHAR(32) NULL,
  to_state VARCHAR(32) NULL,
  message VARCHAR(255) NULL,
  created_at DATETIME(6) NOT NULL,
  KEY idx_bfv11_production_events_job_created (job_id, created_at, id),
  CONSTRAINT fk_bfv11_production_event_job FOREIGN KEY (job_id) REFERENCES batch_factory_v11_production_jobs(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bfv11_production_event_task FOREIGN KEY (task_id) REFERENCES batch_factory_v11_production_tasks(id) ON DELETE RESTRICT
) ENGINE=InnoDB`,
	}
}

// V11MergeStatements persists the merge request and its exact ordered inputs.
// Keeping sources separate makes retries idempotent while retaining the media
// list that was actually handed to the merge provider.
// V11ProductionLibraryStatements records a user removing a generated candidate
// from the clip library without breaking task events, merge sources or provider
// audit history that reference the durable task row.
func V11ProductionLibraryStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_hidden_production_tasks (
  task_id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NOT NULL,
  video_id VARCHAR(64) NOT NULL,
  deleted_at DATETIME(6) NOT NULL,
  KEY idx_bfv11_hidden_production_tasks_book (owner_username,batch_id,book_id,video_id),
  CONSTRAINT fk_bfv11_hidden_production_task FOREIGN KEY (task_id) REFERENCES batch_factory_v11_production_tasks(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bfv11_hidden_production_batch FOREIGN KEY (batch_id) REFERENCES batch_factory_v11_batches(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bfv11_hidden_production_book FOREIGN KEY (book_id) REFERENCES batch_factory_v11_books(id) ON DELETE RESTRICT,
  CONSTRAINT fk_bfv11_hidden_production_video FOREIGN KEY (video_id) REFERENCES batch_factory_v11_videos(id) ON DELETE RESTRICT
) ENGINE=InnoDB`,
	}
}

func V11MergeStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_merge_jobs (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(64) NOT NULL,
  request_id VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL,
  output_url MEDIUMTEXT NULL,
  error_message VARCHAR(255) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_bfv11_merge_request (owner_username, batch_id, request_id),
  KEY idx_bfv11_merge_jobs_owner_batch_created (owner_username, batch_id, created_at),
  FOREIGN KEY (batch_id) REFERENCES batch_factory_v11_batches(id) ON DELETE RESTRICT
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_merge_sources (
  job_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  video_id VARCHAR(64) NOT NULL,
  ordinal INT NOT NULL,
  media_url MEDIUMTEXT NOT NULL,
  created_at DATETIME(6) NOT NULL,
  PRIMARY KEY (job_id, video_id),
  UNIQUE KEY uq_bfv11_merge_source_ordinal (job_id, ordinal),
  KEY idx_bfv11_merge_sources_owner_job (owner_username, job_id),
  FOREIGN KEY (job_id) REFERENCES batch_factory_v11_merge_jobs(id) ON DELETE RESTRICT
) ENGINE=InnoDB`,
	}
}

func V11MergePollerStatements() []string {
	return []string{
		`ALTER TABLE batch_factory_v11_merge_jobs ADD COLUMN provider_task_id VARCHAR(255) NULL AFTER request_id`,
	}
}

// V11ExternalStatements stores only encrypted provider credentials, immutable
// submission intents, and bounded redacted audit rows. No provider secret is
// represented in a JSON response or in cleartext columns.
func V11ExternalStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_external_credentials (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  provider VARCHAR(16) NOT NULL,
  credential_name VARCHAR(191) NOT NULL,
  key_id VARCHAR(32) NOT NULL,
  nonce VARBINARY(32) NOT NULL,
  ciphertext BLOB NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_bfv11_external_credential_owner_provider (owner_username, provider),
  KEY idx_bfv11_external_credential_owner (owner_username)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_external_intents (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  provider VARCHAR(16) NOT NULL,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NULL,
  payload_digest CHAR(64) NOT NULL,
  payload_json JSON NOT NULL,
  expires_at DATETIME(6) NOT NULL,
  confirmed_at DATETIME(6) NULL,
  submitted_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL,
  KEY idx_bfv11_external_intent_owner (owner_username, created_at),
  KEY idx_bfv11_external_intent_owner_provider (owner_username, provider, created_at),
  FOREIGN KEY (batch_id) REFERENCES batch_factory_v11_batches(id) ON DELETE RESTRICT
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_external_audits (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  provider VARCHAR(16) NOT NULL,
  intent_id VARCHAR(64) NOT NULL,
  action VARCHAR(32) NOT NULL,
  outcome VARCHAR(32) NOT NULL,
  reference_value VARCHAR(255) NULL,
  message VARCHAR(255) NULL,
  created_at DATETIME(6) NOT NULL,
  KEY idx_bfv11_external_audit_owner_created (owner_username, created_at),
  KEY idx_bfv11_external_audit_intent (intent_id),
  FOREIGN KEY (intent_id) REFERENCES batch_factory_v11_external_intents(id) ON DELETE RESTRICT
) ENGINE=InnoDB`,
	}
}

func V11VideoProviderStatements() []string {
	return []string{
		`ALTER TABLE batch_factory_v11_production_tasks ADD COLUMN provider VARCHAR(48) NOT NULL DEFAULT 'personal_api' AFTER video_id`,
	}
}

func V11SourceLineageStatements() []string {
	return []string{
		`ALTER TABLE batch_factory_v11_book_records
  ADD COLUMN source_book_id VARCHAR(255) NULL AFTER book_id,
  ADD COLUMN source_task_id VARCHAR(255) NULL AFTER source_book_id,
  ADD COLUMN platform VARCHAR(128) NULL AFTER source_task_id,
  ADD COLUMN txt_text MEDIUMTEXT NULL AFTER source_text,
  ADD COLUMN txt_file_name VARCHAR(255) NULL AFTER txt_text,
  ADD COLUMN source_metadata_json JSON NULL AFTER txt_file_name,
  ADD KEY idx_bfv11_book_records_source_book (source_book_id)`,
	}
}

func V11BookAssetsStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_book_assets (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NOT NULL,
  kind VARCHAR(16) NOT NULL,
  name VARCHAR(255) NOT NULL,
  prompt MEDIUMTEXT NOT NULL,
  source VARCHAR(32) NOT NULL,
  extraction_preset_id VARCHAR(191) NULL,
  extraction_preset_version BIGINT NULL,
  revision BIGINT NOT NULL DEFAULT 1,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  KEY idx_bfv11_assets_owner_book (owner_username, batch_id, book_id, kind),
  UNIQUE KEY uq_bfv11_assets_book_kind_name (owner_username, book_id, kind, name)
) ENGINE=InnoDB`,
		`ALTER TABLE batch_factory_v11_book_assets ADD CONSTRAINT fk_bfv11_assets_book FOREIGN KEY (book_id) REFERENCES batch_factory_v11_books(id) ON DELETE RESTRICT`,
	}
}

// reconcileV11BookAssetsLedger repairs only the documented public-release
// failure where the migration ledger was written even though the book-assets
// table was absent. A healthy installation merely performs the existence check.
func reconcileV11BookAssetsLedger(ctx context.Context, tx *sql.Tx) error {
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(
  SELECT 1 FROM information_schema.tables
  WHERE table_schema = DATABASE() AND table_name = 'batch_factory_v11_book_assets'
)`).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	for _, statement := range V11BookAssetsStatements() {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

// reconcileV11SeparateVideoPromptLedger repairs the matching historical
// ledger/schema split for the prompt separation migration. When the column is
// absent, run the original one-time copy before clearing visual_prompt; once
// present, existing data is left untouched.
func reconcileV11SeparateVideoPromptLedger(ctx context.Context, tx *sql.Tx) error {
	exists, err := v11ColumnExists(ctx, tx, "batch_factory_v11_video_records", "video_prompt")
	if err != nil || exists {
		return err
	}
	for _, statement := range V11SeparateVideoPromptStatements() {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func v11ColumnExists(ctx context.Context, tx *sql.Tx, table, column string) (bool, error) {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(
  SELECT 1 FROM information_schema.columns
  WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?
)`, table, column).Scan(&exists)
	return exists, err
}

func v11ColumnsExist(ctx context.Context, tx *sql.Tx, table string, columns ...string) (bool, error) {
	for _, column := range columns {
		exists, err := v11ColumnExists(ctx, tx, table, column)
		if err != nil || !exists {
			return exists, err
		}
	}
	return true, nil
}

func adoptV11BookMergeLedger(ctx context.Context, tx *sql.Tx) (bool, error) {
	return v11ColumnExists(ctx, tx, "batch_factory_v11_merge_jobs", "book_id")
}

func adoptV11MergeProgressLedger(ctx context.Context, tx *sql.Tx) (bool, error) {
	// A historical public release committed timing_mode/speed before its later
	// ALTER statements and migration ledger entry.  Do not adopt a fresh table;
	// only finish that recognizable prefix with the missing idempotent columns.
	timingModeExists, err := v11ColumnExists(ctx, tx, "batch_factory_v11_merge_jobs", "timing_mode")
	if err != nil || !timingModeExists {
		return false, err
	}
	for _, column := range []struct {
		name string
		sql  string
	}{
		{"speed", `ALTER TABLE batch_factory_v11_merge_jobs ADD COLUMN speed DECIMAL(8,3) NOT NULL DEFAULT 1 AFTER timing_mode`},
		{"progress_phase", `ALTER TABLE batch_factory_v11_merge_jobs ADD COLUMN progress_phase VARCHAR(32) NULL AFTER status`},
		{"progress_current", `ALTER TABLE batch_factory_v11_merge_jobs ADD COLUMN progress_current INT NOT NULL DEFAULT 0 AFTER progress_phase`},
		{"progress_total", `ALTER TABLE batch_factory_v11_merge_jobs ADD COLUMN progress_total INT NOT NULL DEFAULT 0 AFTER progress_current`},
	} {
		exists, err := v11ColumnExists(ctx, tx, "batch_factory_v11_merge_jobs", column.name)
		if err != nil {
			return false, err
		}
		if !exists {
			if _, err := tx.ExecContext(ctx, column.sql); err != nil {
				return false, err
			}
		}
	}
	return v11ColumnsExist(ctx, tx, "batch_factory_v11_merge_jobs", "timing_mode", "speed", "progress_phase", "progress_current", "progress_total")
}

func adoptV11OpeningVariantLedger(ctx context.Context, tx *sql.Tx) (bool, error) {
	for _, column := range []struct {
		table string
		name  string
		sql   string
	}{
		{"batch_factory_v11_production_tasks", "opening_variant_index", `ALTER TABLE batch_factory_v11_production_tasks ADD COLUMN opening_variant_index INT NOT NULL DEFAULT 0 AFTER video_id`},
		{"batch_factory_v11_merge_jobs", "opening_variant_index", `ALTER TABLE batch_factory_v11_merge_jobs ADD COLUMN opening_variant_index INT NOT NULL DEFAULT 0 AFTER book_id`},
	} {
		exists, err := v11ColumnExists(ctx, tx, column.table, column.name)
		if err != nil {
			return false, err
		}
		if !exists {
			if _, err := tx.ExecContext(ctx, column.sql); err != nil {
				return false, err
			}
		}
	}
	productionOK, err := v11ColumnsExist(ctx, tx, "batch_factory_v11_production_tasks", "opening_variant_index")
	if err != nil || !productionOK {
		return false, err
	}
	return v11ColumnsExist(ctx, tx, "batch_factory_v11_merge_jobs", "opening_variant_index")
}

func V11BookAssetImagesStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_book_asset_images (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  asset_id VARCHAR(64) NOT NULL,
  url MEDIUMTEXT NOT NULL,
  storage_ref VARCHAR(255) NULL,
  media_type VARCHAR(96) NOT NULL,
  source VARCHAR(32) NOT NULL,
  is_primary BOOLEAN NOT NULL DEFAULT FALSE,
  revision BIGINT NOT NULL DEFAULT 1,
  created_at DATETIME(6) NOT NULL,
  KEY idx_bfv11_asset_images_owner_asset (owner_username, asset_id, created_at),
  KEY idx_bfv11_asset_images_primary (asset_id, is_primary),
  CONSTRAINT fk_bfv11_asset_images_asset FOREIGN KEY (asset_id) REFERENCES batch_factory_v11_book_assets(id) ON DELETE RESTRICT
) ENGINE=InnoDB`,
	}
}

func V11BookStageRunsStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_book_stage_runs (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NOT NULL,
  stage VARCHAR(16) NOT NULL,
  status VARCHAR(32) NOT NULL,
  attempt INT NOT NULL,
  request_id VARCHAR(128) NULL,
  input_revision VARCHAR(128) NULL,
  error_message VARCHAR(255) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  KEY idx_bfv11_stage_owner_book_updated (owner_username, batch_id, book_id, updated_at),
  FOREIGN KEY (batch_id) REFERENCES batch_factory_v11_batches(id) ON DELETE RESTRICT,
  FOREIGN KEY (book_id) REFERENCES batch_factory_v11_books(id) ON DELETE RESTRICT
) ENGINE=InnoDB`,
	}
}

func V11ProductionAssetInputStatements() []string {
	return []string{
		`ALTER TABLE batch_factory_v11_production_tasks ADD COLUMN reference_image_urls MEDIUMTEXT NULL AFTER compiled_prompt`,
		`ALTER TABLE batch_factory_v11_production_tasks ADD COLUMN downgraded_asset_ids MEDIUMTEXT NULL AFTER reference_image_urls`,
	}
}

func V11ProductionDurationStatements() []string {
	return []string{
		`ALTER TABLE batch_factory_v11_production_tasks ADD COLUMN target_duration_seconds DECIMAL(10,3) NOT NULL DEFAULT 0 AFTER downgraded_asset_ids`,
		`ALTER TABLE batch_factory_v11_production_tasks ADD COLUMN requested_duration_seconds DECIMAL(10,3) NOT NULL DEFAULT 0 AFTER target_duration_seconds`,
		`ALTER TABLE batch_factory_v11_production_tasks ADD COLUMN actual_duration_seconds DECIMAL(10,3) NOT NULL DEFAULT 0 AFTER requested_duration_seconds`,
	}
}

func V11MergeSourceDurationStatements() []string {
	return []string{
		`ALTER TABLE batch_factory_v11_merge_sources ADD COLUMN target_duration_seconds DECIMAL(10,3) NOT NULL DEFAULT 0 AFTER media_url`,
		`ALTER TABLE batch_factory_v11_merge_sources ADD COLUMN requested_duration_seconds DECIMAL(10,3) NOT NULL DEFAULT 0 AFTER target_duration_seconds`,
		`ALTER TABLE batch_factory_v11_merge_sources ADD COLUMN actual_duration_seconds DECIMAL(10,3) NOT NULL DEFAULT 0 AFTER requested_duration_seconds`,
	}
}

// V11MergeProgressStatements keeps the selected timing option and durable
// executor progress for each book-level final merge. Every statement runs
// exactly once through the versioned migration marker; MySQL 8.4 does not
// accept ADD COLUMN IF NOT EXISTS here.
func V11MergeProgressStatements() []string {
	return []string{
		`ALTER TABLE batch_factory_v11_merge_jobs ADD COLUMN timing_mode VARCHAR(32) NULL AFTER request_id`,
		`ALTER TABLE batch_factory_v11_merge_jobs ADD COLUMN speed DECIMAL(8,3) NOT NULL DEFAULT 1 AFTER timing_mode`,
		`ALTER TABLE batch_factory_v11_merge_jobs ADD COLUMN progress_phase VARCHAR(32) NULL AFTER status`,
		`ALTER TABLE batch_factory_v11_merge_jobs ADD COLUMN progress_current INT NOT NULL DEFAULT 0 AFTER progress_phase`,
		`ALTER TABLE batch_factory_v11_merge_jobs ADD COLUMN progress_total INT NOT NULL DEFAULT 0 AFTER progress_current`,
	}
}

// V11OpeningVariantStatements records which alternative opening variant (0 =
// original director prompt) each production task and book-level merge job used.
// The adopt callback adds the columns idempotently so installations that
// already carry them are adopted without rerunning non-idempotent ALTER TABLE.
func V11OpeningVariantStatements() []string {
	return []string{
		`ALTER TABLE batch_factory_v11_production_tasks ADD COLUMN opening_variant_index INT NOT NULL DEFAULT 0 AFTER video_id`,
		`ALTER TABLE batch_factory_v11_merge_jobs ADD COLUMN opening_variant_index INT NOT NULL DEFAULT 0 AFTER book_id`,
	}
}

// V11LocalArtifactPurgeStatements keeps a deletion receipt after the parent
// Batch Factory rows have gone. It is intentionally not foreign-keyed: the
// whole point is to retain the exact local file reference until its delayed
// physical cleanup has completed.
func V11LocalArtifactPurgeStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_factory_v11_local_artifact_purges (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(64) NOT NULL,
  book_id VARCHAR(64) NOT NULL DEFAULT '',
  storage_ref VARCHAR(128) NOT NULL,
  purge_after DATETIME(6) NOT NULL,
  state VARCHAR(16) NOT NULL,
  attempt_count INT NOT NULL DEFAULT 0,
  last_error VARCHAR(512) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  deleted_at DATETIME(6) NULL,
  UNIQUE KEY uq_bfv11_local_artifact_purge (owner_username,batch_id,storage_ref),
  KEY idx_bfv11_local_artifact_purge_due (state,purge_after,id)
) ENGINE=InnoDB`,
	}
}

func V11ShuihuoConfigStatements() []string {
	return []string{`CREATE TABLE IF NOT EXISTS batch_factory_v11_shuihuo_configs (
  owner_username VARCHAR(191) NOT NULL PRIMARY KEY,
  config_json JSON NOT NULL,
  revision BIGINT NOT NULL DEFAULT 1,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL
) ENGINE=InnoDB`}
}

func V11Migrations() []Migration {
	return []Migration{
		{Version: 1100001, SQL: V11FoundationStatements(), CallbackChecksum: "batch-factory-v11-foundation-v1"},
		{Version: 1100002, SQL: V11SliceOneStatements(), CallbackChecksum: "batch-factory-v11-slice1-v1"},
		{Version: 1100003, SQL: V11ConfigVersionOwnershipStatements(), CallbackChecksum: "batch-factory-v11-slice1-config-version-ownership-v1"},
		{Version: 1100004, SQL: V11DirectorStatements(), CallbackChecksum: "batch-factory-v11-director-v1"},
		{Version: 1100005, SQL: V11ProductionStatements(), CallbackChecksum: "batch-factory-v11-production-v1"},
		{Version: 1100006, SQL: V11MergeStatements(), CallbackChecksum: "batch-factory-v11-merge-v1"},
		{Version: 1100007, SQL: V11MergePollerStatements(), CallbackChecksum: "batch-factory-v11-merge-poller-v1"},
		{Version: 1100008, SQL: V11ExternalStatements(), CallbackChecksum: "batch-factory-v11-external-publish-v1"},
		{Version: 1100009, SQL: V11VideoProviderStatements(), CallbackChecksum: "batch-factory-v11-video-provider-v1"},
		{Version: 1100010, SQL: V11SourceLineageStatements(), CallbackChecksum: "batch-factory-v11-source-lineage-v1"},
		// The public V78 deployment applied this migration before the V88
		// canonical callback marker existed. Its recorded checksum is accepted
		// only as evidence that the one-time prompt split already ran; it never
		// causes the ALTER/UPDATE sequence to run twice.
		{Version: 1100011, SQL: V11SeparateVideoPromptStatements(), CallbackChecksum: "batch-factory-v11-separate-video-prompt-v1", LegacyChecksums: []string{"0aa0615fbe3d1ffa5d13faee0acc76ec39c0b1232c687630ade52e9f041ab329"}, Reconcile: reconcileV11SeparateVideoPromptLedger},
		// The prior public release recorded the SQL-derived checksum before this
		// migration received its canonical callback marker.  Accept only that
		// verified ledger value so an existing book-assets table is never applied
		// a second time.
		{Version: 1100012, SQL: V11BookAssetsStatements(), CallbackChecksum: "batch-factory-v11-book-assets-v1", LegacyChecksums: []string{"b5e1e45e1ae47240a5d303efe70ad8b29afa21d906b482774d696c8424600624"}, Reconcile: reconcileV11BookAssetsLedger},
		{Version: 1100013, SQL: V11BookAssetImagesStatements(), CallbackChecksum: "batch-factory-v11-book-asset-images-v1"},
		{Version: 1100014, SQL: V11ProductionAssetInputStatements(), CallbackChecksum: "batch-factory-v11-production-asset-input-v1"},
		{Version: 1100015, SQL: V11BookStageRunsStatements(), CallbackChecksum: "batch-factory-v11-book-stage-runs-v1"},
		{Version: 1100016, SQL: V11BookMergeStatements(), CallbackChecksum: "batch-factory-v11-book-merge-v1", Adopt: adoptV11BookMergeLedger},
		{Version: 1100017, SQL: V11ProductionDurationStatements(), CallbackChecksum: "batch-factory-v11-production-durations-v1"},
		{Version: 1100018, SQL: V11MergeSourceDurationStatements(), CallbackChecksum: "batch-factory-v11-merge-source-durations-v1"},
		{Version: 1100019, SQL: V11ProductionLibraryStatements(), CallbackChecksum: "batch-factory-v11-production-library-v1"},
		{Version: 1100020, SQL: V11MergeProgressStatements(), CallbackChecksum: "batch-factory-v11-merge-progress-v1", Adopt: adoptV11MergeProgressLedger},
		{Version: 1100021, SQL: V11OpeningVariantStatements(), CallbackChecksum: "batch-factory-v11-opening-variant-v1", Adopt: adoptV11OpeningVariantLedger},
		{Version: 1100022, SQL: V11LocalArtifactPurgeStatements(), CallbackChecksum: "batch-factory-v11-local-artifact-purge-v1"},
		{Version: 1100023, SQL: V11ShuihuoConfigStatements(), CallbackChecksum: "batch-factory-v11-shuihuo-config-v1"},
	}
}

// V11SeparateVideoPromptStatements separates the legacy overloaded
// visual_prompt column. Existing Director output was historically stored in
// visual_prompt even though it is a video instruction. Preserve it as
// video_prompt once, leaving visual_prompt free for the independent still
// image chain required by V11.
func V11SeparateVideoPromptStatements() []string {
	return []string{
		`ALTER TABLE batch_factory_v11_video_records ADD COLUMN video_prompt MEDIUMTEXT NULL AFTER label`,
		`UPDATE batch_factory_v11_video_records SET video_prompt=visual_prompt WHERE (video_prompt IS NULL OR video_prompt='') AND visual_prompt IS NOT NULL AND visual_prompt<>''`,
		`UPDATE batch_factory_v11_video_records SET visual_prompt=NULL WHERE video_prompt IS NOT NULL AND video_prompt<>''`,
	}
}

func V11BookMergeStatements() []string {
	return []string{`ALTER TABLE batch_factory_v11_merge_jobs ADD COLUMN book_id VARCHAR(64) NULL AFTER batch_id`}
}
