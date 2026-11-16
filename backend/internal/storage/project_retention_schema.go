package storage

// ProjectRetentionMigrations 保存漫剧创作、CM 画布、小说创作的作品集级保留状态。
// 批量工厂继续使用 batch_factory_v12_batches 上的保留字段，其余模块统一记录
// 在这张表中，避免每加一个模块就改一次业务表。
func ProjectRetentionMigrations() []Migration {
	return []Migration{{
		Version:          8800400,
		CallbackChecksum: "v88-project-retention-v1",
		SQL: []string{
			`CREATE TABLE IF NOT EXISTS project_retention (
  owner_username VARCHAR(191) NOT NULL,
  module VARCHAR(32) NOT NULL,
  project_id VARCHAR(96) NOT NULL,
  title VARCHAR(255) NOT NULL DEFAULT '',
  last_activity_at DATETIME(6) NOT NULL,
  retention_expires_at DATETIME(6) NOT NULL,
  archived_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (owner_username, module, project_id),
  KEY idx_project_retention_owner (owner_username, module, archived_at, retention_expires_at)
) ENGINE=InnoDB`,
		},
	}, {Version: 8800401, CallbackChecksum: "v88-retention-manual-novel-guard-v1", SQL: []string{`ALTER TABLE project_retention ADD COLUMN manual_archived TINYINT(1) NOT NULL DEFAULT 0`, `UPDATE novel_projects p JOIN project_retention r ON r.owner_username=p.owner_username AND r.project_id=p.project_id AND r.module='novel_creation' SET p.status='archived' WHERE r.archived_at IS NOT NULL`}}}
}
