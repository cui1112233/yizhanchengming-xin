package storage

func screenplayStoryboardGenerationMigration() Migration {
	return Migration{Version: 8800112, CallbackChecksum: "v88-screenplay-storyboard-generation-attempts-v1", SQL: []string{`CREATE TABLE IF NOT EXISTS screenplay_storyboard_generations(owner_username VARCHAR(191) NOT NULL,project_id VARCHAR(64) NOT NULL,state_json JSON NOT NULL,PRIMARY KEY(owner_username,project_id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`}}
}
