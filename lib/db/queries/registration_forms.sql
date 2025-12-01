-- name: CreateRegistrationForm :exec
INSERT INTO registration_forms (
    id,
    first_name,
    last_name,
    bsn,
    date_of_birth,
    org_name,
    org_contact_person,
    org_phone_number,
    org_email,
    care_type,
    coordinator_id,
    registration_date,
    registration_reason,
    additional_notes,
    attachment_ids
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, 
    $10, $11, $12, $13, $14, $15
);




-- name: ListRegistrationForms :many
SELECT r.id,
        r.first_name,
        r.last_name,
        r.bsn,
        r.date_of_birth,
        r.org_name,
        r.care_type,
        r.coordinator_id,
        e.first_name AS coordinator_first_name,
        e.last_name AS coordinator_last_name,
        r.registration_date,
        r.registration_reason,
        r.attachment_ids
FROM registration_forms r
JOIN employees e ON r.coordinator_id = e.id;