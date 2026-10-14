-- name: CreateSourceRecord :one
INSERT INTO source_record
(source_id, entity_type, external_id, payload, payload_sha256, raw_document_id, first_seen_at, last_seen_at, entity_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;
