-- +goose Up

CREATE TABLE IF NOT EXISTS media_assets (
    id VARCHAR(64) PRIMARY KEY,
    owner VARCHAR(128) NOT NULL,
    media_type VARCHAR(16) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'ready',
    tos_bucket VARCHAR(128) NOT NULL,
    tos_key VARCHAR(1024) NOT NULL,
    mime_type VARCHAR(128) NOT NULL DEFAULT '',
    size_bytes BIGINT UNSIGNED NOT NULL DEFAULT 0,
    width_px INT UNSIGNED NOT NULL DEFAULT 0,
    height_px INT UNSIGNED NOT NULL DEFAULT 0,
    duration_ms BIGINT UNSIGNED NOT NULL DEFAULT 0,
    source_task_id VARCHAR(64) NULL,
    metadata_json JSON NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    UNIQUE KEY uk_media_owner_tos_object (owner, tos_bucket, tos_key),
    KEY idx_media_owner_created (owner, created_at),
    KEY idx_media_type_created (media_type, created_at),
    KEY idx_media_source_task (source_task_id),
    CONSTRAINT fk_media_source_task FOREIGN KEY (source_task_id) REFERENCES agent_tasks(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down

DROP TABLE IF EXISTS media_assets;
