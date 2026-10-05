-- +goose Up
CREATE TABLE video_provider_configs (
    id BIGINT NOT NULL AUTO_INCREMENT,
    provider_key VARCHAR(64) NOT NULL,
    model VARCHAR(128) NOT NULL,
    create_url VARCHAR(1024) NOT NULL DEFAULT '',
    tasks_url VARCHAR(1024) NOT NULL DEFAULT '',
    result_url VARCHAR(1024) NOT NULL DEFAULT '',
    encrypted_secret BLOB NOT NULL,
    secret_nonce VARBINARY(32) NOT NULL,
    enabled TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_video_provider_configs_provider_model (provider_key, model)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE video_production_jobs (
    id BIGINT NOT NULL AUTO_INCREMENT,
    batch_project_id BIGINT NOT NULL,
    book_id BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    input_revision VARCHAR(64) NOT NULL,
    final_prompt_stage_run_id BIGINT NOT NULL,
    final_prompt_version INT NOT NULL,
    final_prompt_text MEDIUMTEXT NOT NULL,
    provider VARCHAR(64) NOT NULL,
    model VARCHAR(128) NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL,
    error_code VARCHAR(64) NOT NULL DEFAULT '',
    error_message VARCHAR(1024) NOT NULL DEFAULT '',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_video_production_jobs_idempotency (idempotency_key),
    KEY idx_video_production_jobs_project_book (batch_project_id, book_id, id),
    KEY idx_video_production_jobs_status (status, id),
    CONSTRAINT fk_video_jobs_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_video_jobs_book FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
    CONSTRAINT fk_video_jobs_final_prompt FOREIGN KEY (final_prompt_stage_run_id) REFERENCES stage_runs(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE video_production_tasks (
    id BIGINT NOT NULL AUTO_INCREMENT,
    production_job_id BIGINT NOT NULL,
    attempt INT NOT NULL,
    provider VARCHAR(64) NOT NULL,
    model VARCHAR(128) NOT NULL,
    request_id VARCHAR(191) NOT NULL DEFAULT '',
    provider_job_id VARCHAR(191) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    error_code VARCHAR(64) NOT NULL DEFAULT '',
    error_message VARCHAR(1024) NOT NULL DEFAULT '',
    artifact_source_url VARCHAR(2048) NOT NULL DEFAULT '',
    output_bucket VARCHAR(255) NOT NULL DEFAULT '',
    output_object_key VARCHAR(1024) NOT NULL DEFAULT '',
    output_url VARCHAR(2048) NOT NULL DEFAULT '',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_video_production_tasks_attempt (production_job_id, attempt),
    KEY idx_video_production_tasks_recoverable (status, id),
    KEY idx_video_production_tasks_provider_job (provider, provider_job_id),
    CONSTRAINT fk_video_tasks_job FOREIGN KEY (production_job_id) REFERENCES video_production_jobs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE video_local_executors (
    id VARCHAR(64) NOT NULL,
    name VARCHAR(191) NOT NULL,
    provider_key VARCHAR(64) NOT NULL,
    model VARCHAR(128) NOT NULL,
    capabilities_json JSON NOT NULL,
    token_hash BINARY(32) NOT NULL,
    last_seen_at DATETIME(6) NOT NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_video_local_executors_token_hash (token_hash),
    KEY idx_video_local_executors_provider_model_seen (provider_key, model, last_seen_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE video_local_executor_tasks (
    id VARCHAR(64) NOT NULL,
    source_task_id VARCHAR(64) NOT NULL,
    provider_key VARCHAR(64) NOT NULL,
    model VARCHAR(128) NOT NULL,
    prompt MEDIUMTEXT NOT NULL,
    request_id VARCHAR(191) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    executor_id VARCHAR(64) NULL,
    artifact_url VARCHAR(2048) NOT NULL DEFAULT '',
    error_code VARCHAR(64) NOT NULL DEFAULT '',
    error_message VARCHAR(1024) NOT NULL DEFAULT '',
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    KEY idx_video_local_executor_tasks_source (source_task_id),
    KEY idx_video_local_executor_tasks_status (status, created_at),
    KEY idx_video_local_executor_tasks_executor (executor_id, status),
    CONSTRAINT fk_video_local_executor_tasks_executor FOREIGN KEY (executor_id) REFERENCES video_local_executors(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE video_merge_jobs (
    id BIGINT NOT NULL AUTO_INCREMENT,
    batch_project_id BIGINT NOT NULL,
    book_id BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    current_attempt INT NOT NULL DEFAULT 1,
    error_message VARCHAR(1024) NOT NULL DEFAULT '',
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    KEY idx_video_merge_jobs_project_book (batch_project_id, book_id, id),
    KEY idx_video_merge_jobs_status (status, id),
    CONSTRAINT fk_video_merge_jobs_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_video_merge_jobs_book FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE video_merge_attempts (
    id BIGINT NOT NULL AUTO_INCREMENT,
    merge_job_id BIGINT NOT NULL,
    attempt INT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    aspect_ratio VARCHAR(16) NOT NULL DEFAULT '9:16',
    speed DOUBLE NOT NULL DEFAULT 1,
    output_bucket VARCHAR(255) NOT NULL DEFAULT '',
    output_object_key VARCHAR(1024) NOT NULL DEFAULT '',
    output_url VARCHAR(2048) NOT NULL DEFAULT '',
    error_message VARCHAR(1024) NOT NULL DEFAULT '',
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_video_merge_attempts_job_attempt (merge_job_id, attempt),
    KEY idx_video_merge_attempts_status (status, id),
    CONSTRAINT fk_video_merge_attempts_job FOREIGN KEY (merge_job_id) REFERENCES video_merge_jobs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE video_merge_inputs (
    id BIGINT NOT NULL AUTO_INCREMENT,
    merge_attempt_id BIGINT NOT NULL,
    production_task_id BIGINT NOT NULL,
    input_order INT NOT NULL,
    source_url VARCHAR(2048) NOT NULL,
    created_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_video_merge_inputs_attempt_order (merge_attempt_id, input_order),
    KEY idx_video_merge_inputs_production_task (production_task_id),
    CONSTRAINT fk_video_merge_inputs_attempt FOREIGN KEY (merge_attempt_id) REFERENCES video_merge_attempts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS video_merge_inputs;
DROP TABLE IF EXISTS video_merge_attempts;
DROP TABLE IF EXISTS video_merge_jobs;
DROP TABLE IF EXISTS video_local_executor_tasks;
DROP TABLE IF EXISTS video_local_executors;
DROP TABLE IF EXISTS video_production_tasks;
DROP TABLE IF EXISTS video_production_jobs;
DROP TABLE IF EXISTS video_provider_configs;
