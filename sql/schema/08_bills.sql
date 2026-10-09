-- =====================================================================
-- 08_bills.sql: Bills: proyectos, relations, authors, ponentes, timeline, documents
-- Depends on: 01-07
-- =====================================================================

-- +goose Up
-- +goose StatementBegin

CREATE TABLE bill (
  id                        uuid PRIMARY KEY DEFAULT uuidv7(),
  titulo                    text NOT NULL,
  titulo_norm               text GENERATED ALWAYS AS (lower(immutable_unaccent(titulo))) STORED,
  numero_camara             text,         -- '123/2026C'
  numero_senado             text,         -- '045/2026S'
  tipo                      text NOT NULL CHECK (tipo IN ('ley','acto_legislativo',
                                                           'ley_estatutaria')),   -- TODO: complete
  iniciativa                text NOT NULL CHECK (iniciativa IN ('legislativa','gubernamental',
                                                                'popular')),      -- TODO: complete
  camara_origen             chamber_t,
  radicado_at               date,
  current_stage             text NOT NULL REFERENCES stage,
  current_result            text NOT NULL CHECK (current_result IN ('APROBADO','RECHAZADO',
                                                                     'PENDIENTE')),
  tema                      text,
  term_id                   smallint REFERENCES term,
  legislatura_radicacion_id smallint REFERENCES legislatura,
  created_at                timestamptz NOT NULL DEFAULT now(),
  created_by                uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at                timestamptz NOT NULL DEFAULT now(),
  updated_by                uuid NOT NULL DEFAULT app_actor() REFERENCES actor
);

CREATE TRIGGER bill_audit BEFORE UPDATE ON bill
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE UNIQUE INDEX bill_numero_camara_uq ON bill (numero_camara) WHERE numero_camara IS NOT NULL;
CREATE UNIQUE INDEX bill_numero_senado_uq ON bill (numero_senado) WHERE numero_senado IS NOT NULL;
CREATE INDEX bill_titulo_trgm   ON bill   USING gin (titulo_norm gin_trgm_ops);

CREATE TABLE bill_relation (               -- insert/delete only
  bill_id         uuid NOT NULL REFERENCES bill ON DELETE CASCADE,
  related_bill_id uuid NOT NULL REFERENCES bill ON DELETE CASCADE,
  kind            text NOT NULL CHECK (kind IN ('acumulado','reradicado')),
  created_at      timestamptz NOT NULL DEFAULT now(),
  created_by      uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  PRIMARY KEY (bill_id, related_bill_id, kind),
  CHECK (bill_id <> related_bill_id)
);

CREATE TRIGGER bill_relation_no_update BEFORE UPDATE ON bill_relation
  FOR EACH ROW EXECUTE FUNCTION forbid_update();

CREATE INDEX ON bill_relation (related_bill_id);

CREATE TABLE bill_author (                 -- insert/delete only
  bill_id     uuid NOT NULL REFERENCES bill ON DELETE CASCADE,
  person_id   uuid NOT NULL REFERENCES person,
  role        text NOT NULL DEFAULT 'autor' CHECK (role IN ('autor','coautor')),
  created_at  timestamptz NOT NULL DEFAULT now(),
  created_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  PRIMARY KEY (bill_id, person_id, role)
);

CREATE TRIGGER bill_author_no_update BEFORE UPDATE ON bill_author
  FOR EACH ROW EXECUTE FUNCTION forbid_update();

CREATE INDEX ON bill_author (person_id);

CREATE TABLE bill_ponente (                -- insert/delete only
  bill_id       uuid NOT NULL REFERENCES bill ON DELETE CASCADE,
  mandate_id    uuid NOT NULL REFERENCES mandate,
  stage         text NOT NULL REFERENCES stage,
  committee_id  uuid REFERENCES committee, -- NULL = plenaria
  coordinador   boolean NOT NULL DEFAULT false,
  designated_at date,
  created_at    timestamptz NOT NULL DEFAULT now(),
  created_by    uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  PRIMARY KEY (bill_id, mandate_id, stage)
);

CREATE TRIGGER bill_ponente_no_update BEFORE UPDATE ON bill_ponente
  FOR EACH ROW EXECUTE FUNCTION forbid_update();

CREATE INDEX ON bill_ponente (mandate_id);

CREATE TABLE bill_stage_event (            -- append-only timeline
  id           uuid PRIMARY KEY DEFAULT uuidv7(),
  bill_id      uuid NOT NULL REFERENCES bill ON DELETE CASCADE,
  stage        text NOT NULL REFERENCES stage,
  result       text NOT NULL CHECK (result IN ('APROBADO','RECHAZADO','PENDIENTE')),
  occurred_at  date,
  committee_id uuid REFERENCES committee,
  source_id    smallint REFERENCES source,
  created_at   timestamptz NOT NULL DEFAULT now(),
  created_by   uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  UNIQUE NULLS NOT DISTINCT (bill_id, stage, result, occurred_at)
);

CREATE TRIGGER bill_stage_event_no_update BEFORE UPDATE ON bill_stage_event
  FOR EACH ROW EXECUTE FUNCTION forbid_update();

CREATE INDEX ON bill_stage_event (bill_id, occurred_at);

CREATE TABLE bill_document (
  id              uuid PRIMARY KEY DEFAULT uuidv7(),
  bill_id         uuid NOT NULL REFERENCES bill ON DELETE CASCADE,
  kind            text NOT NULL REFERENCES document_kind,
  gaceta_number   text,
  published_at    date,
  raw_document_id uuid REFERENCES raw_document,  -- PDF in R2
  text_extracted  boolean NOT NULL DEFAULT false,
  created_at      timestamptz NOT NULL DEFAULT now(),
  created_by      uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at      timestamptz NOT NULL DEFAULT now(),
  updated_by      uuid NOT NULL DEFAULT app_actor() REFERENCES actor
);

CREATE TRIGGER bill_document_audit BEFORE UPDATE ON bill_document
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE INDEX ON bill_document (bill_id);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX bill_document CASCADE;
DROP TRIGGER bill_document_audit CASCADE;
DROP INDEX bill_stage_event CASCADE;
DROP TRIGGER bill_stage_event_no_update CASCADE;
DROP INDEX bill_ponente CASCADE;
DROP TRIGGER bill_ponente_no_update CASCADE;
DROP INDEX bill_author CASCADE;
DROP TRIGGER bill_author_no_update CASCADE;
DROP INDEX bill_relation CASCADE;
DROP TRIGGER bill_relation_no_update CASCADE;
DROP INDEX bill_titulo_trgm CASCADE;
DROP INDEX bill_numero_senado_uq CASCADE;
DROP INDEX bill_numero_camara_uq CASCADE;
DROP TRIGGER bill_audit CASCADE;


DROP TABLE bill_document CASCADE;
DROP TABLE bill_stage_event CASCADE;
DROP TABLE bill_document CASCADE;
DROP TABLE bill_ponente CASCADE;
DROP TABLE bill_author CASCADE;
DROP TABLE bill_relation CASCADE;
DROP TABLE bill CASCADE;


-- +goose StatementEnd
