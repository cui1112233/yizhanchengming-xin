package storage

func screenplaySubjectUploadMigration() Migration {
	return Migration{Version: 8800120, CallbackChecksum: "v88-screenplay-subject-upload-v1", SQL: []string{`CREATE TABLE IF NOT EXISTS screenplay_subject_uploads (
 owner_username VARCHAR(191) NOT NULL,
 request_key VARCHAR(191) NOT NULL,
 revision BIGINT NOT NULL,
 operation_json JSON NOT NULL,
 PRIMARY KEY(owner_username,request_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`}}
}
