-- name: CreateTerm :one
INSERT INTO term (label, start_date, end_date)
VALUES ($1, $2, $3)
RETURNING *;

