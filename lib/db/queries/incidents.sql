-- name: CreateIncident :exec
INSERT INTO incidents (
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
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
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
WHERE
(
  sqlc.narg('search')::text IS NULL OR
  c.first_name ILIKE '%' || sqlc.narg('search')::text || '%' OR
  c.last_name ILIKE '%' || sqlc.narg('search')::text || '%' OR
  CONCAT(c.first_name, ' ', c.last_name) ILIKE '%' || sqlc.narg('search')::text || '%'
)
ORDER BY i.incident_date DESC
LIMIT $1 OFFSET $2;