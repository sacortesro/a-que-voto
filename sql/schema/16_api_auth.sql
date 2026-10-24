-- +goose Up
-- +goose StatementBegin
CREATE TABLE api_token (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    actor_id uuid NOT NULL REFERENCES actor,
    token_hash bytea NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,    

    created_at          timestamptz NOT NULL DEFAULT now(),
    created_by          uuid NOT NULL DEFAULT app_actor() REFERENCES actor,
    updated_at          timestamptz NOT NULL DEFAULT now(),
    updated_by          uuid NOT NULL DEFAULT app_actor() REFERENCES actor
    
);

CREATE TRIGGER api_token_audit BEFORE UPDATE ON api_token
    FOR EACH ROW EXECUTE FUNCTION audit_on_update();

CREATE INDEX ON api_token (actor_id);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE api_token CASCADE;

-- +goose StatementEnd
