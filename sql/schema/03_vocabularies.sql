-- =====================================================================
-- 03_vocabularies.sql: Curated vocabularies: legislative stages and document kinds
-- Depends on: 01, 02
-- =====================================================================

-- +goose Up
-- +goose StatementBegin

CREATE TABLE stage (
  slug        text PRIMARY KEY,
  label       text NOT NULL,
  sort_order  smallint NOT NULL UNIQUE,
  created_at  timestamptz NOT NULL DEFAULT now(),
  created_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at  timestamptz NOT NULL DEFAULT now(),
  updated_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor
);

CREATE TRIGGER stage_audit BEFORE UPDATE ON stage
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE TABLE document_kind (
  slug        text PRIMARY KEY,           -- texto_radicado, ponencia_primer_debate, ...
  label       text NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  created_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at  timestamptz NOT NULL DEFAULT now(),
  updated_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor
);

CREATE TRIGGER document_kind_audit BEFORE UPDATE ON document_kind
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER document_kind_audit ON document_kind CASCADE;
DROP TRIGGER stage_audit ON stage CASCADE;
DROP TABLE document_kind CASCADE;
DROP TABLE stage CASCADE;

-- +goose StatementEnd

