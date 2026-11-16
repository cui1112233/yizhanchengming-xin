package storage

func novelContentStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS novel_content_records (
 owner_username VARCHAR(191) NOT NULL, project_id VARCHAR(96) NOT NULL,
 kind VARCHAR(32) NOT NULL, record_id VARCHAR(96) NOT NULL,
 status VARCHAR(32) NOT NULL, lease_until DATETIME(6) NULL,
 snapshot_json JSON NOT NULL, updated_at DATETIME(6) NOT NULL,
 PRIMARY KEY(owner_username,project_id,kind,record_id),
 KEY idx_novel_content_dispatch(kind,status,lease_until,updated_at),
 CONSTRAINT fk_novel_content_project FOREIGN KEY(owner_username,project_id) REFERENCES novel_projects(owner_username,project_id)
) ENGINE=InnoDB`,

		`ALTER TABLE novel_sessions ADD COLUMN archived_at DATETIME(6) NULL`,
	}
}
