-- =====================================================================
-- 06_people.sql: People and mandates: parties, districts, persons, curules
-- Depends on: 01, 02, 05
-- =====================================================================

-- +goose Up
-- +goose StatementBegin
CREATE TABLE party (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  name        text UNIQUE NOT NULL,
  acronym     text,
  created_at  timestamptz NOT NULL DEFAULT now(),
  created_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at  timestamptz NOT NULL DEFAULT now(),
  updated_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor
);

CREATE TRIGGER party_audit BEFORE UPDATE ON party
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE TABLE district (                    -- circunscripción
  id          smallint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  chamber     chamber_t NOT NULL,
  kind        text NOT NULL CHECK (kind IN ('nacional','territorial','indigena','afro',
                                            'internacional','paz','estatuto_oposicion')),
  department  text,
  name        text NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  created_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at  timestamptz NOT NULL DEFAULT now(),
  updated_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  UNIQUE (chamber, kind, name)
);

CREATE TRIGGER district_audit BEFORE UPDATE ON district
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE TABLE person (
  id             uuid PRIMARY KEY DEFAULT uuidv7(),
  full_name      text NOT NULL,
  name_norm      text GENERATED ALWAYS AS (lower(immutable_unaccent(full_name))) STORED,
  birth_date     date,
  photo_url      text,
  social_media   jsonb NOT NULL DEFAULT '[]'::jsonb
                   CHECK (valid_social_media(social_media)),
  merged_into_id uuid REFERENCES person,  -- set when deduplicated into another row
  created_at     timestamptz NOT NULL DEFAULT now(),
  created_by     uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at     timestamptz NOT NULL DEFAULT now(),
  updated_by     uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  CHECK (merged_into_id <> id)
);

CREATE TRIGGER person_audit BEFORE UPDATE ON person
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE INDEX person_name_trgm   ON person USING gin (name_norm gin_trgm_ops);
CREATE INDEX person_social_gin  ON person USING gin (social_media jsonb_path_ops);
-- e.g. WHERE social_media @> '[{"platform":"x","handle":"@nombre"}]'

CREATE TABLE mandate (                     -- curul: person x term x chamber
  id                  uuid PRIMARY KEY DEFAULT uuidv7(),
  person_id           uuid NOT NULL REFERENCES person,
  term_id             smallint NOT NULL REFERENCES term,
  chamber             chamber_t NOT NULL,
  district_id         smallint REFERENCES district,
  party_id            uuid REFERENCES party,
  start_date          date NOT NULL,
  end_date            date,
  replaces_mandate_id uuid REFERENCES mandate,
  status              text NOT NULL CHECK (status IN ('activo','reemplazado',
                                                      'renuncia','suspendido')),
  created_at          timestamptz NOT NULL DEFAULT now(),
  created_by          uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at          timestamptz NOT NULL DEFAULT now(),
  updated_by          uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  CONSTRAINT mandate_dates_ck CHECK (end_date IS NULL OR end_date >= start_date),
  CONSTRAINT mandate_no_overlap EXCLUDE USING gist (
    person_id WITH =, daterange(start_date, end_date, '[]') WITH &&)
);

CREATE TRIGGER mandate_audit BEFORE UPDATE ON mandate
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE INDEX ON mandate (term_id, chamber);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE mandate CASCADE;
DROP TABLE person CASCADE;
DROP TABLE district CASCADE;
DROP TABLE party CASCADE;

-- +goose StatementEnd