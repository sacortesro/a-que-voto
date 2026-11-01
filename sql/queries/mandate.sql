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

-- name: GetMandateByID :one
SELECT id
FROM mandate
WHERE id = $1;

-- name: ListMandateByPersonID :many
SELECT *
FROM mandate
WHERE person_id = $1
AND daterange(start_date, end_date, '[]')
    && daterange(sqlc.arg(from_date)::date, sqlc.arg(to_date)::date, '[]')
ORDER BY start_date;
