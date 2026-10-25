-- name: CreateCommittee :one
INSERT INTO committee (chamber, kind, name, number)
VALUES ($1, $2, $3, $4)
RETURNING *;