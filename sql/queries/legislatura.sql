-- name: CreateLegislatura :one
INSERT INTO legislatura (term_id, label, start_date, end_date)
VALUES ($1, $2, $3, $4)
RETURNING *;