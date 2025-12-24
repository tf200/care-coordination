-- name: CreateReferringOrg :exec
INSERT INTO referring_orgs (
    id,
    name,
    contact_person,
    phone_number,
    email
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetReferringOrgByID :one
SELECT
    id,
    name,
    contact_person,
    phone_number,
    email,
    created_at,
    updated_at
FROM referring_orgs
WHERE id = $1;

-- name: ListReferringOrgs :many
SELECT
    id,
    name,
    contact_person,
    phone_number,
    email,
    created_at,
    updated_at,
    COUNT(id) OVER () AS total_count
FROM referring_orgs
WHERE
    (
        -- If search term is NULL or empty, ignore filters
        sqlc.narg('search')::text IS NULL OR sqlc.narg('search')::text = '' OR
        -- Search by Org Name
        name ILIKE '%' || sqlc.narg('search') || '%' OR
        -- Search by Contact Person
        contact_person ILIKE '%' || sqlc.narg('search') || '%' OR
        -- Search by Email
        email ILIKE '%' || sqlc.narg('search') || '%'
    )
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: ListReferringOrgsWithCounts :many
SELECT
    ro.id,
    ro.name,
    ro.contact_person,
    ro.phone_number,
    ro.email,
    ro.created_at,
    ro.updated_at,
    COUNT(CASE WHEN c.status = 'in_care' THEN 1 END)::bigint AS in_care_count,
    COUNT(CASE WHEN c.status = 'waiting_list' THEN 1 END)::bigint AS waiting_list_count,
    COUNT(CASE WHEN c.status = 'discharged' THEN 1 END)::bigint AS discharged_count,
    COUNT(ro.id) OVER () AS total_count
FROM referring_orgs ro
LEFT JOIN clients c ON c.referring_org_id = ro.id
WHERE
    (
        -- If search term is NULL or empty, ignore filters
        sqlc.narg('search')::text IS NULL OR sqlc.narg('search')::text = '' OR
        -- Search by Org Name
        ro.name ILIKE '%' || sqlc.narg('search') || '%' OR
        -- Search by Contact Person
        ro.contact_person ILIKE '%' || sqlc.narg('search') || '%' OR
        -- Search by Email
        ro.email ILIKE '%' || sqlc.narg('search') || '%'
    )
GROUP BY ro.id, ro.name, ro.contact_person, ro.phone_number, ro.email, ro.created_at, ro.updated_at
ORDER BY ro.name
LIMIT $1 OFFSET $2;

-- name: UpdateReferringOrg :exec
UPDATE referring_orgs
SET
    name = COALESCE(sqlc.narg('name'), name),
    contact_person = COALESCE(sqlc.narg('contact_person'), contact_person),
    phone_number = COALESCE(sqlc.narg('phone_number'), phone_number),
    email = COALESCE(sqlc.narg('email'), email),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteReferringOrg :exec
DELETE FROM referring_orgs
WHERE id = $1;

-- name: GetReferringOrgStats :one
SELECT 
    COUNT(DISTINCT ro.id) as total_orgs,
    COUNT(DISTINCT CASE WHEN c.status = 'in_care' THEN ro.id END) as orgs_with_in_care_clients,
    COUNT(DISTINCT CASE WHEN c.status = 'waiting_list' THEN ro.id END) as orgs_with_waitlist_clients,
    COUNT(c.id) as total_clients_referred
FROM referring_orgs ro
LEFT JOIN clients c ON c.referring_org_id = ro.id;