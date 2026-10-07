-- +goose Up
-- Workshop belongs to an existing intake even before a BatchProject exists.
ALTER TABLE intakes ADD COLUMN workshop_settings_json JSON NULL AFTER status;

-- +goose Down
ALTER TABLE intakes DROP COLUMN workshop_settings_json;
