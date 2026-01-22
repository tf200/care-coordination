-- ============================================================
-- Dashboard
-- ============================================================

-- name: GetDashboardOverviewStats :one
SELECT
    (SELECT COUNT(*) FROM clients WHERE status = 'in_care') as total_active_clients,
    (SELECT COUNT(*) FROM clients WHERE status = 'waiting_list') as waiting_list_count,
    (SELECT COUNT(*) FROM registration_forms WHERE status = 'pending' AND is_deleted = FALSE) as pending_registrations,
    (SELECT COUNT(*) FROM employees e 
     JOIN user_roles ur ON e.user_id = ur.user_id 
     JOIN roles r ON ur.role_id = r.id 
     WHERE r.name = 'coordinator' AND e.is_deleted = FALSE) as total_coordinators,
    (SELECT COUNT(*) FROM employees WHERE is_deleted = FALSE) as total_employees,
    (SELECT COUNT(*) FROM incidents WHERE (status = 'pending' OR status = 'under_investigation') AND is_deleted = FALSE) as open_incidents;

-- name: GetCriticalAlertsData :one
SELECT
    -- Overdue evaluations (next_evaluation_date < today)
    (SELECT COUNT(*) FROM clients 
     WHERE status = 'in_care' 
     AND next_evaluation_date IS NOT NULL 
     AND next_evaluation_date < CURRENT_DATE) as overdue_evaluations,
    
    -- Care end date approaching (within 30 days)
    (SELECT COUNT(*) FROM clients 
     WHERE status = 'in_care' 
     AND care_end_date IS NOT NULL 
     AND care_end_date <= (CURRENT_DATE + INTERVAL '30 days')::date
     AND care_end_date >= CURRENT_DATE) as care_ending_soon,
    
    -- Open incidents (pending or under_investigation)
    (SELECT COUNT(*) FROM incidents 
     WHERE (status = 'pending' OR status = 'under_investigation') 
     AND is_deleted = FALSE) as open_incidents,
    
    -- Severe incidents count (for description)
    (SELECT COUNT(*) FROM incidents 
     WHERE (status = 'pending' OR status = 'under_investigation') 
     AND incident_severity = 'severe'
     AND is_deleted = FALSE) as severe_incidents,
    
    -- Moderate incidents count (for description)
    (SELECT COUNT(*) FROM incidents 
     WHERE (status = 'pending' OR status = 'under_investigation') 
     AND incident_severity = 'moderate'
     AND is_deleted = FALSE) as moderate_incidents,
    
    -- High priority waiting list
    (SELECT COUNT(*) FROM clients 
     WHERE status = 'waiting_list' 
     AND waiting_list_priority = 'high') as high_priority_waiting,
    
    -- Pending location transfers
    (SELECT COUNT(*) FROM client_location_transfers 
     WHERE status = 'pending') as pending_transfers;

-- name: GetPipelineStats :one
SELECT
    (SELECT COUNT(*) FROM registration_forms WHERE is_deleted = FALSE) as registrations,
    (SELECT COUNT(*) FROM intake_forms) as intakes,
    (SELECT COUNT(*) FROM clients WHERE status = 'waiting_list') as waiting_list,
    (SELECT COUNT(*) FROM clients WHERE status = 'in_care') as in_care,
    (SELECT COUNT(*) FROM clients WHERE status = 'discharged') as discharged;

-- name: GetCareTypeDistribution :one
SELECT
    (SELECT COUNT(*) FROM clients WHERE status = 'in_care' AND care_type = 'protected_living') as protected_living,
    (SELECT COUNT(*) FROM clients WHERE status = 'in_care' AND care_type = 'semi_independent_living') as semi_independent_living,
    (SELECT COUNT(*) FROM clients WHERE status = 'in_care' AND care_type = 'independent_assisted_living') as independent_assisted_living,
    (SELECT COUNT(*) FROM clients WHERE status = 'in_care' AND care_type = 'ambulatory_care') as ambulatory_care,
    (SELECT COUNT(*) FROM clients WHERE status = 'in_care') as total;

-- name: GetLocationCapacityList :many
SELECT
    l.id,
    l.name,
    l.capacity,
    l.occupied
FROM locations l
WHERE l.is_deleted = FALSE;

-- name: GetLocationCapacityTotals :one
SELECT
    COALESCE(SUM(capacity), 0)::bigint as total_capacity,
    COALESCE(SUM(occupied), 0)::bigint as total_occupied
FROM locations
WHERE is_deleted = FALSE;

-- name: GetTodayAppointmentsForEmployee :many
SELECT
    a.id,
    a.title,
    a.type,
    a.start_time,
    a.end_time,
    a.location,
    -- Get client participant if exists
    COALESCE(
        (SELECT c.id FROM clients c 
         JOIN appointment_participants ap ON ap.participant_id = c.id 
         WHERE ap.appointment_id = a.id AND ap.participant_type = 'client' 
         LIMIT 1), 
        ''
    )::text as client_id,
    COALESCE(
        (SELECT CONCAT(c.first_name, ' ', c.last_name) FROM clients c 
         JOIN appointment_participants ap ON ap.participant_id = c.id 
         WHERE ap.appointment_id = a.id AND ap.participant_type = 'client' 
         LIMIT 1), 
        ''
    )::text as client_name
FROM appointments a
WHERE 
    DATE(a.start_time AT TIME ZONE 'UTC') = CURRENT_DATE
    AND (
        a.organizer_id = $1
        OR EXISTS (
            SELECT 1 FROM appointment_participants ap 
            WHERE ap.appointment_id = a.id 
            AND ap.participant_id = $1 
            AND ap.participant_type = 'employee'
        )
    )
ORDER BY a.start_time ASC;

-- name: GetEvaluationStats :one
SELECT
    -- Total in-care clients that need evaluations
    (SELECT COUNT(*) FROM clients WHERE status = 'in_care')::bigint as total,
    -- Completed evaluations (submitted status)
    (SELECT COUNT(DISTINCT client_id) FROM client_evaluations WHERE status = 'submitted')::bigint as completed,
    -- Overdue evaluations (next_evaluation_date < today)
    (SELECT COUNT(*) FROM clients 
     WHERE status = 'in_care' 
     AND next_evaluation_date IS NOT NULL 
     AND next_evaluation_date < CURRENT_DATE)::bigint as overdue,
    -- Due soon (next_evaluation_date within 7 days)
    (SELECT COUNT(*) FROM clients 
     WHERE status = 'in_care' 
     AND next_evaluation_date IS NOT NULL 
     AND next_evaluation_date >= CURRENT_DATE
     AND next_evaluation_date <= (CURRENT_DATE + INTERVAL '7 days')::date)::bigint as due_soon;

-- name: GetDashboardDischargeStats :one
SELECT
    -- Discharged this month
    (SELECT COUNT(*) FROM clients 
     WHERE status = 'discharged' 
     AND discharge_date IS NOT NULL
     AND EXTRACT(MONTH FROM discharge_date) = EXTRACT(MONTH FROM CURRENT_DATE)
     AND EXTRACT(YEAR FROM discharge_date) = EXTRACT(YEAR FROM CURRENT_DATE))::bigint as this_month,
    -- Discharged this year
    (SELECT COUNT(*) FROM clients 
     WHERE status = 'discharged' 
     AND discharge_date IS NOT NULL
     AND EXTRACT(YEAR FROM discharge_date) = EXTRACT(YEAR FROM CURRENT_DATE))::bigint as this_year,
    -- Total discharged clients
    (SELECT COUNT(*) FROM clients WHERE status = 'discharged')::bigint as total_discharged,
    -- Planned discharges (those with a reason indicating planned)
    (SELECT COUNT(*) FROM clients 
     WHERE status = 'discharged' 
     AND reason_for_discharge IS NOT NULL 
     )::bigint as planned_discharges,
    -- Average days in care (from care_start_date to discharge_date)
    COALESCE((SELECT AVG(EXTRACT(DAY FROM (discharge_date - care_start_date)))::bigint 
     FROM clients 
     WHERE status = 'discharged' 
     AND care_start_date IS NOT NULL 
     AND discharge_date IS NOT NULL), 0)::bigint as avg_days_in_care;

-- ============================================================
-- Coordinator Dashboard
-- ============================================================

-- name: GetCoordinatorUrgentAlertsData :one
SELECT
    -- Overdue evaluations for coordinator's clients
    (SELECT COUNT(*) FROM clients c1
     WHERE c1.coordinator_id = $1
     AND c1.status = 'in_care' 
     AND c1.next_evaluation_date IS NOT NULL 
     AND c1.next_evaluation_date < CURRENT_DATE)::bigint as overdue_evaluations,
    
    -- Contracts expiring within 7 days for coordinator's clients
    (SELECT COUNT(*) FROM clients c2
     WHERE c2.coordinator_id = $1
     AND c2.status = 'in_care' 
     AND c2.care_end_date IS NOT NULL 
     AND c2.care_end_date >= CURRENT_DATE
     AND c2.care_end_date <= (CURRENT_DATE + INTERVAL '7 days')::date)::bigint as expiring_contracts,
    
    -- Draft evaluations not completed by coordinator
    (SELECT COUNT(*) FROM client_evaluations ce
     WHERE ce.coordinator_id = $1
     AND ce.status = 'draft')::bigint as draft_evaluations,
    
    -- Unresolved incidents for coordinator's clients
    (SELECT COUNT(*) FROM incidents i
     WHERE i.coordinator_id = $1
     AND (i.status = 'pending' OR i.status = 'under_investigation')
     AND i.is_deleted = FALSE)::bigint as unresolved_incidents,
    
    -- Waiting list clients > 60 days for coordinator
    (SELECT COUNT(*) FROM clients c3
     WHERE c3.coordinator_id = $1
     AND c3.status = 'waiting_list' 
     AND c3.waiting_list_date IS NOT NULL
     AND c3.waiting_list_date < (CURRENT_DATE - INTERVAL '60 days')::date)::bigint as long_waiting;


-- name: GetCoordinatorOverdueEvaluationClients :many
SELECT id, first_name, last_name
FROM clients
WHERE coordinator_id = $1
AND status = 'in_care'
AND next_evaluation_date IS NOT NULL
AND next_evaluation_date < CURRENT_DATE
LIMIT 5;

-- name: GetCoordinatorExpiringContractClients :many
SELECT id, first_name, last_name
FROM clients
WHERE coordinator_id = $1
AND status = 'in_care'
AND care_end_date IS NOT NULL
AND care_end_date >= CURRENT_DATE
AND care_end_date <= (CURRENT_DATE + INTERVAL '7 days')::date
LIMIT 5;

-- name: GetCoordinatorDraftEvaluationClients :many
SELECT c.id, c.first_name, c.last_name
FROM clients c
JOIN client_evaluations ce ON ce.client_id = c.id
WHERE ce.coordinator_id = $1
AND ce.status = 'draft'
LIMIT 5;

-- name: GetCoordinatorUnresolvedIncidentClients :many
SELECT DISTINCT c.id, c.first_name, c.last_name
FROM clients c
JOIN incidents i ON i.client_id = c.id
WHERE i.coordinator_id = $1
AND (i.status = 'pending' OR i.status = 'under_investigation')
AND i.is_deleted = FALSE
LIMIT 5;

-- name: GetCoordinatorLongWaitingClients :many
SELECT id, first_name, last_name
FROM clients
WHERE coordinator_id = $1
AND status = 'waiting_list'
AND waiting_list_date IS NOT NULL
AND waiting_list_date < (CURRENT_DATE - INTERVAL '60 days')::date
LIMIT 5;

-- name: GetCoordinatorTodaySchedule :many
SELECT
    a.id,
    a.title,
    a.type,
    a.start_time,
    a.end_time,
    a.location,
    a.status,
    -- Get client participant if exists
    COALESCE(
        (SELECT c.id FROM clients c 
         JOIN appointment_participants ap ON ap.participant_id = c.id 
         WHERE ap.appointment_id = a.id AND ap.participant_type = 'client' 
         LIMIT 1), 
        ''
    )::text as client_id,
    COALESCE(
        (SELECT CONCAT(c.first_name, ' ', c.last_name) FROM clients c 
         JOIN appointment_participants ap ON ap.participant_id = c.id 
         WHERE ap.appointment_id = a.id AND ap.participant_type = 'client' 
         LIMIT 1), 
        ''
    )::text as client_name,
    -- Get location info if linked
    COALESCE(
        (SELECT l.id FROM locations l WHERE l.name = a.location LIMIT 1),
        ''
    )::text as location_id,
    COALESCE(a.location, '')::text as location_name
FROM appointments a
WHERE 
    DATE(a.start_time AT TIME ZONE 'UTC') = CURRENT_DATE
    AND (
        a.organizer_id = $1
        OR EXISTS (
            SELECT 1 FROM appointment_participants ap 
            WHERE ap.appointment_id = a.id 
            AND ap.participant_id = $1 
            AND ap.participant_type = 'employee'
        )
    )
ORDER BY a.start_time ASC;

-- name: GetCoordinatorStats :one
SELECT
    (SELECT COUNT(*) FROM clients c1
     WHERE c1.coordinator_id = $1 
     AND c1.status = 'in_care')::bigint as my_active_clients,
    
    (SELECT COUNT(*) FROM clients c2
     WHERE c2.coordinator_id = $1 
     AND c2.status = 'in_care' 
     AND c2.next_evaluation_date IS NOT NULL 
     AND c2.next_evaluation_date >= CURRENT_DATE
     AND c2.next_evaluation_date <= (CURRENT_DATE + INTERVAL '30 days')::date)::bigint as my_upcoming_evaluations,
    
    (SELECT COUNT(*) FROM intake_forms i
     WHERE i.coordinator_id = $1 
     AND i.status = 'pending')::bigint as my_pending_intakes,
    
    (SELECT COUNT(*) FROM clients c3
     WHERE c3.coordinator_id = $1 
     AND c3.status = 'waiting_list')::bigint as my_waiting_list_clients;

-- name: GetCoordinatorReminders :many
SELECT
    r.id,
    r.title,
    r.due_time
FROM reminders r
WHERE r.user_id = $1
AND r.is_completed = FALSE
ORDER BY r.due_time ASC
LIMIT 10;

-- name: GetCoordinatorClients :many
SELECT
    c.id,
    c.first_name,
    c.last_name,
    c.care_type,
    c.status,
    c.care_end_date,
    c.next_evaluation_date,
    l.name as location_name
FROM clients c
LEFT JOIN locations l ON c.assigned_location_id = l.id
WHERE c.coordinator_id = $1
AND c.status IN ('in_care', 'waiting_list')
ORDER BY c.care_end_date ASC NULLS LAST, c.next_evaluation_date ASC NULLS LAST;

-- name: GetCoordinatorGoalsProgress :one
WITH coordinator_goals AS (
    SELECT cg.id as goal_id
    FROM client_goals cg
    JOIN clients c ON cg.client_id = c.id
    WHERE c.coordinator_id = $1
    AND c.status = 'in_care'
),
latest_progress AS (
    SELECT DISTINCT ON (gpl.goal_id) 
        gpl.goal_id,
        gpl.status
    FROM goal_progress_logs gpl
    WHERE gpl.goal_id IN (SELECT goal_id FROM coordinator_goals)
    ORDER BY gpl.goal_id, gpl.created_at DESC
)
SELECT
    (SELECT COUNT(*) FROM coordinator_goals)::bigint as total,
    COALESCE((SELECT COUNT(*) FROM latest_progress WHERE status IN ('on_track', 'in_progress', 'starting')), 0)::bigint as on_track,
    COALESCE((SELECT COUNT(*) FROM latest_progress WHERE status IN ('delayed', 'stagnant', 'deteriorating')), 0)::bigint as delayed,
    COALESCE((SELECT COUNT(*) FROM latest_progress WHERE status = 'achieved'), 0)::bigint as achieved,
    (
        (SELECT COUNT(*) FROM coordinator_goals) - 
        (SELECT COUNT(*) FROM latest_progress WHERE status != 'not_started')
    )::bigint as not_started;

-- name: GetCoordinatorIncidents :many
SELECT
    i.id,
    i.incident_type,
    i.incident_severity,
    i.incident_date,
    i.status,
    c.first_name as client_first_name,
    c.last_name as client_last_name
FROM incidents i
JOIN clients c ON i.client_id = c.id
WHERE i.coordinator_id = $1
AND i.is_deleted = FALSE
ORDER BY i.incident_date DESC
LIMIT 10;
