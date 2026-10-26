-- name: CreateCommittee :one
INSERT INTO committee (chamber, kind, name, number)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: InsertNewCommitteeMember :one
INSERT INTO committee_membership (mandate_id, committee_id, role, start_date, end_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCommitteeByID :one
SELECT id
FROM committee
WHERE id = $1;

-- name: GetCommitteeMembership :one
SELECT mandate_id, committee_id, start_date
FROM committee_membership
WHERE mandate_id = $1
AND committee_id = $2
AND start_date = $3;