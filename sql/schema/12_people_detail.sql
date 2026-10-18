-- +goose Up
-- +goose StatementBegin
ALTER TABLE person ADD COLUMN id_congreso_visible INTEGER;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

ALTER TABLE person DROP COLUMN id_congreso_visible;

-- +goose StatementEnd
