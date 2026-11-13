package storage

func ScreenplayMediaMigrations() []Migration {
	return []Migration{{Version: 8800106, CallbackChecksum: "screenplay-media-workbench-v1", SQL: []string{`CREATE TABLE IF NOT EXISTS screenplay_media_workbench(owner_username VARCHAR(191) NOT NULL, project_id VARCHAR(64) NOT NULL,state_json JSON NOT NULL,PRIMARY KEY(owner_username,project_id)) ENGINE=InnoDB`}}}
}
