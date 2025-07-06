-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens(
  token, 
  created_at, 
  updated_at, 
  user_id,
  expires_at,
  revoked_at
)
VALUES($1, NOW(), NOW(), $2, $3, NULL
);

-- name: GetUserFromRefreshToken :one
SELECT users.*
FROM users
JOIN refresh_tokens ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = $1;

-- name: UpdateRefreshToken :exec
UPDATE refresh_tokens
SET token = $1, updated_at = NOW()
WHERE user_id = $2;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE token = $1;

