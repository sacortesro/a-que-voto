-- =====================================================================
-- 02_actor.sql: Actors: who performed each write (users, jobs, system)
-- Depends on: 01
-- =====================================================================

-- +goose Up
-- +goose StatementBegin

CREATE TABLE actor (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  kind        text NOT NULL CHECK (kind IN ('user','job','system')),
  name        text UNIQUE NOT NULL,       -- 'santiago', 'job:congresistas.sync'
  created_at  timestamptz NOT NULL DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE actor;

-- +goose StatementEnd