-- name: CreateChirp :one
INSERT INTO chirps (id, body, user_id, created_at, updated_at)
VALUES (
	GEN_RANDOM_UUID(),
	$1,
	$2,
	NOW(),
	NOW()
) RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirps ORDER BY created_at ASC;
