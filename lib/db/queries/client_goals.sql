-- name: CreateClientGoal :exec
INSERT INTO client_goals (
    id,
    intake_form_id,
    client_id,
    title,
    description
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: ListGoalsByIntakeID :many
SELECT * FROM client_goals WHERE intake_form_id = $1;

-- name: ListGoalsByClientID :many
SELECT * FROM client_goals WHERE client_id = $1;

-- name: UpdateClientGoal :exec
UPDATE client_goals SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    updated_at = NOW()
WHERE id = $1;

-- name: LinkGoalsToClient :exec
UPDATE client_goals 
SET client_id = $1, updated_at = NOW() 
WHERE intake_form_id = $2 AND client_id IS NULL;

-- name: DeleteGoal :exec
DELETE FROM client_goals WHERE id = $1;
