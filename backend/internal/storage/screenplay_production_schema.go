package storage

func ScreenplayProductionMigrations() []Migration {
	return []Migration{{Version: 8800105, CallbackChecksum: "screenplay-production-v1", SQL: []string{
		`CREATE TABLE IF NOT EXISTS screenplay_production_jobs (owner_username VARCHAR(191) NOT NULL,project_id VARCHAR(64) NOT NULL,status VARCHAR(32) NOT NULL,error TEXT NOT NULL,units INT NOT NULL DEFAULT 0,rebuild BOOLEAN NOT NULL DEFAULT FALSE,updated_at DATETIME(6) NOT NULL,PRIMARY KEY(owner_username,project_id),KEY production_pending(status,updated_at)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS screenplay_production_shots (owner_username VARCHAR(191) NOT NULL,project_id VARCHAR(64) NOT NULL,storyboard_id VARCHAR(64) NOT NULL,shot_seq INT NOT NULL,request_hash VARCHAR(64) NOT NULL,payload JSON NOT NULL,updated_at DATETIME(6) NOT NULL,PRIMARY KEY(owner_username,project_id,storyboard_id,shot_seq,request_hash)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}}, {Version: 8800108, CallbackChecksum: "screenplay-production-schedule-v1", SQL: []string{`ALTER TABLE screenplay_production_jobs ADD COLUMN run_at DATETIME(6) NULL, ADD KEY production_scheduled(status,run_at)`}}}
}
