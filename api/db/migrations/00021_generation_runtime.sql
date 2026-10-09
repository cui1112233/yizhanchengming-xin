-- +goose Up
ALTER TABLE runs
    ADD COLUMN run_kind VARCHAR(32) NOT NULL DEFAULT 'legacy',
    ADD COLUMN target_book_id BIGINT NULL,
    ADD COLUMN request_schema_version INT NOT NULL DEFAULT 0,
    ADD COLUMN request_snapshot JSON NULL,
    ADD COLUMN request_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
    ADD COLUMN requested_by_user_id BIGINT NULL,
    ADD COLUMN cancel_requested_at DATETIME(6) NULL,
    ADD COLUMN cancelled_at DATETIME(6) NULL,
    ADD KEY idx_runs_generation_due (run_kind, status, run_at, id);

ALTER TABLE book_runs
    ADD COLUMN cancel_requested_at DATETIME(6) NULL,
    ADD COLUMN cancelled_at DATETIME(6) NULL;

-- +goose Down
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Generation runtime metadata cannot be rolled back without losing execution facts';
