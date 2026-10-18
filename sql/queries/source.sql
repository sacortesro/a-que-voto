-- name: CreateSource :one
INSERT INTO source (code, base_url, priority)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSourceByCode :one
SELECT *
FROM source
WHERE code = $1;
