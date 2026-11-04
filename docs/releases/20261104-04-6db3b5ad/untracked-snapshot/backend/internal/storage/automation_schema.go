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

func AutomationMigrations() []Migration {
	return []Migration{{Version: 8800016, SQL: AutomationStatements(), CallbackChecksum: "v88-batch-automation-mysql-v1"}}
}
