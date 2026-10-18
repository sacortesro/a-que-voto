-- name: CreatePerson :one
INSERT INTO person (full_name, birth_date, photo_url, social_media, merged_into_id, id_congreso_visible)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
