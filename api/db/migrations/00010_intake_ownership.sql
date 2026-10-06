-- +goose Up
CREATE TABLE auth_intake_ownership (
    intake_id BIGINT NOT NULL,
    owner_user_id BIGINT UNSIGNED NOT NULL,
    team_id BIGINT UNSIGNED NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (intake_id),
    KEY idx_auth_intake_owner (owner_user_id),
    KEY idx_auth_intake_team (team_id),
    CONSTRAINT fk_auth_intake_intake FOREIGN KEY (intake_id) REFERENCES intakes(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_intake_owner FOREIGN KEY (owner_user_id) REFERENCES auth_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS auth_intake_ownership;
