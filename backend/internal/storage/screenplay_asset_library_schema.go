package storage

// The durable schema is CM's storage adapter; the original native persistence
// format and duplicate-name policy have not been recovered. Independent image
// object copies are CM storage behavior, not recovered native copy semantics.
func screenplayAssetLibraryMigration() Migration {
	return Migration{Version: 8800116, CallbackChecksum: "v88-screenplay-ordinary-asset-library-v1", SQL: []string{`CREATE TABLE IF NOT EXISTS screenplay_asset_library (
 id VARCHAR(64) NOT NULL,
 owner_username VARCHAR(191) NOT NULL,
 asset_type VARCHAR(16) NOT NULL,
 name VARCHAR(191) NOT NULL,
 description LONGTEXT NOT NULL,
 category VARCHAR(191) NOT NULL DEFAULT '',
 image_json JSON NOT NULL,
 snapshot_json JSON NOT NULL,
 source_project_id VARCHAR(64) NOT NULL DEFAULT '',
 source_asset_id VARCHAR(64) NOT NULL DEFAULT '',
 source_media_id VARCHAR(64) NOT NULL DEFAULT '',
 image_object_json JSON NOT NULL,
 revision BIGINT NOT NULL,
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 PRIMARY KEY(id),
 KEY idx_screenplay_library_owner_type(owner_username,asset_type,updated_at),
 KEY idx_screenplay_library_category(owner_username,asset_type,category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`}}
}
