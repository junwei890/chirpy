-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
	GEN_RANDOM_UUID(),
	NOW(),
	NOW(),
	$1,
	$2
) RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserDetails :one
UPDATE users SET email = $1, hashed_password = $2, updated_at = NOW() WHERE id = $3 RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: UpdateRedUser :exec
UPDATE users SET is_chirpy_red = TRUE, updated_at = NOW() WHERE id = $1;
