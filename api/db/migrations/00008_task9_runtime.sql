-- +goose Up
ALTER TABLE runs
    ADD COLUMN idempotency_key VARCHAR(191) NULL AFTER batch_project_id,
    ADD COLUMN max_attempts INT NOT NULL DEFAULT 3 AFTER status,
    ADD COLUMN error_code VARCHAR(64) NOT NULL DEFAULT '' AFTER max_attempts,
    ADD COLUMN error_message VARCHAR(1024) NOT NULL DEFAULT '' AFTER error_code,
    ADD COLUMN started_at DATETIME(6) NULL AFTER error_message,
    ADD COLUMN finished_at DATETIME(6) NULL AFTER started_at,
    ADD UNIQUE KEY uq_runs_project_idempotency (batch_project_id, idempotency_key),
    ADD KEY idx_runs_status_run_at (status, run_at);

ALTER TABLE book_runs
    ADD COLUMN run_id BIGINT NULL AFTER id,
    ADD COLUMN attempt INT NOT NULL DEFAULT 1 AFTER book_id,
    ADD COLUMN max_attempts INT NOT NULL DEFAULT 3 AFTER attempt,
    ADD COLUMN retryable TINYINT(1) NOT NULL DEFAULT 1 AFTER max_attempts,
    ADD COLUMN error_code VARCHAR(64) NOT NULL DEFAULT '' AFTER request_id,
    ADD COLUMN execution_token BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER error_message,
    ADD COLUMN execution_owner VARCHAR(191) NOT NULL DEFAULT '' AFTER execution_token,
    ADD COLUMN running_since DATETIME(6) NULL AFTER execution_owner,
    ADD COLUMN lease_deadline DATETIME(6) NULL AFTER running_since,
    ADD COLUMN heartbeat_at DATETIME(6) NULL AFTER lease_deadline,
    ADD CONSTRAINT fk_book_runs_run FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE,
    ADD UNIQUE KEY uq_book_runs_run_book_attempt (run_id, book_id, attempt),
    ADD KEY idx_book_runs_runtime_claim (status, lease_deadline, id),
    ADD KEY idx_book_runs_run_status (run_id, status, id);

-- +goose Down
ALTER TABLE book_runs
    DROP FOREIGN KEY fk_book_runs_run,
    DROP INDEX uq_book_runs_run_book_attempt,
    DROP INDEX idx_book_runs_runtime_claim,
    DROP INDEX idx_book_runs_run_status,
    DROP COLUMN heartbeat_at,
    DROP COLUMN lease_deadline,
    DROP COLUMN running_since,
    DROP COLUMN execution_owner,
    DROP COLUMN execution_token,
    DROP COLUMN error_code,
    DROP COLUMN retryable,
    DROP COLUMN max_attempts,
    DROP COLUMN attempt,
    DROP COLUMN run_id;

-- MySQL may choose uq_runs_project_idempotency as the supporting index for
-- the pre-existing fk_runs_batch_project constraint and discard the older
-- implicit FK index. Recreate the FK after removing the Task 9 index so the
-- pre-Task9 schema remains rollback-safe.
ALTER TABLE runs
    DROP FOREIGN KEY fk_runs_batch_project;

ALTER TABLE runs
    DROP INDEX uq_runs_project_idempotency,
    DROP INDEX idx_runs_status_run_at,
    DROP COLUMN finished_at,
    DROP COLUMN started_at,
    DROP COLUMN error_message,
    DROP COLUMN error_code,
    DROP COLUMN max_attempts,
    DROP COLUMN idempotency_key;

ALTER TABLE runs
    ADD CONSTRAINT fk_runs_batch_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE;
