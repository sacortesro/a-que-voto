-- +goose Up
-- +goose StatementBegin

ALTER TABLE actor
  ADD COLUMN role text NOT NULL DEFAULT 'member' CHECK (role IN ('member','admin'));

INSERT INTO actor (kind, name, role)
VALUES ('user', 'admin', 'admin')
ON CONFLICT (name) DO UPDATE SET role = 'admin';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE actor DROP COLUMN role;

-- +goose StatementEnd
