package storage

func AICutAssetFolderMigrations() []Migration {
	return []Migration{{Version: 8800033, SQL: []string{
		`CREATE TABLE IF NOT EXISTS aicut_asset_folders (
 id VARCHAR(64) NOT NULL, project_id VARCHAR(64) NOT NULL, owner_username VARCHAR(191) NOT NULL,
 name VARCHAR(255) NOT NULL, created_at DATETIME(6) NOT NULL,
 PRIMARY KEY (id, project_id, owner_username),
 CONSTRAINT fk_aicut_folder_project FOREIGN KEY (project_id, owner_username) REFERENCES aicut_projects(id, owner_username) ON DELETE CASCADE
 ) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS aicut_asset_folder_members (
 asset_id VARCHAR(64) NOT NULL, project_id VARCHAR(64) NOT NULL, owner_username VARCHAR(191) NOT NULL, folder_id VARCHAR(64) NOT NULL,
 PRIMARY KEY (asset_id, project_id, owner_username),
 CONSTRAINT fk_aicut_folder_member_asset FOREIGN KEY (asset_id, project_id, owner_username) REFERENCES aicut_assets(id, project_id, owner_username) ON DELETE CASCADE,
 CONSTRAINT fk_aicut_folder_member_folder FOREIGN KEY (folder_id, project_id, owner_username) REFERENCES aicut_asset_folders(id, project_id, owner_username) ON DELETE CASCADE
 ) ENGINE=InnoDB`,
	}}}
}
