-- name: CreatePerson :one
INSERT INTO person (full_name, birth_date, photo_url, social_media, merged_into_id, id_congreso_visible)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPersons :many
SELECT id, name_norm, id_congreso_visible
FROM person
ORDER BY id_congreso_visible;

-- name: UpdatePerson :one
UPDATE person
SET birth_date = $1,
    social_media = $2
WHERE id = $3
RETURNING *;
