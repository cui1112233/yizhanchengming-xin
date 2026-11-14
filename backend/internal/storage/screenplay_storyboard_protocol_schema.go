package storage

func screenplayStoryboardProtocolMigration() Migration {
	return Migration{Version: 8800114, CallbackChecksum: "v88-screenplay-storyboard-protocol-v1", SQL: []string{`ALTER TABLE screenplay_storyboard_shots ADD COLUMN protocol_json JSON NULL`}}
}
