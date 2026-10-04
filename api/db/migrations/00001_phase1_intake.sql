-- +goose Up
CREATE TABLE intakes (
    id BIGINT NOT NULL AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_intakes_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE books (
    id BIGINT NOT NULL AUTO_INCREMENT,
    intake_id BIGINT NOT NULL,
    source VARCHAR(64) NOT NULL,
    external_book_id VARCHAR(191) NOT NULL,
    title VARCHAR(255) NOT NULL DEFAULT '',
    body_ref VARCHAR(1024) NOT NULL DEFAULT '',
    category VARCHAR(128) NOT NULL DEFAULT '',
    genre VARCHAR(128) NOT NULL DEFAULT '',
    gender VARCHAR(32) NOT NULL DEFAULT '',
    gender_source VARCHAR(32) NOT NULL DEFAULT '',
    style VARCHAR(128) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    error_message VARCHAR(1024) NOT NULL DEFAULT '',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_books_intake_source_external (intake_id, source, external_book_id),
    KEY idx_books_intake_status (intake_id, status),
    CONSTRAINT fk_books_intake FOREIGN KEY (intake_id) REFERENCES intakes(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE batch_projects (
    id BIGINT NOT NULL AUTO_INCREMENT,
    intake_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_batch_projects_intake (intake_id),
    CONSTRAINT fk_batch_projects_intake FOREIGN KEY (intake_id) REFERENCES intakes(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE runs (
    id BIGINT NOT NULL AUTO_INCREMENT,
    batch_project_id BIGINT NOT NULL,
    run_at DATETIME(6) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_runs_due (status, run_at),
    CONSTRAINT fk_runs_batch_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS runs;
DROP TABLE IF EXISTS batch_projects;
DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS intakes;
