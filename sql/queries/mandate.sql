-- name: CreatePersonMandate :one
INSERT INTO mandate (person_id, term_id, chamber, district_id, party_id, start_date, end_date, replaces_mandate_id, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetMandate :one
SELECT *
FROM mandate
WHERE person_id = $1
AND term_id = $2
AND chamber = $3;

