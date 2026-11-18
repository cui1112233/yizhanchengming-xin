package storage

func novelStudioStatements() []string {
	return []string{`CREATE TABLE IF NOT EXISTS novel_studio_sessions (
 owner_username VARCHAR(191) NOT NULL,
 studio_id VARCHAR(96) NOT NULL,
 create_key VARCHAR(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin NOT NULL,
 create_fingerprint CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 record_json JSON NOT NULL,
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 PRIMARY KEY (owner_username, studio_id),
 UNIQUE KEY uq_novel_studio_creation (owner_username, create_key),
 KEY idx_novel_studio_owner (owner_username, updated_at)
) ENGINE=InnoDB`}
}
