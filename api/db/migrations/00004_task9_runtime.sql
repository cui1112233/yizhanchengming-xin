-- +goose Up
CREATE TABLE book_runs (
    id BIGINT NOT NULL AUTO_INCREMENT,
    run_id BIGINT NOT NULL,
    book_id BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    attempt INT NOT NULL DEFAULT 0,
    error_message VARCHAR(1024) NOT NULL DEFAULT '',
    idempotency_key VARCHAR(191) NOT NULL,
    lease_until DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_book_runs_run_book (run_id, book_id),
    UNIQUE KEY uq_book_runs_idempotency (idempotency_key),
    KEY idx_book_runs_run_status (run_id, status),
    KEY idx_book_runs_lease (status, lease_until),
    CONSTRAINT fk_book_runs_run FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE,
    CONSTRAINT fk_book_runs_book FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS book_runs;
