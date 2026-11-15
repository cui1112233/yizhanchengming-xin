package storage

// Tombstones prevent a delayed upload retry from reusing a deleted image's ID
// and revision. This is CM's durable concurrency safeguard, not native parity.
func screenplayAssetLibraryTombstoneMigration() Migration {
	return Migration{Version: 8800117, CallbackChecksum: "v88-screenplay-library-tombstones-v1", SQL: []string{`ALTER TABLE screenplay_asset_library ADD COLUMN deleted_at DATETIME(6) NULL, ADD KEY idx_screenplay_library_active (owner_username,deleted_at,asset_type,updated_at)`}}
}
