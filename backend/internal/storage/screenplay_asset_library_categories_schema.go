package storage

func screenplayAssetLibraryCategoriesMigration() Migration {
	return Migration{Version: 8800118, CallbackChecksum: "v88-screenplay-owner-library-categories-v1", SQL: []string{`CREATE TABLE IF NOT EXISTS screenplay_asset_library_categories (owner_username VARCHAR(191) NOT NULL,state_json JSON NOT NULL,PRIMARY KEY(owner_username)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`}}
}
