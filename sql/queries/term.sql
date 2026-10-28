-- name: CreateTerm :one
INSERT INTO term (label, start_date, end_date)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTermByLabel :one
SELECT id
FROM term
WHERE label = $1;

-- name: GetTerms :many
SELECT *
FROM term
ORDER BY start_date ASC
LIMIT 50;
