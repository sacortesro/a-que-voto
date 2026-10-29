-- name: CreateDistrict :one
INSERT INTO district (chamber, kind, department, name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetDistrict :one
SELECT id
FROM district
WHERE chamber = $1
AND kind = $2
AND department IS NOT DISTINCT FROM sqlc.narg(department)::text;

-- name: ListDistricts :many
SELECT *
FROM district
ORDER BY id;
