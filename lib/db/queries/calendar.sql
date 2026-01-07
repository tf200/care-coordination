-- name: CreateAppointment :one
INSERT INTO appointments (
    id, title, description, start_time, end_time, location, organizer_id, status, type, recurrence_rule
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetAppointment :one
SELECT * FROM appointments WHERE id = $1;

-- name: UpdateAppointment :one
UPDATE appointments
SET title = CASE WHEN sqlc.arg('title')::text <> '' THEN sqlc.arg('title')::text ELSE title END,
    description = COALESCE(sqlc.narg('description'), description),
    start_time = COALESCE(sqlc.narg('start_time'), start_time),
    end_time = COALESCE(sqlc.narg('end_time'), end_time),
    location = COALESCE(sqlc.narg('location'), location),
    status = COALESCE(sqlc.narg('status'), status),
    type = COALESCE(sqlc.narg('type'), type),
    recurrence_rule = COALESCE(sqlc.narg('recurrence_rule'), recurrence_rule),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteAppointment :exec
DELETE FROM appointments WHERE id = $1;

-- name: ListAppointmentsByOrganizer :many
SELECT * FROM appointments WHERE organizer_id = $1 ORDER BY start_time ASC;

-- name: ListAppointmentsByParticipant :many
SELECT a.* FROM appointments a
JOIN appointment_participants ap ON a.id = ap.appointment_id
WHERE ap.participant_id = $1 AND ap.participant_type = $2
ORDER BY a.start_time ASC;

-- name: AddAppointmentParticipant :exec
INSERT INTO appointment_participants (
    appointment_id, participant_id, participant_type
) VALUES (
    $1, $2, $3
);

-- name: RemoveAppointmentParticipants :exec
DELETE FROM appointment_participants WHERE appointment_id = $1;

-- name: ListAppointmentParticipants :many
SELECT * FROM appointment_participants WHERE appointment_id = $1;

-- name: CreateReminder :one
INSERT INTO reminders (
    id, user_id, title, description, due_time, is_completed
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetReminder :one
SELECT * FROM reminders WHERE id = $1;

-- name: UpdateReminder :one
UPDATE reminders
SET title = CASE WHEN sqlc.arg('title')::text <> '' THEN sqlc.arg('title')::text ELSE title END,
    description = COALESCE(sqlc.narg('description'), description),
    due_time = COALESCE(sqlc.narg('due_time'), due_time),
    is_completed = COALESCE(sqlc.narg('is_completed'), is_completed),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteReminder :exec
DELETE FROM reminders WHERE id = $1;

-- name: ListRemindersByUser :many
SELECT * FROM reminders WHERE user_id = $1 ORDER BY due_time ASC;

-- name: ListAppointmentsByRange :many
SELECT * FROM appointments 
WHERE organizer_id = $1 
AND start_time >= sqlc.arg('start_time')::timestamptz 
AND start_time <= sqlc.arg('end_time')::timestamptz
ORDER BY start_time ASC;

-- name: ListRemindersByRange :many
SELECT * FROM reminders 
WHERE user_id = $1 
AND due_time >= sqlc.arg('start_time')::timestamptz 
AND due_time <= sqlc.arg('end_time')::timestamptz
ORDER BY due_time ASC;

-- name: ListRecurringAppointments :many
SELECT * FROM appointments 
WHERE organizer_id = $1 
AND recurrence_rule IS NOT NULL 
AND recurrence_rule <> ''
AND start_time <= sqlc.arg('end_time')::timestamptz
ORDER BY start_time ASC;
