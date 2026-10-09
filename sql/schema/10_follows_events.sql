-- =====================================================================
-- 10_follows_events.sql: Follows and notification events
-- Depends on: 01, 02
-- =====================================================================

-- +goose Up
-- +goose StatementBegin

CREATE TABLE follow (                      -- insert/delete only
  actor_id    uuid NOT NULL REFERENCES actor,   -- the user following
  entity_type entity_type_t NOT NULL,
  entity_id   uuid NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  created_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  PRIMARY KEY (actor_id, entity_type, entity_id)
);

CREATE TRIGGER follow_no_update BEFORE UPDATE ON follow
  FOR EACH ROW EXECUTE FUNCTION forbid_update();

CREATE INDEX ON follow (entity_type, entity_id);

CREATE TABLE event (                       -- append-only, except notified_at
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  entity_type entity_type_t NOT NULL,
  entity_id   uuid NOT NULL,
  kind        text NOT NULL,              -- bill.stage_changed, vote.cast, ...
  payload     jsonb NOT NULL,
  dedupe_key  text UNIQUE NOT NULL,       -- makes re-runs idempotent
  notified_at timestamptz,
  created_at  timestamptz NOT NULL DEFAULT now(),
  created_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor
);

CREATE INDEX ON event (entity_type, entity_id, created_at DESC);
CREATE INDEX event_pending ON event (created_at) WHERE notified_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE event CASCADE;
DROP TABLE follow CASCADE;

-- +goose StatementEnd