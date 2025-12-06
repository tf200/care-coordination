-- name: CreateIntakeForm :exec
INSERT INTO intake_forms (
    id,
    registration_form_id,
    intake_date,
    intake_time,
    location_id,
    coordinator_id,
    family_situation,
    main_provider,
    limitations,
    focus_areas,
    goals,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
);

-- name: ListIntakeForms :many
SELECT
    i.id,
    i.registration_form_id,
    i.intake_date,
    i.intake_time,
    i.location_id,
    i.coordinator_id,
    i.main_provider,
    i.status,
    r.first_name,
    r.last_name,
    r.bsn,
    r.care_type,
    ro.name as org_name,
    l.name as location_name,
    e.first_name as coordinator_first_name,
    e.last_name as coordinator_last_name,
    COUNT(i.id) OVER () AS total_count
FROM intake_forms i
LEFT JOIN registration_forms r ON i.registration_form_id = r.id
LEFT JOIN referring_orgs ro ON r.reffering_org_id = ro.id
LEFT JOIN locations l ON i.location_id = l.id
LEFT JOIN employees e ON i.coordinator_id = e.id
WHERE
    (
        -- If search term is NULL or empty, ignore filters
        $3::text IS NULL OR $3::text = '' OR
        -- Search by client first name
        r.first_name ILIKE '%' || $3 || '%' OR
        -- Search by client last name
        r.last_name ILIKE '%' || $3 || '%' OR
        -- Search by org name
        ro.name ILIKE '%' || $3 || '%'
    )
ORDER BY i.intake_date DESC, i.intake_time DESC
LIMIT $1 OFFSET $2;



-- name: GetIntakeForm :one
SELECT * FROM intake_forms WHERE id = $1;

-- name: UpdateIntakeFormStatus :exec
UPDATE intake_forms SET status = $2, updated_at = NOW() WHERE id = $1;