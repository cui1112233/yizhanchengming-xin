-- +goose Up
CREATE TABLE publishing_credentials (
    ref VARCHAR(191) NOT NULL,
    owner_user_id BIGINT UNSIGNED NOT NULL,
    team_id BIGINT UNSIGNED NULL,
    platform VARCHAR(64) NOT NULL,
    name VARCHAR(191) NOT NULL,
    key_id VARCHAR(64) NOT NULL,
    nonce VARBINARY(32) NOT NULL,
    ciphertext BLOB NOT NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    PRIMARY KEY (ref),
    KEY idx_publishing_credentials_owner (owner_user_id),
    KEY idx_publishing_credentials_team (team_id),
    CONSTRAINT fk_publishing_credentials_owner FOREIGN KEY (owner_user_id) REFERENCES auth_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE publishing_accounts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    owner_user_id BIGINT UNSIGNED NOT NULL,
    team_id BIGINT UNSIGNED NULL,
    platform VARCHAR(64) NOT NULL,
    display_name VARCHAR(191) NOT NULL,
    credential_ref VARCHAR(191) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    KEY idx_publishing_accounts_owner (owner_user_id),
    KEY idx_publishing_accounts_team (team_id),
    KEY idx_publishing_accounts_platform (platform),
    CONSTRAINT fk_publishing_accounts_owner FOREIGN KEY (owner_user_id) REFERENCES auth_users(id) ON DELETE CASCADE,
    CONSTRAINT fk_publishing_accounts_credential FOREIGN KEY (credential_ref) REFERENCES publishing_credentials(ref) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE publish_intents (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    batch_project_id BIGINT NOT NULL,
    book_id BIGINT NULL,
    publishing_account_id BIGINT UNSIGNED NOT NULL,
    requested_by_user_id BIGINT UNSIGNED NOT NULL,
    platform VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    requested_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    KEY idx_publish_intents_project (batch_project_id),
    KEY idx_publish_intents_account (publishing_account_id),
    KEY idx_publish_intents_requester (requested_by_user_id),
    KEY idx_publish_intents_status (status),
    CONSTRAINT fk_publish_intents_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_publish_intents_book FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE SET NULL,
    CONSTRAINT fk_publish_intents_account FOREIGN KEY (publishing_account_id) REFERENCES publishing_accounts(id) ON DELETE RESTRICT,
    CONSTRAINT fk_publish_intents_requester FOREIGN KEY (requested_by_user_id) REFERENCES auth_users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE publish_audits (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    intent_id BIGINT UNSIGNED NOT NULL,
    batch_project_id BIGINT NOT NULL,
    publishing_account_id BIGINT UNSIGNED NOT NULL,
    actor_user_id BIGINT UNSIGNED NOT NULL,
    platform VARCHAR(64) NOT NULL,
    action VARCHAR(64) NOT NULL,
    result VARCHAR(32) NOT NULL,
    error_summary VARCHAR(512) NOT NULL DEFAULT '',
    created_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    KEY idx_publish_audits_project (batch_project_id),
    KEY idx_publish_audits_account (publishing_account_id),
    KEY idx_publish_audits_actor (actor_user_id),
    KEY idx_publish_audits_intent (intent_id),
    CONSTRAINT fk_publish_audits_intent FOREIGN KEY (intent_id) REFERENCES publish_intents(id) ON DELETE CASCADE,
    CONSTRAINT fk_publish_audits_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_publish_audits_account FOREIGN KEY (publishing_account_id) REFERENCES publishing_accounts(id) ON DELETE RESTRICT,
    CONSTRAINT fk_publish_audits_actor FOREIGN KEY (actor_user_id) REFERENCES auth_users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS publish_audits;
DROP TABLE IF EXISTS publish_intents;
DROP TABLE IF EXISTS publishing_accounts;
DROP TABLE IF EXISTS publishing_credentials;
