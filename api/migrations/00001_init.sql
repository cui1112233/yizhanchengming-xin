-- +goose Up

CREATE TABLE IF NOT EXISTS intakes (
    id VARCHAR(36) PRIMARY KEY,
    owner VARCHAR(128) NOT NULL,
    title VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'created',
    run_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    KEY idx_intakes_owner_created (owner, created_at),
    KEY idx_intakes_status_run_at (status, run_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS intake_books (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    intake_id VARCHAR(36) NOT NULL,
    source_platform_id VARCHAR(32) NOT NULL,
    source_platform_name VARCHAR(128) NOT NULL DEFAULT '',
    book_id VARCHAR(128) NOT NULL,
    book_name VARCHAR(255) NOT NULL DEFAULT '',
    author VARCHAR(255) NOT NULL DEFAULT '',
    manual_gender VARCHAR(16) NOT NULL DEFAULT '',
    resolved_gender VARCHAR(16) NOT NULL DEFAULT '',
    gender_source VARCHAR(32) NOT NULL DEFAULT 'unresolved',
    style VARCHAR(128) NOT NULL DEFAULT '',
    style_source VARCHAR(32) NOT NULL DEFAULT '',
    category VARCHAR(255) NOT NULL DEFAULT '',
    genre INT NULL,
    source_text LONGTEXT NULL,
    source_metadata JSON NULL,
    fetch_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    error_message VARCHAR(1000) NOT NULL DEFAULT '',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    CONSTRAINT fk_intake_books_intake FOREIGN KEY (intake_id) REFERENCES intakes(id) ON DELETE CASCADE,
    KEY idx_intake_books_intake (intake_id),
    KEY idx_intake_books_source_book (source_platform_id, book_id),
    KEY idx_intake_books_gender (resolved_gender, gender_source)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS batches (
    id VARCHAR(36) PRIMARY KEY,
    owner VARCHAR(128) NOT NULL,
    intake_id VARCHAR(36) NULL,
    title VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'created',
    settings_json JSON NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    CONSTRAINT fk_batches_intake FOREIGN KEY (intake_id) REFERENCES intakes(id) ON DELETE SET NULL,
    KEY idx_batches_owner_created (owner, created_at),
    KEY idx_batches_intake (intake_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS batch_books (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    batch_id VARCHAR(36) NOT NULL,
    intake_book_id BIGINT UNSIGNED NULL,
    book_id VARCHAR(128) NOT NULL,
    title VARCHAR(255) NOT NULL DEFAULT '',
    source_platform_id VARCHAR(32) NOT NULL DEFAULT '',
    gender VARCHAR(16) NOT NULL DEFAULT '',
    gender_source VARCHAR(32) NOT NULL DEFAULT 'unresolved',
    style VARCHAR(128) NOT NULL DEFAULT '',
    source_text LONGTEXT NULL,
    state_json JSON NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    CONSTRAINT fk_batch_books_batch FOREIGN KEY (batch_id) REFERENCES batches(id) ON DELETE CASCADE,
    CONSTRAINT fk_batch_books_intake_book FOREIGN KEY (intake_book_id) REFERENCES intake_books(id) ON DELETE SET NULL,
    KEY idx_batch_books_batch (batch_id),
    KEY idx_batch_books_book (book_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS pipeline_jobs (
    id VARCHAR(36) PRIMARY KEY,
    owner VARCHAR(128) NOT NULL,
    intake_id VARCHAR(36) NULL,
    batch_id VARCHAR(36) NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    run_at DATETIME(6) NULL,
    current_stage VARCHAR(64) NOT NULL DEFAULT '',
    retry_count INT UNSIGNED NOT NULL DEFAULT 0,
    error_message VARCHAR(1000) NOT NULL DEFAULT '',
    idempotency_key VARCHAR(191) NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    CONSTRAINT fk_pipeline_jobs_intake FOREIGN KEY (intake_id) REFERENCES intakes(id) ON DELETE SET NULL,
    CONSTRAINT fk_pipeline_jobs_batch FOREIGN KEY (batch_id) REFERENCES batches(id) ON DELETE SET NULL,
    UNIQUE KEY uk_pipeline_jobs_owner_idempotency (owner, idempotency_key),
    KEY idx_pipeline_jobs_status_run_at (status, run_at),
    KEY idx_pipeline_jobs_owner_created (owner, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS pipeline_stages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    job_id VARCHAR(36) NOT NULL,
    stage VARCHAR(64) NOT NULL,
    ordinal_no INT UNSIGNED NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    attempt INT UNSIGNED NOT NULL DEFAULT 0,
    retry_count INT UNSIGNED NOT NULL DEFAULT 0,
    error_message VARCHAR(1000) NOT NULL DEFAULT '',
    input_json JSON NULL,
    output_json JSON NULL,
    started_at DATETIME(6) NULL,
    completed_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    CONSTRAINT fk_pipeline_stages_job FOREIGN KEY (job_id) REFERENCES pipeline_jobs(id) ON DELETE CASCADE,
    UNIQUE KEY uk_pipeline_stage_ordinal (job_id, ordinal_no),
    KEY idx_pipeline_stages_job_status (job_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down

DROP TABLE IF EXISTS pipeline_stages;
DROP TABLE IF EXISTS pipeline_jobs;
DROP TABLE IF EXISTS batch_books;
DROP TABLE IF EXISTS batches;
DROP TABLE IF EXISTS intake_books;
DROP TABLE IF EXISTS intakes;
