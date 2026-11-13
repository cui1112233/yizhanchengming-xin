package storage

func screenplayCreativeMigration() Migration {
	return Migration{Version: 8800109, CallbackChecksum: "v88-screenplay-private-creative-v1", SQL: []string{
		`CREATE TABLE IF NOT EXISTS screenplay_user_styles (
 owner_username VARCHAR(191) NOT NULL, id VARCHAR(64) NOT NULL,
 name VARCHAR(255) NOT NULL, prompt MEDIUMTEXT NOT NULL, images JSON NOT NULL,
 revision BIGINT NOT NULL DEFAULT 0, created_at DATETIME(6) NOT NULL, updated_at DATETIME(6) NOT NULL,
 PRIMARY KEY(owner_username,id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS screenplay_project_creative_configs (
 project_id VARCHAR(64) NOT NULL, owner_username VARCHAR(191) NOT NULL,
 config JSON NOT NULL, updated_at DATETIME(6) NOT NULL,
 PRIMARY KEY(project_id), KEY creative_owner(owner_username)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}}
}
