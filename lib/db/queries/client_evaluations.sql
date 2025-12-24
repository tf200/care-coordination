-- name: CreateClientEvaluation :one
INSERT INTO client_evaluations (
    id,
    client_id,
    coordinator_id,
    evaluation_date,
    overall_notes
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: CreateGoalProgressLog :exec
INSERT INTO goal_progress_logs (
    id,
    evaluation_id,
    goal_id,
    status,
    progress_notes
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: UpdateClientNextEvaluationDate :exec
UPDATE clients 
SET next_evaluation_date = $2, updated_at = NOW() 
WHERE id = $1;

-- name: GetClientEvaluationHistory :many
SELECT 
    e.id as evaluation_id,
    e.evaluation_date,
    e.overall_notes,
    g.id as goal_id,
    g.title as goal_title,
    l.status,
    l.progress_notes,
    emp.first_name as coordinator_first_name,
    emp.last_name as coordinator_last_name
FROM client_evaluations e
JOIN goal_progress_logs l ON e.id = l.evaluation_id
JOIN client_goals g ON l.goal_id = g.id
JOIN employees emp ON e.coordinator_id = emp.id
WHERE e.client_id = $1
ORDER BY e.evaluation_date DESC, g.title ASC;

-- name: GetCriticalEvaluations :many
SELECT 
    c.id,
    c.first_name,
    c.last_name,
    c.next_evaluation_date,
    c.evaluation_interval_weeks,
    l.name as location_name,
    e.first_name as coordinator_first_name,
    e.last_name as coordinator_last_name,
    COUNT(*) OVER() as total_count
FROM clients c
JOIN locations l ON c.assigned_location_id = l.id
JOIN employees e ON c.coordinator_id = e.id
WHERE c.status = 'in_care' 
  AND c.next_evaluation_date IS NOT NULL
  AND c.next_evaluation_date <= (CURRENT_DATE + INTERVAL '7 days')::date
ORDER BY c.next_evaluation_date ASC
LIMIT $1 OFFSET $2;

-- name: GetScheduledEvaluations :many
SELECT 
    c.id,
    c.first_name,
    c.last_name,
    c.next_evaluation_date,
    c.evaluation_interval_weeks,
    l.name as location_name,
    e.first_name as coordinator_first_name,
    e.last_name as coordinator_last_name,
    COUNT(*) OVER() as total_count
FROM clients c
JOIN locations l ON c.assigned_location_id = l.id
JOIN employees e ON c.coordinator_id = e.id
WHERE c.status = 'in_care' 
  AND c.next_evaluation_date IS NOT NULL
  AND c.next_evaluation_date > (CURRENT_DATE + INTERVAL '7 days')::date
  AND c.next_evaluation_date <= (CURRENT_DATE + INTERVAL '30 days')::date
ORDER BY c.next_evaluation_date ASC
LIMIT $1 OFFSET $2;

-- name: GetRecentEvaluationsGlobal :many
SELECT 
    e.id as evaluation_id,
    e.evaluation_date,
    c.first_name as client_first_name,
    c.last_name as client_last_name,
    emp.first_name as coordinator_first_name,
    emp.last_name as coordinator_last_name,
    (SELECT COUNT(*) FROM goal_progress_logs WHERE evaluation_id = e.id) as total_goals,
    (SELECT COUNT(*) FROM goal_progress_logs WHERE evaluation_id = e.id AND status = 'achieved') as goals_achieved,
    COUNT(*) OVER() as total_count
FROM client_evaluations e
JOIN clients c ON e.client_id = c.id
JOIN employees emp ON e.coordinator_id = emp.id
ORDER BY e.evaluation_date DESC
LIMIT $1 OFFSET $2;

-- name: GetLastClientEvaluation :many
SELECT 
    e.id as evaluation_id,
    e.evaluation_date,
    e.overall_notes,
    emp.first_name as coordinator_first_name,
    emp.last_name as coordinator_last_name,
    g.id as goal_id,
    g.title as goal_title,
    l.status,
    l.progress_notes
FROM client_evaluations e
JOIN goal_progress_logs l ON e.id = l.evaluation_id
JOIN client_goals g ON l.goal_id = g.id
JOIN employees emp ON e.coordinator_id = emp.id
WHERE e.client_id = $1
ORDER BY e.evaluation_date DESC, g.title ASC
LIMIT (SELECT COUNT(*) FROM goal_progress_logs WHERE evaluation_id = (
    SELECT id FROM client_evaluations WHERE client_id = $1 ORDER BY evaluation_date DESC LIMIT 1
));
