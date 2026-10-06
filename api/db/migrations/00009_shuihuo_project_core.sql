-- +goose Up
CREATE TABLE shuihuo_projects (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_user_id BIGINT UNSIGNED NOT NULL,
    team_id BIGINT UNSIGNED NULL,
    name VARCHAR(255) NOT NULL,
    production_mode VARCHAR(64) NOT NULL DEFAULT 'commentary',
    source_text LONGTEXT NOT NULL,
    segmentation_status VARCHAR(32) NOT NULL DEFAULT 'draft',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_shuihuo_projects_owner (owner_user_id),
    KEY idx_shuihuo_projects_team (team_id),
    KEY idx_shuihuo_projects_updated (updated_at),
    CONSTRAINT fk_shuihuo_projects_owner FOREIGN KEY (owner_user_id) REFERENCES auth_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE shuihuo_segments (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    project_id BIGINT UNSIGNED NOT NULL,
    position INT UNSIGNED NOT NULL,
    source_text LONGTEXT NOT NULL,
    subtitle_text LONGTEXT NOT NULL,
    speaker VARCHAR(191) NOT NULL DEFAULT '旁白',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_shuihuo_segments_project_position (project_id, position),
    KEY idx_shuihuo_segments_project (project_id),
    CONSTRAINT fk_shuihuo_segments_project FOREIGN KEY (project_id) REFERENCES shuihuo_projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS shuihuo_segments;
DROP TABLE IF EXISTS shuihuo_projects;
