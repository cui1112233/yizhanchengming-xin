-- +goose Up

CREATE TABLE IF NOT EXISTS agent_threads (
    id VARCHAR(64) PRIMARY KEY,
    owner VARCHAR(191) NOT NULL,
    title VARCHAR(255) NOT NULL DEFAULT '新对话',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    KEY idx_agent_threads_owner_updated (owner, updated_at, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS agent_messages (
    id VARCHAR(64) PRIMARY KEY,
    thread_id VARCHAR(64) NOT NULL,
    owner VARCHAR(191) NOT NULL,
    role VARCHAR(16) NOT NULL,
    content MEDIUMTEXT NOT NULL,
    metadata_json JSON NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    CONSTRAINT fk_agent_messages_thread FOREIGN KEY (thread_id) REFERENCES agent_threads(id) ON DELETE CASCADE,
    KEY idx_agent_messages_owner_thread_created (owner, thread_id, created_at, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS agent_tasks (
    id VARCHAR(64) PRIMARY KEY,
    thread_id VARCHAR(64) NOT NULL,
    owner VARCHAR(191) NOT NULL,
    title VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'in_progress',
    progress_current INT UNSIGNED NOT NULL DEFAULT 0,
    progress_total INT UNSIGNED NOT NULL DEFAULT 0,
    detail TEXT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    CONSTRAINT fk_agent_tasks_thread FOREIGN KEY (thread_id) REFERENCES agent_threads(id) ON DELETE CASCADE,
    KEY idx_agent_tasks_owner_status_updated (owner, status, updated_at),
    KEY idx_agent_tasks_thread_updated (thread_id, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS agent_tool_calls (
    id VARCHAR(64) PRIMARY KEY,
    thread_id VARCHAR(64) NOT NULL,
    message_id VARCHAR(64) NULL,
    owner VARCHAR(191) NOT NULL,
    tool_name VARCHAR(96) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'proposed',
    arguments_json JSON NOT NULL,
    result_json JSON NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    CONSTRAINT fk_agent_tool_calls_thread FOREIGN KEY (thread_id) REFERENCES agent_threads(id) ON DELETE CASCADE,
    CONSTRAINT fk_agent_tool_calls_message FOREIGN KEY (message_id) REFERENCES agent_messages(id) ON DELETE SET NULL,
    KEY idx_agent_tool_calls_owner_thread_created (owner, thread_id, created_at, id),
    KEY idx_agent_tool_calls_status_updated (status, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS agent_message_media (
    message_id VARCHAR(64) NOT NULL,
    media_asset_id VARCHAR(64) NOT NULL,
    ordinal INT UNSIGNED NOT NULL DEFAULT 0,
    PRIMARY KEY (message_id, media_asset_id),
    UNIQUE KEY uk_agent_message_media_ordinal (message_id, ordinal),
    CONSTRAINT fk_agent_message_media_message FOREIGN KEY (message_id) REFERENCES agent_messages(id) ON DELETE CASCADE,
    KEY idx_agent_message_media_asset (media_asset_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Agent V1 task states include: in_progress, waiting, needs_decision, completed, failed.

-- +goose Down

DROP TABLE IF EXISTS agent_message_media;
DROP TABLE IF EXISTS agent_tool_calls;
DROP TABLE IF EXISTS agent_tasks;
DROP TABLE IF EXISTS agent_messages;
DROP TABLE IF EXISTS agent_threads;
