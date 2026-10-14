-- name: CreateDataConflict :one
INSERT INTO data_conflict
(entity_type, entity_id, field, candidate_values, status, resolution)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;