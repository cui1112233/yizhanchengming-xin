package storage

func screenplayStoryboardPreferencesMigration() Migration {
	return Migration{Version: 8800111, CallbackChecksum: "v88-screenplay-owner-storyboard-preferences-v1", SQL: []string{`CREATE TABLE IF NOT EXISTS screenplay_storyboard_preferences (owner_username VARCHAR(191) NOT NULL,state_json JSON NOT NULL,PRIMARY KEY(owner_username)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`}}
}
