-- +goose Up
-- +goose StatementBegin

CREATE INDEX raw_document_source_fetched_idx ON raw_document (source_id, fetched_at DESC);
CREATE INDEX raw_document_source_url_fetched_idx ON raw_document (source_id, url, fetched_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX raw_document_source_url_fetched_idx;
DROP INDEX raw_document_source_fetched_idx;

-- +goose StatementEnd
