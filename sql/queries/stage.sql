-- name: GetStages :many
SELECT *
FROM stage
ORDER BY sort_order ASC;
