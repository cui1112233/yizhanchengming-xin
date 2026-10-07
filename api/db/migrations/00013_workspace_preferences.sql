-- +goose Up
CREATE TABLE user_workspace_preferences (
    user_id BIGINT UNSIGNED NOT NULL,
    theme VARCHAR(16) NOT NULL DEFAULT 'dark',
    notifications_enabled TINYINT(1) NOT NULL DEFAULT 1,
    storage_preference VARCHAR(32) NOT NULL DEFAULT 'tos',
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (user_id),
    CONSTRAINT fk_workspace_preferences_user FOREIGN KEY (user_id) REFERENCES auth_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose Down
DROP TABLE IF EXISTS user_workspace_preferences;
