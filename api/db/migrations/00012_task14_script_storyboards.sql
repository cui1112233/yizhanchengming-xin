-- +goose Up
CREATE TABLE script_storyboard_documents (
    batch_project_id BIGINT NOT NULL,
    book_id BIGINT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    source_stage_run_id BIGINT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (batch_project_id, book_id),
    CONSTRAINT fk_script_storyboard_document_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_script_storyboard_document_book FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
    CONSTRAINT fk_script_storyboard_document_stage FOREIGN KEY (source_stage_run_id) REFERENCES stage_runs(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE script_storyboard_cards (
    id BIGINT NOT NULL AUTO_INCREMENT,
    batch_project_id BIGINT NOT NULL,
    book_id BIGINT NOT NULL,
    position INT NOT NULL,
    title VARCHAR(255) NOT NULL DEFAULT '',
    content MEDIUMTEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_script_storyboard_card_position (batch_project_id, book_id, position),
    KEY idx_script_storyboard_card_scope (batch_project_id, book_id, position),
    CONSTRAINT fk_script_storyboard_card_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_script_storyboard_card_book FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS script_storyboard_cards;
DROP TABLE IF EXISTS script_storyboard_documents;
