-- name: CreatePerson :one
INSERT INTO person (full_name, birth_date, photo_url, social_media, merged_into_id, id_congreso_visible)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPersonsCongresoVisible :many
SELECT id, name_norm, id_congreso_visible
FROM person
WHERE merged_into_id IS NULL
AND id_congreso_visible IS NOT NULL
ORDER BY id_congreso_visible;

-- name: UpdatePersonDetail :one
UPDATE person
SET birth_date = $1,
    social_media = $2
WHERE id = $3
RETURNING *;

-- name: GetPersonIDByName :one
SELECT id  
FROM person
WHERE name_norm = lower(immutable_unaccent(@name::text))
AND merged_into_id IS NULL;

-- name: ListPersons :many
SELECT *
FROM person
WHERE merged_into_id IS NULL
ORDER BY name_norm, id
LIMIT $1
OFFSET $2;

-- name: GetPersonByID :one
SELECT *
FROM person
WHERE id = $1
AND merged_into_id IS NULL;
