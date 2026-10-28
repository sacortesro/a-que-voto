-- name: CreateRawDocument :one
INSERT INTO raw_document
(source_id, url, content_sha256, storage_key, content_type, fetched_at, fetch_run_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetLatestRawDocumentHash :one
SELECT content_sha256  
  FROM raw_document
 WHERE source_id = $1
 AND url = $2
 ORDER BY fetched_at DESC
 LIMIT 1;

-- name: ListRawDocumentsBySource :many
SELECT *
FROM raw_document
WHERE source_id = $1
ORDER BY fetched_at DESC
LIMIT 20;
