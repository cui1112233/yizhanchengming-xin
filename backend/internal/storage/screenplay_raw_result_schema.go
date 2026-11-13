package storage

func screenplayRawResultMigration() Migration {
	return Migration{Version: 8800110, CallbackChecksum: "v88-screenplay-original-result-v1", SQL: []string{`ALTER TABLE screenplay_documents ADD COLUMN raw_result LONGTEXT NULL AFTER content`}}
}
