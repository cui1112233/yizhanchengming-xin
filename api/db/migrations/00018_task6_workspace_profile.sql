-- +goose Up
ALTER TABLE user_workspace_preferences
    ADD COLUMN pet_id VARCHAR(64) NOT NULL DEFAULT 'default' AFTER storage_preference,
    ADD COLUMN sound_volume TINYINT UNSIGNED NOT NULL DEFAULT 60 AFTER pet_id,
    ADD COLUMN pet_visible TINYINT(1) NOT NULL DEFAULT 1 AFTER sound_volume,
    ADD COLUMN companion_active TINYINT(1) NOT NULL DEFAULT 0 AFTER pet_visible;
ALTER TABLE video_local_executors
    ADD COLUMN owner_user_id BIGINT UNSIGNED NULL AFTER id,
    ADD KEY idx_video_local_executors_owner_seen (owner_user_id, last_seen_at),
    ADD CONSTRAINT fk_video_local_executors_owner FOREIGN KEY (owner_user_id) REFERENCES auth_users(id) ON DELETE RESTRICT;

CREATE TABLE video_local_executor_pairing_intents (
    token_hash BINARY(32) NOT NULL,
    owner_user_id BIGINT UNSIGNED NOT NULL,
    expires_at DATETIME(6) NOT NULL,
    used_at DATETIME(6) NULL,
    revoked_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL,
    PRIMARY KEY (token_hash),
    KEY idx_video_local_executor_pairings_owner (owner_user_id, created_at),
    KEY idx_video_local_executor_pairings_expiry (expires_at),
    CONSTRAINT fk_video_local_executor_pairings_owner FOREIGN KEY (owner_user_id) REFERENCES auth_users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose Down
DROP TABLE video_local_executor_pairing_intents;
ALTER TABLE user_workspace_preferences
    DROP COLUMN companion_active,
    DROP COLUMN pet_visible,
    DROP COLUMN sound_volume,
    DROP COLUMN pet_id;
ALTER TABLE video_local_executors
    DROP FOREIGN KEY fk_video_local_executors_owner,
    DROP KEY idx_video_local_executors_owner_seen,
    DROP COLUMN owner_user_id;
