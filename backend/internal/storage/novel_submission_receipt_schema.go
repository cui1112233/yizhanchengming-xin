package storage

func novelSubmissionReceiptStatements() []string {
	return []string{`CREATE TABLE IF NOT EXISTS novel_submission_receipts (
 owner_username VARCHAR(191) NOT NULL,
 kind VARCHAR(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin NOT NULL,
 idempotency_key VARCHAR(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin NOT NULL,
 fingerprint CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 project_id VARCHAR(96) NULL,
 result_json JSON NULL,
 created_at DATETIME(6) NOT NULL,
 PRIMARY KEY (owner_username,kind,idempotency_key),
 KEY idx_novel_submission_project (owner_username,project_id),
 CONSTRAINT fk_novel_submission_project FOREIGN KEY (owner_username,project_id) REFERENCES novel_projects(owner_username,project_id),
 CONSTRAINT chk_novel_submission_completed CHECK ((project_id IS NULL AND result_json IS NULL) OR (project_id IS NOT NULL AND result_json IS NOT NULL))
) ENGINE=InnoDB`}
}
