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
    notes,
    evaluation_interval_weeks
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
ORDER BY i.created_at DESC  
LIMIT $1 OFFSET $2;



-- name: GetIntakeForm :one
SELECT * FROM intake_forms WHERE id = $1;

-- name: GetIntakeFormWithDetails :one
SELECT
    i.id,
    i.registration_form_id,
    i.intake_date,
    i.intake_time,
    i.location_id,
    i.coordinator_id,
    i.family_situation,
    i.main_provider,
    i.limitations,
    i.focus_areas,
    i.notes,
    i.evaluation_interval_weeks,
    i.status,
    i.created_at,
    i.updated_at,
    r.first_name as client_first_name,
    r.last_name as client_last_name,
    r.bsn as client_bsn,
    r.care_type,
    ro.name as org_name,
    l.name as location_name,
    e.first_name as coordinator_first_name,
    e.last_name as coordinator_last_name,
    EXISTS (SELECT 1 FROM clients c WHERE c.intake_form_id = i.id) AS has_client
FROM intake_forms i
LEFT JOIN registration_forms r ON i.registration_form_id = r.id
LEFT JOIN referring_orgs ro ON r.reffering_org_id = ro.id
LEFT JOIN locations l ON i.location_id = l.id
LEFT JOIN employees e ON i.coordinator_id = e.id
WHERE i.id = $1;

-- name: UpdateIntakeFormStatus :exec
UPDATE intake_forms SET status = $2, updated_at = NOW() WHERE id = $1;

-- name: UpdateIntakeForm :exec
UPDATE intake_forms SET
    intake_date = COALESCE(sqlc.narg('intake_date'), intake_date),
    intake_time = COALESCE(sqlc.narg('intake_time'), intake_time),
    location_id = COALESCE(sqlc.narg('location_id'), location_id),
    coordinator_id = COALESCE(sqlc.narg('coordinator_id'), coordinator_id),
    family_situation = COALESCE(sqlc.narg('family_situation'), family_situation),
    main_provider = COALESCE(sqlc.narg('main_provider'), main_provider),
    limitations = COALESCE(sqlc.narg('limitations'), limitations),
    focus_areas = COALESCE(sqlc.narg('focus_areas'), focus_areas),
    notes = COALESCE(sqlc.narg('notes'), notes),
    evaluation_interval_weeks = COALESCE(sqlc.narg('evaluation_interval_weeks'), evaluation_interval_weeks),
    status = COALESCE(sqlc.narg('status'), status),
    updated_at = NOW()
WHERE id = $1;

-- name: GetIntakeStats :one
SELECT 
    COUNT(*) as total_count,
    COUNT(*) FILTER (WHERE status = 'pending') as pending_count,
    CASE 
        WHEN COUNT(*) > 0 THEN 
            ROUND((COUNT(*) FILTER (WHERE status = 'completed')::DECIMAL / COUNT(*)::DECIMAL) * 100, 2)
        ELSE 0
    END as conversion_percentage
FROM intake_forms;