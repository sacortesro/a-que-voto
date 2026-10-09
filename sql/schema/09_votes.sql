-- =====================================================================
-- 09_votes.sql: Votes: vote events and individual nominal votes
-- Depends on: 01, 02, 06, 07, 08
-- =====================================================================

-- +goose Up
-- +goose StatementBegin

CREATE TABLE vote_event (
  id            uuid PRIMARY KEY DEFAULT uuidv7(),
  bill_id       uuid REFERENCES bill,     -- NULL for non-bill votes
  chamber       chamber_t NOT NULL,
  committee_id  uuid REFERENCES committee, -- NULL = plenaria
  held_at       date NOT NULL,
  subject       text NOT NULL,            -- 'articulado', 'informe de ponencia', ...
  method        text NOT NULL CHECK (method IN ('nominal','ordinaria')),
  yes_count     int CHECK (yes_count >= 0),
  no_count      int CHECK (no_count >= 0),
  abstain_count int CHECK (abstain_count >= 0),
  outcome       text CHECK (outcome IN ('aprobado','negado')),
  created_at    timestamptz NOT NULL DEFAULT now(),
  created_by    uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at    timestamptz NOT NULL DEFAULT now(),
  updated_by    uuid NOT NULL DEFAULT app_actor() REFERENCES actor
);

CREATE TRIGGER vote_event_audit BEFORE UPDATE ON vote_event
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE INDEX ON vote_event (bill_id);

CREATE TABLE vote_cast (                   -- append-only, only for nominal votes
  vote_event_id uuid NOT NULL REFERENCES vote_event ON DELETE CASCADE,
  mandate_id    uuid NOT NULL REFERENCES mandate,
  value         text NOT NULL CHECK (value IN ('si','no','abstencion',
                                               'ausente','impedimento')),
  created_at    timestamptz NOT NULL DEFAULT now(),
  created_by    uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  PRIMARY KEY (vote_event_id, mandate_id)
);

CREATE TRIGGER vote_cast_no_update BEFORE UPDATE ON vote_cast
  FOR EACH ROW EXECUTE FUNCTION forbid_update();

CREATE INDEX ON vote_cast (mandate_id);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX vote_cast CASCADE;
DROP INDEX vote_event CASCADE;

DROP TRIGGER vote_cast_no_update CASCADE;
DROP TRIGGER vote_event_audit CASCADE;

DROP TABLE vote_cast CASCADE;
DROP TABLE vote_event CASCADE;

-- +goose StatementEnd


