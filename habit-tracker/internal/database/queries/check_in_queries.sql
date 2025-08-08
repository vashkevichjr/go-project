-- name: CreateCheckIn :one
INSERT INTO check_ins (habit_id, check_in_date)
VALUES ($1, $2)
RETURNING *;