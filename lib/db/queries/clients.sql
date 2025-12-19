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
    first_name = COALESCE(sqlc.narg('first_name'), first_name),
    last_name = COALESCE(sqlc.narg('last_name'), last_name),
    bsn = COALESCE(sqlc.narg('bsn'), bsn),
    date_of_birth = COALESCE(sqlc.narg('date_of_birth'), date_of_birth),
    phone_number = COALESCE(sqlc.narg('phone_number'), phone_number),
    gender = COALESCE(sqlc.narg('gender')::gender_enum, gender),
    registration_form_id = COALESCE(sqlc.narg('registration_form_id'), registration_form_id),
    intake_form_id = COALESCE(sqlc.narg('intake_form_id'), intake_form_id),
    care_type = COALESCE(sqlc.narg('care_type')::care_type_enum, care_type),
    referring_org_id = COALESCE(sqlc.narg('referring_org_id'), referring_org_id),
    waiting_list_priority = COALESCE(sqlc.narg('waiting_list_priority')::waiting_list_priority_enum, waiting_list_priority),
    status = COALESCE(sqlc.narg('status')::client_status_enum, status),
    assigned_location_id = COALESCE(sqlc.narg('assigned_location_id'), assigned_location_id),
    coordinator_id = COALESCE(sqlc.narg('coordinator_id'), coordinator_id),
    family_situation = COALESCE(sqlc.narg('family_situation'), family_situation),
    limitations = COALESCE(sqlc.narg('limitations'), limitations),
    focus_areas = COALESCE(sqlc.narg('focus_areas'), focus_areas),
    goals = COALESCE(sqlc.narg('goals'), goals),
    notes = COALESCE(sqlc.narg('notes'), notes),
    ambulatory_weekly_hours = COALESCE(sqlc.narg('ambulatory_weekly_hours'), ambulatory_weekly_hours),
    care_start_date = COALESCE(sqlc.narg('care_start_date'), care_start_date),
    care_end_date = COALESCE(sqlc.narg('care_end_date'), care_end_date),
    discharge_date = COALESCE(sqlc.narg('discharge_date'), discharge_date),
    closing_report = COALESCE(sqlc.narg('closing_report'), closing_report),
    evaluation_report = COALESCE(sqlc.narg('evaluation_report'), evaluation_report),
    reason_for_discharge = COALESCE(sqlc.narg('reason_for_discharge')::discharge_reason_enum, reason_for_discharge),
    discharge_attachment_ids = COALESCE(sqlc.narg('discharge_attachment_ids'), discharge_attachment_ids),
    discharge_status = COALESCE(sqlc.narg('discharge_status')::discharge_status_enum, discharge_status),
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

-- name: ListInCareClients :many
SELECT
    c.id,
    c.first_name,
    c.last_name,
    c.bsn,
    c.date_of_birth,
    c.phone_number,
    c.gender,
    c.care_type,
    c.care_start_date,
    c.care_end_date,
    c.ambulatory_weekly_hours,
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
WHERE c.status = 'in_care'
    AND (sqlc.narg('search')::text IS NULL OR
         LOWER(c.first_name) LIKE LOWER('%' || sqlc.narg('search')::text || '%') OR
         LOWER(c.last_name) LIKE LOWER('%' || sqlc.narg('search')::text || '%') OR
         LOWER(c.first_name || ' ' || c.last_name) LIKE LOWER('%' || sqlc.narg('search')::text || '%'))
ORDER BY c.care_start_date DESC
LIMIT $1 OFFSET $2;

-- name: ListDischargedClients :many
SELECT
    c.id,
    c.first_name,
    c.last_name,
    c.bsn,
    c.date_of_birth,
    c.phone_number,
    c.gender,
    c.care_type,
    c.care_start_date,
    c.discharge_date,
    c.reason_for_discharge,
    c.discharge_status,
    c.closing_report,
    c.evaluation_report,
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
WHERE c.discharge_status IS NOT NULL
    AND (sqlc.narg('search')::text IS NULL OR
         LOWER(c.first_name) LIKE LOWER('%' || sqlc.narg('search')::text || '%') OR
         LOWER(c.last_name) LIKE LOWER('%' || sqlc.narg('search')::text || '%') OR
         LOWER(c.first_name || ' ' || c.last_name) LIKE LOWER('%' || sqlc.narg('search')::text || '%'))
    AND (sqlc.narg('discharge_status')::discharge_status_enum IS NULL OR
         c.discharge_status = sqlc.narg('discharge_status')::discharge_status_enum)
ORDER BY c.discharge_date DESC
LIMIT $1 OFFSET $2;

-- name: UpdateClientByRegistrationFormID :exec
UPDATE clients SET
    first_name = COALESCE(sqlc.narg('first_name'), first_name),
    last_name = COALESCE(sqlc.narg('last_name'), last_name),
    bsn = COALESCE(sqlc.narg('bsn'), bsn),
    date_of_birth = COALESCE(sqlc.narg('date_of_birth'), date_of_birth),
    gender = COALESCE(sqlc.narg('gender')::gender_enum, gender),
    care_type = COALESCE(sqlc.narg('care_type')::care_type_enum, care_type),
    referring_org_id = COALESCE(sqlc.narg('referring_org_id'), referring_org_id),
    updated_at = NOW()
WHERE registration_form_id = $1;

-- name: UpdateClientByIntakeFormID :exec
UPDATE clients SET
    coordinator_id = COALESCE(sqlc.narg('coordinator_id'), coordinator_id),
    assigned_location_id = COALESCE(sqlc.narg('assigned_location_id'), assigned_location_id),
    family_situation = COALESCE(sqlc.narg('family_situation'), family_situation),
    limitations = COALESCE(sqlc.narg('limitations'), limitations),
    focus_areas = COALESCE(sqlc.narg('focus_areas'), focus_areas),
    goals = COALESCE(sqlc.narg('goals'), goals),
    notes = COALESCE(sqlc.narg('notes'), notes),
    updated_at = NOW()
WHERE intake_form_id = $1;