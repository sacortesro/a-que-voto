-- name: CreateParty :one
INSERT INTO party (name, acronym)
VALUES ($1, $2)
RETURNING *;

-- name: GetPartyByName :one
SELECT id
FROM party
WHERE name = $1;
