-- ============================================================
-- Incidents
-- ============================================================

-- name: CreateIncident :exec
INSERT INTO incidents (
    id,
    client_id,
    incident_date,
    incident_time,
    incident_type,
    incident_severity,
    location_id,
    coordinator_id,
    incident_description,
    action_taken,
    other_parties,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
);


-- name: ListIncidents :many
SELECT i.*,
       c.first_name AS client_first_name,
       c.last_name AS client_last_name,
       l.name AS location_name,
       e.first_name AS coordinator_first_name,
       e.last_name AS coordinator_last_name,
       COUNT(*) OVER() as total_count
FROM incidents i
JOIN clients c ON i.client_id = c.id
JOIN locations l ON i.location_id = l.id
JOIN employees e ON i.coordinator_id = e.id
WHERE i.is_deleted = FALSE
AND (
  sqlc.narg('search')::text IS NULL OR
  c.first_name ILIKE '%' || sqlc.narg('search')::text || '%' OR
  c.last_name ILIKE '%' || sqlc.narg('search')::text || '%' OR
  CONCAT(c.first_name, ' ', c.last_name) ILIKE '%' || sqlc.narg('search')::text || '%'
)
ORDER BY i.incident_date DESC
LIMIT $1 OFFSET $2;

-- name: GetIncidentStats :one
SELECT 
    COUNT(*) as total_count,
    -- Counts by severity
    COUNT(*) FILTER (WHERE incident_severity = 'minor') as minor_count,
    COUNT(*) FILTER (WHERE incident_severity = 'moderate') as moderate_count,
    COUNT(*) FILTER (WHERE incident_severity = 'severe') as severe_count,
    -- Counts by status
    COUNT(*) FILTER (WHERE status = 'pending') as pending_count,
    COUNT(*) FILTER (WHERE status = 'under_investigation') as under_investigation_count,
    COUNT(*) FILTER (WHERE status = 'completed') as completed_count,
    -- Counts by type
    COUNT(*) FILTER (WHERE incident_type = 'aggression') as aggression_count,
    COUNT(*) FILTER (WHERE incident_type = 'medical_emergency') as medical_emergency_count,
    COUNT(*) FILTER (WHERE incident_type = 'safety_concern') as safety_concern_count,
    COUNT(*) FILTER (WHERE incident_type = 'unwanted_behavior') as unwanted_behavior_count,
    COUNT(*) FILTER (WHERE incident_type = 'other') as other_type_count
FROM incidents
WHERE is_deleted = FALSE;

-- name: GetIncident :one
SELECT i.*,
       c.first_name AS client_first_name,
       c.last_name AS client_last_name,
       l.name AS location_name,
       e.first_name AS coordinator_first_name,
       e.last_name AS coordinator_last_name
FROM incidents i
JOIN clients c ON i.client_id = c.id
JOIN locations l ON i.location_id = l.id
JOIN employees e ON i.coordinator_id = e.id
WHERE i.id = $1 AND i.is_deleted = FALSE;

-- name: UpdateIncident :exec
UPDATE incidents
SET 
    incident_date = COALESCE(sqlc.narg('incident_date')::DATE, incident_date),
    incident_time = COALESCE(sqlc.narg('incident_time')::TIME, incident_time),
    incident_type = COALESCE(sqlc.narg('incident_type')::incident_type_enum, incident_type),
    incident_severity = COALESCE(sqlc.narg('incident_severity')::incident_severity_enum, incident_severity),
    location_id = COALESCE(sqlc.narg('location_id')::TEXT, location_id),
    coordinator_id = COALESCE(sqlc.narg('coordinator_id')::TEXT, coordinator_id),
    incident_description = COALESCE(sqlc.narg('incident_description')::TEXT, incident_description),
    action_taken = COALESCE(sqlc.narg('action_taken')::TEXT, action_taken),
    other_parties = CASE 
        WHEN sqlc.narg('other_parties')::TEXT = '' THEN NULL
        ELSE COALESCE(sqlc.narg('other_parties')::TEXT, other_parties)
    END,
    status = COALESCE(sqlc.narg('status')::incident_status_enum, status),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND is_deleted = FALSE;

-- name: SoftDeleteIncident :exec
UPDATE incidents
SET 
    is_deleted = TRUE,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;