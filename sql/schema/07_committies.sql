-- =====================================================================
-- 07_committees.sql: Committees and committee membership
-- Depends on: 01, 02, 06
-- =====================================================================

-- +goose Up
-- +goose StatementBegin

CREATE TABLE committee (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  chamber     chamber_t NOT NULL,
  kind        text NOT NULL CHECK (kind IN ('constitucional','legal','accidental')),
  name        text NOT NULL,
  number      smallint,
  created_at  timestamptz NOT NULL DEFAULT now(),
  created_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at  timestamptz NOT NULL DEFAULT now(),
  updated_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  UNIQUE (chamber, name)
);

CREATE TRIGGER committee_audit BEFORE UPDATE ON committee
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE TABLE committee_membership (
  mandate_id   uuid NOT NULL REFERENCES mandate ON DELETE CASCADE,
  committee_id uuid NOT NULL REFERENCES committee,
  role         text NOT NULL DEFAULT 'miembro'
                 CHECK (role IN ('miembro','presidente','vicepresidente')),
  start_date   date NOT NULL,
  end_date     date,
  created_at   timestamptz NOT NULL DEFAULT now(),
  created_by   uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at   timestamptz NOT NULL DEFAULT now(),
  updated_by   uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  PRIMARY KEY (mandate_id, committee_id, start_date),
  CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE TRIGGER committee_membership_audit BEFORE UPDATE ON committee_membership
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE INDEX ON committee_membership (committee_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX committee_membership CASCADE;
DROP TRIGGER committee_membership_audit CASCADE;
DROP TRIGGER committee_audit CASCADE;
DROP TABLE committee_membership CASCADE;
DROP TABLE committee CASCADE;

-- +goose StatementEnd