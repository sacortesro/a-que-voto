-- =====================================================================
-- 05_periods.sql: Periods: cuatrienios (term) and legislaturas
-- Depends on: 01, 02
-- =====================================================================

-- +goose up
-- +goose StatementBegin

CREATE TABLE term (                        -- cuatrienio
  id          smallint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  label       text UNIQUE NOT NULL,       -- '2026-2030'
  start_date  date NOT NULL,
  end_date    date NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  created_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at  timestamptz NOT NULL DEFAULT now(),
  updated_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  CHECK (end_date > start_date),
  EXCLUDE USING gist (daterange(start_date, end_date, '[]') WITH &&)
);

CREATE TRIGGER term_audit BEFORE UPDATE ON term
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE TABLE legislatura (                 -- July 20 to June 20
  id          smallint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  term_id     smallint NOT NULL REFERENCES term,
  label       text UNIQUE NOT NULL,       -- '2026-2027'
  start_date  date NOT NULL,
  end_date    date NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  created_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  updated_at  timestamptz NOT NULL DEFAULT now(),
  updated_by  uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
  CHECK (end_date > start_date),
  EXCLUDE USING gist (daterange(start_date, end_date, '[]') WITH &&)
);

CREATE TRIGGER legislatura_audit BEFORE UPDATE ON legislatura
  FOR EACH ROW EXECUTE FUNCTION audit_on_update();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS legislatura CASCADE;
DROP TABLE IF EXISTS term CASCADE;

-- +goose StatementEnd