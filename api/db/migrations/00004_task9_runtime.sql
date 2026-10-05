-- +goose Up
CREATE TABLE run_book_executions (
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
    UNIQUE KEY uq_run_book_executions_run_book (run_id, book_id),
    UNIQUE KEY uq_run_book_executions_idempotency (idempotency_key),
    KEY idx_run_book_executions_run_status (run_id, status),
    KEY idx_run_book_executions_lease (status, lease_until),
    CONSTRAINT fk_run_book_executions_run FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE,
    CONSTRAINT fk_run_book_executions_book FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS run_book_executions;
