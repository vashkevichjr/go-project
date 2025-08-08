-- name: CreateHabit :one
INSERT INTO habits (user_id, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetHabitByID :one
SELECT * FROM habits
WHERE id = $1 AND user_id = $2;

-- name: ListHabitsByUserID :many
SELECT * FROM habits
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteHabit :exec
DELETE FROM habits
WHERE id = $1 AND user_id = $2;

-- name: UpdateHabit :one
UPDATE habits
SET name = $3, description = $4
WHERE id = $1 AND user_id = $2
RETURNING *;