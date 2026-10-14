-- name: CreateRawDocument :one
INSERT INTO raw_document
(source_id, url, content_sha256, storage_key, content_type, fetched_at, fetch_run_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
