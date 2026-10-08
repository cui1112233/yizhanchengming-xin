-- +goose Up
ALTER TABLE batch_projects
    ADD COLUMN archived_at DATETIME(6) NULL,
    ADD COLUMN archived_by_user_id BIGINT UNSIGNED NULL,
    ADD KEY idx_batch_projects_archive_updated (archived_at, updated_at, id),
    ADD KEY idx_batch_projects_archived_by (archived_by_user_id),
    ADD CONSTRAINT fk_batch_projects_archived_by
        FOREIGN KEY (archived_by_user_id) REFERENCES auth_users(id) ON DELETE SET NULL;

-- +goose Down
-- +goose StatementBegin
SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'batch project archive migration is irreversible';
-- +goose StatementEnd
