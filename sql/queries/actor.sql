-- name: UpsertActor :one
INSERT INTO actor (kind, name)
VALUES (sqlc.arg(kind), sqlc.arg(name))
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING *;

-- name: GetActorByID :one
SELECT * FROM actor
WHERE id = sqlc.arg(id);
