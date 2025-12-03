-- name: CreateRegistrationForm :exec
INSERT INTO registration_forms (
    id,
    first_name,
    last_name,
    bsn,
    gender,
    date_of_birth,
    reffering_org_id,
    care_type,
    registration_date,
    registration_reason,
    additional_notes,
    attachment_ids
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11, $12
);


-- name: ListRegistrationForms :many
SELECT r.id,
        r.first_name,
        r.last_name,
        r.bsn,
        r.date_of_birth,
        r.reffering_org_id,
        r.care_type,
        r.registration_date,
        r.registration_reason,
        r.additional_notes,
        r.attachment_ids,
        r.status,
        ro.name as org_name,
        ro.contact_person as org_contact_person,
        ro.phone_number as org_phone_number,
        ro.email as org_email,
        COUNT(r.id) OVER () AS total_count
FROM registration_forms r
LEFT JOIN referring_orgs ro ON r.reffering_org_id = ro.id
WHERE
    (
        -- If search term is NULL or empty, ignore filters
        sqlc.narg('search')::text IS NULL OR sqlc.narg('search')::text = '' OR
        -- Search by Org Name
        ro.name ILIKE '%' || sqlc.narg('search') || '%' OR
        -- Search by Client First Name
        r.first_name ILIKE '%' || sqlc.narg('search') || '%' OR
        -- Search by Client Last Name
        r.last_name ILIKE '%' || sqlc.narg('search') || '%'
    )
ORDER BY r.registration_date DESC
LIMIT $1 OFFSET $2;


-- name: GetRegistrationForm :one
SELECT * FROM registration_forms WHERE id = $1;
