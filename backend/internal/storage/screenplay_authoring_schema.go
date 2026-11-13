package storage

func ScreenplayAuthoringMigrations() []Migration {
	return []Migration{{Version: 8800107, CallbackChecksum: "screenplay-authoring-v1", SQL: []string{`CREATE TABLE IF NOT EXISTS screenplay_authoring(project_id VARCHAR(64) NOT NULL,owner_username VARCHAR(191) NOT NULL,revision BIGINT NOT NULL DEFAULT 0,body JSON NOT NULL,PRIMARY KEY(project_id,owner_username)) ENGINE=InnoDB`}}}
}
