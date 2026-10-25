-- name: CreateToken :one
INSERT INTO api_token (actor_id, token_hash, expires_at)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: RevokeToken :exec
UPDATE api_token SET revoked_at = now() WHERE id = $1;

-- name: GetActorByTokenHash :one
SELECT a.*
FROM api_token t
JOIN actor a ON a.id = t.actor_id
WHERE t.token_hash = $1
  AND t.revoked_at IS NULL
  AND t.expires_at > now();
