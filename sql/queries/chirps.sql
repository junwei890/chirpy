-- name: CreateChirp :one
WITH chirpinsert AS (
	INSERT INTO chirps (id, body, user_id, created_at, updated_at)
	VALUES (
		GEN_RANDOM_UUID(),
		$1,
		$2,
		NOW(),
		NOW()
	) RETURNING *
)
SELECT chirpinsert.*, users.is_chirpy_red FROM chirpinsert
INNER JOIN users ON chirpinsert.user_id = users.id;

-- name: GetAllChirps :many
SELECT chirps.*, users.is_chirpy_red FROM chirps
INNER JOIN users ON chirps.user_id = users.id
ORDER BY chirps.created_at ASC;

-- name: GetOneChirp :one
SELECT chirps.*, users.is_chirpy_red FROM chirps
INNER JOIN users ON chirps.user_id = users.id
WHERE chirps.id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps WHERE id = $1;
