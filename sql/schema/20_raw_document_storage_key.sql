-- +goose Up
-- +goose StatementBegin

ALTER TABLE raw_document DROP COLUMN storage_key;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE raw_document ADD COLUMN storage_key text NOT NULL DEFAULT '';
ALTER TABLE raw_document ALTER COLUMN storage_key DROP DEFAULT;

-- +goose StatementEnd
