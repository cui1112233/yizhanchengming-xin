package storage

func AICutPersonalTemplateMigrations() []Migration {
	return []Migration{{Version: 8800034, SQL: []string{`CREATE TABLE IF NOT EXISTS aicut_personal_templates (
 id VARCHAR(64) NOT NULL,
 owner_username VARCHAR(191) NOT NULL,
 title VARCHAR(255) NOT NULL,
 layout_json JSON NOT NULL,
 created_at DATETIME(6) NOT NULL,
 PRIMARY KEY (id,owner_username),
 KEY idx_aicut_personal_templates_owner (owner_username,created_at)
 ) ENGINE=InnoDB`}}}
}
