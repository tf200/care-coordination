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
    updated_at = NOW()
WHERE id = $1
RETURNING id;