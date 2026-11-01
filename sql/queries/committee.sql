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

-- name: ListCommittees :many
SELECT *
FROM committee
ORDER BY chamber, name;

-- name: UpdateCommittee :one
UPDATE committee
SET kind = $2,
    number = $3
WHERE id = $1
RETURNING *;

-- name: ListCommitteeMembers :many
SELECT *
FROM committee_membership
WHERE committee_id = $1
ORDER BY start_date, mandate_id;

-- name: UpdateCommitteeMembership :one
UPDATE committee_membership
SET role = sqlc.arg(role),
    end_date = sqlc.arg(end_date)
WHERE mandate_id = sqlc.arg(mandate_id)
AND committee_id = sqlc.arg(committee_id)
AND start_date = sqlc.arg(start_date)
RETURNING *;
