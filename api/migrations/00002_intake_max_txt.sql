-- +goose Up

ALTER TABLE intake_books
    ADD COLUMN max_txt INT UNSIGNED NOT NULL DEFAULT 4000 AFTER source_platform_name;

-- +goose Down

ALTER TABLE intake_books
    DROP COLUMN max_txt;
