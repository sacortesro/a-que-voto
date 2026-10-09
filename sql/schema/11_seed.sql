-- =====================================================================
-- 11_seed.sql: Initial data
-- Depends on: 02
-- =====================================================================

-- +goose Up
-- +goose StatementBegin

INSERT INTO actor (kind, name) VALUES
  ('system', 'system:migration'),
  ('user',   'santiago');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM actor
WHERE kind IN ('system','user');

-- +goose StatementEnd
