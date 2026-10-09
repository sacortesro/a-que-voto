-- =====================================================================
-- 04_provenance.sql: Provenance: sources, fetch runs, raw documents, source records, conflicts
-- Depends on: 01, 02
-- =====================================================================

-- +goose Up
-- +goose StatementBegin

CREATE TABLE source (
  id          smallint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  code        text UNIQUE NOT NULL,       -- 'camara','senado','congresovisible','gaceta'
  base_url    text NOT NULL,
  priority    smallint NOT NULL,          -- lower = more trusted
  created_at  timestamptz NOT NULL DEFAULT now(),
  created_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at  timestamptz NOT NULL DEFAULT now(),
  updated_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor
);

CREATE TRIGGER source_audit BEFORE UPDATE ON source
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE TABLE fetch_run (
  id            uuid PRIMARY KEY DEFAULT uuidv7(),
  job           text NOT NULL,            -- 'congresistas.sync'
  source_id     smallint REFERENCES source,
  started_at    timestamptz NOT NULL DEFAULT now(),
  finished_at   timestamptz,
  status        text NOT NULL CHECK (status IN ('running','ok','failed','partial')),
  items_seen    int CHECK (items_seen >= 0),
  items_changed int CHECK (items_changed >= 0),
  error         text,
  created_at    timestamptz NOT NULL DEFAULT now(),
  created_by    uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at    timestamptz NOT NULL DEFAULT now(),
  updated_by    uuid NOT NULL DEFAULT app_actor() REFERENCES actor
);

CREATE TRIGGER fetch_run_audit BEFORE UPDATE ON fetch_run
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE INDEX ON fetch_run (job, started_at DESC);

CREATE TABLE raw_document (                -- append-only (content-addressed)
  id             uuid PRIMARY KEY DEFAULT uuidv7(),
  source_id      smallint NOT NULL REFERENCES source,
  url            text NOT NULL,
  content_sha256 bytea NOT NULL CHECK (octet_length(content_sha256) = 32),
  storage_key    text NOT NULL,           -- R2 key
  content_type   text,
  fetched_at     timestamptz NOT NULL,
  fetch_run_id   uuid REFERENCES fetch_run,
  created_at     timestamptz NOT NULL DEFAULT now(),
  created_by     uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  UNIQUE (source_id, url, content_sha256)
);

CREATE TRIGGER raw_document_no_update BEFORE UPDATE ON raw_document
  FOR EACH ROW EXECUTE FUNCTION forbid_update();

CREATE INDEX ON raw_document (fetch_run_id);
CREATE INDEX ON raw_document (content_sha256);

CREATE TABLE source_record (
  id              uuid PRIMARY KEY DEFAULT uuidv7(),
  source_id       smallint NOT NULL REFERENCES source,
  entity_type     entity_type_t NOT NULL,
  external_id     text NOT NULL,          -- the source's own id or URL slug
  payload         jsonb NOT NULL,
  payload_sha256  bytea NOT NULL CHECK (octet_length(payload_sha256) = 32),
  raw_document_id uuid REFERENCES raw_document,
  first_seen_at   timestamptz NOT NULL,
  last_seen_at    timestamptz NOT NULL,
  entity_id       uuid,                   -- linked final row (NULL until matched)
  created_at      timestamptz NOT NULL DEFAULT now(),
  created_by      uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at      timestamptz NOT NULL DEFAULT now(),
  updated_by      uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  UNIQUE (source_id, entity_type, external_id),
  CHECK (last_seen_at >= first_seen_at)
);

CREATE TRIGGER source_record_audit BEFORE UPDATE ON source_record
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE INDEX ON source_record (entity_type, entity_id);

CREATE TABLE data_conflict (
  id               uuid PRIMARY KEY DEFAULT uuidv7(),
  entity_type      entity_type_t NOT NULL,
  entity_id        uuid NOT NULL,
  field            text NOT NULL,
  candidate_values jsonb NOT NULL,        -- {"camara": "...", "congresovisible": "..."}
  status           text NOT NULL DEFAULT 'open'
                     CHECK (status IN ('open','resolved','ignored')),
  resolution       jsonb,
  created_at       timestamptz NOT NULL DEFAULT now(),
  created_by       uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at       timestamptz NOT NULL DEFAULT now(),
  updated_by       uuid NOT NULL DEFAULT app_actor() REFERENCES actor
);

CREATE TRIGGER data_conflict_audit BEFORE UPDATE ON data_conflict
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE UNIQUE INDEX data_conflict_open_uq ON data_conflict (entity_type, entity_id, field);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS data_conflict_audit ON data_conflict CASCADE;
DROP TRIGGER IF EXISTS source_record_audit ON source_record CASCADE;
DROP TRIGGER IF EXISTS raw_document_no_update ON raw_document CASCADE;
DROP TRIGGER IF EXISTS fetch_run_audit ON fetch_run CASCADE;
DROP TRIGGER IF EXISTS source_audit ON source CASCADE;

DROP TABLE IF EXISTS data_conflict CASCADE;
DROP TABLE IF EXISTS source_record CASCADE;
DROP TABLE IF EXISTS raw_document CASCADE;
DROP TABLE IF EXISTS fetch_run CASCADE;
DROP TABLE IF EXISTS source CASCADE;

-- +goose StatementEnd