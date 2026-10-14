-- name: CreateRun :one
INSERT INTO fetch_run 
(job, source_id, started_at, finished_at, status, items_seen, items_changed, error)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;
-- name: UpdateRun :one
UPDATE fetch_run
SET finished_at = sqlc.arg(finished_at),
    status = sqlc.arg(status),
    items_seen = sqlc.arg(items_seen),
    items_changed = sqlc.arg(items_changed),
    error = sqlc.arg(error)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: FailStaleRuns :execrows
UPDATE fetch_run
SET status = 'failed',
    finished_at = now(),
    error = 'stale: no run finished it'
WHERE job = sqlc.arg(job)
  AND status = 'running'
  AND started_at < sqlc.arg(started_before);
