-- +goose Up
CREATE TABLE batch_project_settings (
    batch_project_id BIGINT NOT NULL,
    settings_json JSON NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (batch_project_id),
    CONSTRAINT fk_batch_project_settings_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE batch_version_config_profiles (
    id BIGINT NOT NULL AUTO_INCREMENT,
    batch_project_id BIGINT NOT NULL,
    profile_name VARCHAR(255) NOT NULL,
    version VARCHAR(64) NOT NULL,
    settings_json JSON NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_batch_version_profile_project (batch_project_id),
    CONSTRAINT fk_batch_version_profile_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS batch_version_config_profiles;
DROP TABLE IF EXISTS batch_project_settings;
