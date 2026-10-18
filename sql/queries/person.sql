-- name: CreatePerson :one
INSERT INTO person (full_name, birth_date, photo_url, social_media, merged_into_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
