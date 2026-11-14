package storage

func screenplayAssetCheckMigration() Migration {
	return Migration{Version: 8800115, CallbackChecksum: "v88-screenplay-asset-check-drafts-v1", SQL: []string{
		`CREATE TABLE IF NOT EXISTS screenplay_asset_check_jobs (
 id VARCHAR(64) NOT NULL,
 owner_username VARCHAR(191) NOT NULL,
 project_id VARCHAR(64) NOT NULL,
 request_key VARCHAR(191) COLLATE utf8mb4_bin NOT NULL,
 input_digest VARCHAR(64) NOT NULL,
 status VARCHAR(32) NOT NULL,
 applied BOOLEAN NOT NULL DEFAULT FALSE,
 candidate_count INT NOT NULL DEFAULT 0,
 state_json JSON NOT NULL,
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 PRIMARY KEY(id),
 UNIQUE KEY uk_asset_check_request(owner_username,project_id,request_key),
 KEY idx_asset_check_project(owner_username,project_id,updated_at),
 KEY idx_asset_check_runnable(status,updated_at),
 KEY idx_asset_check_unapplied(owner_username,project_id,status,applied,candidate_count)
 ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		// Exact names are distinct in the source contract. Binary collation makes the
		// existing unique name key accept A and a without changing stored text.
		`ALTER TABLE screenplay_assets MODIFY COLUMN name VARCHAR(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL`,
	}}
}
