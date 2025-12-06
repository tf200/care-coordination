-- name: CreateClient :one
INSERT INTO clients (
    id,
    first_name,
    last_name,
    bsn,
    date_of_birth,
    phone_number,
    gender,
    registration_form_id,
    intake_form_id,
    care_type,
    referring_org_id,
    waiting_list_priority,
    status,
    assigned_location_id,
    coordinator_id,
    family_situation,
    limitations,
    focus_areas,
    goals,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
)
RETURNING id, first_name, last_name, bsn, date_of_birth, phone_number, gender, registration_form_id, intake_form_id, care_type, referring_org_id, status, assigned_location_id, coordinator_id, family_situation, limitations, focus_areas, goals, notes, created_at, updated_at;



-- name: GetClientByID :one
SELECT * FROM clients WHERE id = $1;

-- name: UpdateClient :one
UPDATE clients SET
    first_name = COALESCE($2, first_name),
    last_name = COALESCE($3, last_name),
    bsn = COALESCE($4, bsn),
    date_of_birth = COALESCE($5, date_of_birth),
    phone_number = COALESCE($6, phone_number),
    gender = COALESCE($7, gender),
    registration_form_id = COALESCE($8, registration_form_id),
    intake_form_id = COALESCE($9, intake_form_id),
    care_type = COALESCE($10, care_type),
    referring_org_id = COALESCE($11, referring_org_id),
    waiting_list_priority = COALESCE($12, waiting_list_priority),
    status = COALESCE($13, status),
    assigned_location_id = COALESCE($14, assigned_location_id),
    coordinator_id = COALESCE($15, coordinator_id),
    family_situation = COALESCE($16, family_situation),
    limitations = COALESCE($17, limitations),
    focus_areas = COALESCE($18, focus_areas),
    goals = COALESCE($19, goals),
    notes = COALESCE($20, notes),
    ambulatory_weekly_hours = COALESCE($21, ambulatory_weekly_hours),
    care_start_date = COALESCE($22, care_start_date),
    care_end_date = COALESCE($23, care_end_date),
    discharge_date = COALESCE($24, discharge_date),
    closing_report = COALESCE($25, closing_report),
    evaluation_report = COALESCE($26, evaluation_report),
    reason_for_discharge = COALESCE($27, reason_for_discharge),
    discharge_attachment_ids = COALESCE($28, discharge_attachment_ids),
    discharge_status = COALESCE($29, discharge_status),
    updated_at = NOW()
WHERE id = $1
RETURNING id;

-- name: ListWaitingListClients :many
SELECT
    c.id,
    c.first_name,
    c.last_name,
    c.bsn,
    c.date_of_birth,
    c.phone_number,
    c.gender,
    c.care_type,
    c.waiting_list_priority,
    c.focus_areas,
    c.notes,
    c.created_at,
    l.id AS location_id,
    l.name AS location_name,
    e.id AS coordinator_id,
    e.first_name AS coordinator_first_name,
    e.last_name AS coordinator_last_name,
    ro.name AS referring_org_name,
    COUNT(*) OVER() AS total_count
FROM clients c
JOIN locations l ON c.assigned_location_id = l.id
JOIN employees e ON c.coordinator_id = e.id
LEFT JOIN referring_orgs ro ON c.referring_org_id = ro.id
WHERE c.status = 'waiting_list'
    AND (sqlc.narg('search')::text IS NULL OR
         LOWER(c.first_name) LIKE LOWER('%' || sqlc.narg('search')::text || '%') OR
         LOWER(c.last_name) LIKE LOWER('%' || sqlc.narg('search')::text || '%') OR
         LOWER(c.first_name || ' ' || c.last_name) LIKE LOWER('%' || sqlc.narg('search')::text || '%'))
ORDER BY
    CASE c.waiting_list_priority
        WHEN 'high' THEN 1
        WHEN 'normal' THEN 2
        WHEN 'low' THEN 3
    END,
    c.created_at ASC
LIMIT $1 OFFSET $2;