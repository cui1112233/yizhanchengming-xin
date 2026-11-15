package storage

func LocalExecutorStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS local_executor_pairings (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  platform VARCHAR(32) NOT NULL,
  code_hash BINARY(32) NOT NULL,
  expires_at DATETIME(6) NOT NULL,
  consumed_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_local_executor_pairings_code_hash (code_hash),
  KEY idx_local_executor_pairings_owner_created (owner_username, created_at)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS local_executors (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  platform VARCHAR(32) NOT NULL,
  token_hash BINARY(32) NOT NULL,
  device_name VARCHAR(191) NOT NULL DEFAULT '',
  os VARCHAR(32) NOT NULL DEFAULT '',
  app_version VARCHAR(64) NOT NULL DEFAULT '',
  accounts_total INT NOT NULL DEFAULT 0,
  accounts_available INT NOT NULL DEFAULT 0,
  accounts_busy INT NOT NULL DEFAULT 0,
  accounts_quota_exhausted INT NOT NULL DEFAULT 0,
  accounts_login_error INT NOT NULL DEFAULT 0,
  accounts_human_verification INT NOT NULL DEFAULT 0,
  last_seen_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_local_executors_token_hash (token_hash),
  KEY idx_local_executors_owner_updated (owner_username, updated_at)
) ENGINE=InnoDB`,
	}
}

func LocalExecutorJobStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS local_executor_jobs (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  source_task_id VARCHAR(96) NOT NULL,
  platform VARCHAR(32) NOT NULL,
  payload_json JSON NOT NULL,
  state VARCHAR(32) NOT NULL,
  cancel_requested BOOLEAN NOT NULL DEFAULT FALSE,
  lease_executor_id VARCHAR(64) NULL,
  lease_token_hash BINARY(32) NULL,
  lease_generation BIGINT NOT NULL DEFAULT 0,
  lease_expires_at DATETIME(6) NULL,
  accepted_at DATETIME(6) NULL,
  accepted_account_id VARCHAR(191) NULL,
  submission_id VARCHAR(191) NULL,
  artifact_id VARCHAR(64) NULL,
  error_code VARCHAR(96) NULL,
  error_message VARCHAR(512) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_local_executor_job_source (owner_username, source_task_id),
  KEY idx_local_executor_jobs_claim (owner_username, platform, state, cancel_requested, lease_expires_at, created_at),
  KEY idx_local_executor_jobs_executor (lease_executor_id, updated_at)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS local_executor_job_events (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  job_id VARCHAR(64) NOT NULL,
  executor_id VARCHAR(64) NULL,
  event_type VARCHAR(64) NOT NULL,
  state VARCHAR(32) NOT NULL,
  detail VARCHAR(512) NULL,
  created_at DATETIME(6) NOT NULL,
  KEY idx_local_executor_job_events_job (job_id, id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS local_executor_artifacts (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  job_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  media_type VARCHAR(96) NOT NULL DEFAULT 'video/mp4',
  byte_size BIGINT NULL,
  sha256 CHAR(64) NULL,
  storage_ref VARCHAR(512) NULL,
  created_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_local_executor_artifact_job (job_id),
  KEY idx_local_executor_artifacts_owner_created (owner_username, created_at)
) ENGINE=InnoDB`,
	}
}

func LocalExecutorJobSourceTaskIDStatements() []string {
	return []string{
		`ALTER TABLE local_executor_jobs MODIFY COLUMN source_task_id VARCHAR(255) NOT NULL`,
	}
}

func LocalExecutorMultipleArtifactStatements() []string {
	return []string{
		`ALTER TABLE local_executor_artifacts DROP INDEX uq_local_executor_artifact_job, ADD KEY idx_local_executor_artifacts_job (job_id, created_at)`,
	}
}

func LocalExecutorMigrations() []Migration {
	return []Migration{
		{Version: 7801001, SQL: LocalExecutorStatements(), CallbackChecksum: "v78-local-executor-control-plane-v1"},
		{Version: 7801002, SQL: LocalExecutorJobStatements(), CallbackChecksum: "v78-local-executor-job-leasing-v1"},
		// The live database was created by the previous released binary, which
		// recorded the historical checksum before the callback label was
		// normalized. Accept that exact ledger value so a Go cutover does not
		// rerun the ALTER or require editing application data.
		{Version: 7801003, SQL: LocalExecutorJobSourceTaskIDStatements(), CallbackChecksum: "v78-local-executor-source-task-id-v2", LegacyChecksums: []string{"e2b97af4dcf7d240f2d9f41b457724507d55071d02c1287a7db52e544a8088ad"}},
		{Version: 7801004, SQL: LocalExecutorMultipleArtifactStatements(), CallbackChecksum: "v88-local-executor-multiple-artifacts-v1"},
	}
}

func AppMigrations() []Migration {
	out := append([]Migration(nil), V11Migrations()...)
	out = append(out, V12Migrations()...)
	out = append(out, LocalExecutorMigrations()...)
	out = append(out, GiantMaterialExecutorMigrations()...)
	out = append(out, AuthMigrations()...)
	out = append(out, AgentWorkspaceMigrations()...)
	out = append(out, AICutMigrations()...)
	out = append(out, AutomationMigrations()...)
	out = append(out, ScreenplayMigrations()...)
	out = append(out, ScreenplayProductionMigrations()...)
	out = append(out, ScreenplayMediaMigrations()...)
	out = append(out, ScreenplayAuthoringMigrations()...)
	out = append(out, NovelCreationMigrations()...)
	out = append(out, ProjectRetentionMigrations()...)
	out = append(out, ImageCapacityMigrations()...)
	return out
}
