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
