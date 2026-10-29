-- +goose Up
-- +goose StatementBegin

ALTER TABLE party 
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE party DROP COLUMN is_active;

-- +goose StatementEnd