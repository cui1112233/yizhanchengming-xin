package storage

func AutomationStatements() []string {
	return []string{`CREATE TABLE IF NOT EXISTS batch_automation_jobs (
  id VARCHAR(191) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(96) NOT NULL,
  state VARCHAR(32) NOT NULL,
  scheduled_at DATETIME(6) NULL,
  next_retry_at DATETIME(6) NULL,
  job_json JSON NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  KEY idx_batch_automation_due (state, scheduled_at, next_retry_at),
  KEY idx_batch_automation_owner_batch (owner_username, batch_id, updated_at)
) ENGINE=InnoDB`}
}

func BatchExecutionStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS batch_executions (
  id VARCHAR(191) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(96) NOT NULL,
  mode VARCHAR(32) NOT NULL,
  target VARCHAR(32) NOT NULL,
  client_request_id VARCHAR(191) NOT NULL,
  status VARCHAR(32) NOT NULL,
  legacy_job_id VARCHAR(191) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_batch_execution_request (owner_username, batch_id, client_request_id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_execution_config_snapshots (
  execution_id VARCHAR(191) NOT NULL PRIMARY KEY,
  config_json JSON NOT NULL,
  config_digest CHAR(64) NOT NULL,
  created_at DATETIME(6) NOT NULL
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_execution_books (
  id VARCHAR(191) NOT NULL PRIMARY KEY,
  execution_id VARCHAR(191) NOT NULL,
  book_id VARCHAR(64) NOT NULL,
  title VARCHAR(255) NOT NULL,
  status VARCHAR(32) NOT NULL,
  current_stage VARCHAR(64) NULL,
  failure_code VARCHAR(96) NULL,
  failure_message TEXT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_batch_execution_book (execution_id, book_id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_execution_stage_runs (
  id VARCHAR(191) NOT NULL PRIMARY KEY,
  execution_book_id VARCHAR(191) NOT NULL,
  stage_key VARCHAR(64) NOT NULL,
  input_fingerprint CHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  next_retry_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_batch_execution_stage_fingerprint (execution_book_id, stage_key, input_fingerprint)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_execution_stage_attempts (
  id VARCHAR(191) NOT NULL PRIMARY KEY,
  stage_run_id VARCHAR(191) NOT NULL,
  attempt_no INT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL,
  provider_request_id VARCHAR(191) NULL,
  error_code VARCHAR(96) NULL,
  error_message TEXT NULL,
  started_at DATETIME(6) NOT NULL,
  finished_at DATETIME(6) NULL
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_execution_publication_receipts (
  id VARCHAR(191) NOT NULL PRIMARY KEY,
  execution_book_id VARCHAR(191) NOT NULL,
  platform VARCHAR(32) NOT NULL,
  remote_work_id VARCHAR(191) NOT NULL,
  remote_status VARCHAR(32) NOT NULL,
  payload_json JSON NOT NULL,
  confirmed_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS batch_execution_outbox (
  id VARCHAR(191) NOT NULL PRIMARY KEY,
  aggregate_type VARCHAR(64) NOT NULL,
  aggregate_id VARCHAR(191) NOT NULL,
  event_type VARCHAR(64) NOT NULL,
  payload_json JSON NOT NULL,
  available_at DATETIME(6) NOT NULL,
  claimed_at DATETIME(6) NULL,
  delivered_at DATETIME(6) NULL,
  attempts INT UNSIGNED NOT NULL DEFAULT 0,
  last_error TEXT NULL,
  created_at DATETIME(6) NOT NULL
) ENGINE=InnoDB`,
	}
}

func AutomationMigrations() []Migration {
	return []Migration{
		{Version: 8800016, SQL: AutomationStatements(), CallbackChecksum: "v88-batch-automation-mysql-v1"},
		{Version: 8800035, SQL: BatchExecutionStatements()},
	}
}
