package storage

func screenplayAssetLibraryBulkMigration() Migration {
	return Migration{Version: 8800119, CallbackChecksum: "v88-screenplay-character-library-bulk-v1", SQL: []string{`CREATE TABLE IF NOT EXISTS screenplay_asset_library_bulk_operations (
 owner_username VARCHAR(191) NOT NULL,
 request_key VARCHAR(191) NOT NULL,
 digest CHAR(64) NOT NULL,
 result_json JSON NULL,
 PRIMARY KEY(owner_username,request_key)
 ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`, `CREATE TABLE IF NOT EXISTS screenplay_asset_library_bulk_owners (
 owner_username VARCHAR(191) NOT NULL PRIMARY KEY
 ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`}}
}
