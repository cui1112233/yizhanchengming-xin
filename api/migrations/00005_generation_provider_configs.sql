-- +goose Up

CREATE TABLE IF NOT EXISTS generation_provider_configs (
    id VARCHAR(64) PRIMARY KEY,
    owner VARCHAR(128) NOT NULL,
    media_kind VARCHAR(16) NOT NULL,
    provider VARCHAR(64) NOT NULL,
    model VARCHAR(128) NOT NULL DEFAULT '',
    create_url VARCHAR(1024) NOT NULL DEFAULT '',
    tasks_url VARCHAR(1024) NOT NULL DEFAULT '',
    result_url VARCHAR(1024) NOT NULL DEFAULT '',
    credential_nonce VARBINARY(32) NULL,
    credential_ciphertext BLOB NULL,
    settings_json JSON NULL,
    enabled TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    UNIQUE KEY uk_generation_provider_owner_kind_provider (owner, media_kind, provider),
    KEY idx_generation_provider_owner_kind (owner, media_kind, enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down

DROP TABLE IF EXISTS generation_provider_configs;
