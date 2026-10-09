-- =====================================================================
-- 01_foundation.sql: Extensions, domains and shared functions
-- =====================================================================

-- Requires PostgreSQL 18+ for native uuidv7().
-- On PG 13-17: replace uuidv7() with gen_random_uuid() across all files,
-- or install the pg_uuidv7 extension.

-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- ===== Domains & helpers =====

CREATE DOMAIN chamber_t AS text
  CHECK (VALUE IN ('senado','camara'));

-- Entity types that can appear in polymorphic references
-- (source_record, data_conflict, follow, event). All of them are UUID-keyed.
CREATE DOMAIN entity_type_t AS text
  CHECK (VALUE IN ('person','mandate','party','committee','bill',
                   'bill_document','vote_event'));

CREATE FUNCTION immutable_unaccent(text) RETURNS text
  LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT
  AS $$ SELECT public.unaccent('public.unaccent'::regdictionary, $1) $$;

-- social_media must be an array of objects, each with at least
-- "platform" and "url".
-- Example element:
--   {"platform":"x","handle":"@nombre","url":"https://x.com/nombre",
--    "source":"congresovisible","verified_at":"2026-10-01"}
CREATE FUNCTION valid_social_media(j jsonb) RETURNS boolean
  LANGUAGE sql IMMUTABLE PARALLEL SAFE
  AS $$
    SELECT jsonb_typeof(j) = 'array'
       AND NOT EXISTS (
         SELECT 1 FROM jsonb_array_elements(j) e
         WHERE jsonb_typeof(e) <> 'object'
            OR NOT (e ? 'platform' AND e ? 'url')
       )
  $$;

-- The actor performing the write must be set once per transaction:
--   SELECT set_config('app.actor_id', '<actor uuid>', true);
-- Any audited insert/update without it fails loudly.
CREATE FUNCTION app_actor() RETURNS uuid
  LANGUAGE plpgsql STABLE
  AS $$
  DECLARE v text := current_setting('app.actor_id', true);
  BEGIN
    IF v IS NULL OR v = '' THEN
      RAISE EXCEPTION 'app.actor_id is not set for this transaction';
    END IF;
    RETURN v::uuid;
  END $$;

-- Full-audit tables: stamp updated_*, and protect created_* from changes.
CREATE FUNCTION audit_on_update() RETURNS trigger
  LANGUAGE plpgsql
  AS $$
  BEGIN
    NEW.created_at := OLD.created_at;
    NEW.created_by := OLD.created_by;
    NEW.updated_at := now();
    NEW.updated_by := app_actor();
    RETURN NEW;
  END $$;

-- Append-only tables: corrections are delete + insert, never update.
CREATE FUNCTION forbid_update() RETURNS trigger
  LANGUAGE plpgsql
  AS $$
  BEGIN
    RAISE EXCEPTION '% is append-only', TG_TABLE_NAME;
  END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS audit_on_update() CASCADE;
DROP FUNCTION IF EXISTS forbid_update() CASCADE;
DROP FUNCTION IF EXISTS app_actor() CASCADE;
DROP FUNCTION IF EXISTS valid_social_media(jsonb) CASCADE;
DROP DOMAIN IF EXISTS entity_type_t CASCADE;
DROP DOMAIN IF EXISTS chamber_t CASCADE;
DROP EXTENSION IF EXISTS btree_gist CASCADE;
-- +goose StatementEnd