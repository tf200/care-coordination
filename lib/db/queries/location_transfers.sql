-- name: CreateLocationTransfer :one
INSERT INTO client_location_transfers (
    id,
    client_id,
    from_location_id,
    to_location_id,
    current_coordinator_id,
    new_coordinator_id,
    transfer_date,
    reason
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id, client_id, from_location_id, to_location_id, current_coordinator_id, new_coordinator_id, transfer_date;


-- name: ListLocationTransfers :many
SELECT
    clt.id,
    clt.client_id,
    clt.from_location_id,
    clt.to_location_id,
    clt.current_coordinator_id,
    clt.new_coordinator_id,
    clt.transfer_date,
    clt.reason,
    clt.status,
    clt.rejection_reason,
    c.first_name AS client_first_name,
    c.last_name AS client_last_name,
    l_from.name AS from_location_name,
    l_to.name AS to_location_name,
    e_current.first_name AS current_coordinator_first_name,
    e_current.last_name AS current_coordinator_last_name,
    e_new.first_name AS new_coordinator_first_name,
    e_new.last_name AS new_coordinator_last_name,
    COUNT(*) OVER() AS total_count
FROM client_location_transfers clt
JOIN clients c ON clt.client_id = c.id
LEFT JOIN locations l_from ON clt.from_location_id = l_from.id
LEFT JOIN locations l_to ON clt.to_location_id = l_to.id
LEFT JOIN employees e_current ON clt.current_coordinator_id = e_current.id
LEFT JOIN employees e_new ON clt.new_coordinator_id = e_new.id
WHERE
    (sqlc.narg('search')::text IS NULL OR
     c.first_name ILIKE '%' || sqlc.narg('search') || '%' OR
     c.last_name ILIKE '%' || sqlc.narg('search') || '%' OR
     CONCAT(c.first_name, ' ', c.last_name) ILIKE '%' || sqlc.narg('search') || '%'
    )   
ORDER BY clt.transfer_date DESC
LIMIT $1 OFFSET $2;

-- name: GetLocationTransferByID :one
SELECT
    clt.id,
    clt.client_id,
    clt.from_location_id,
    clt.to_location_id,
    clt.current_coordinator_id,
    clt.new_coordinator_id,
    clt.transfer_date,
    clt.reason,
    clt.status,
    clt.rejection_reason,
    c.first_name AS client_first_name,
    c.last_name AS client_last_name,
    l_from.name AS from_location_name,
    l_to.name AS to_location_name,
    e_current.first_name AS current_coordinator_first_name,
    e_current.last_name AS current_coordinator_last_name,
    e_new.first_name AS new_coordinator_first_name,
    e_new.last_name AS new_coordinator_last_name
FROM client_location_transfers clt
JOIN clients c ON clt.client_id = c.id
LEFT JOIN locations l_from ON clt.from_location_id = l_from.id
LEFT JOIN locations l_to ON clt.to_location_id = l_to.id
LEFT JOIN employees e_current ON clt.current_coordinator_id = e_current.id
LEFT JOIN employees e_new ON clt.new_coordinator_id = e_new.id
WHERE clt.id = $1;

-- name: ConfirmLocationTransfer :exec
UPDATE client_location_transfers
SET status = 'approved', transfer_date = NOW(), updated_at = NOW()
WHERE id = $1 AND status = 'pending';

-- name: RefuseLocationTransfer :exec
UPDATE client_location_transfers
SET status = 'rejected', rejection_reason = $2, updated_at = NOW()
WHERE id = $1 AND status = 'pending';

-- name: UpdateLocationTransfer :exec
UPDATE client_location_transfers
SET
    to_location_id = COALESCE(sqlc.narg('to_location_id'), to_location_id),
    new_coordinator_id = COALESCE(sqlc.narg('new_coordinator_id'), new_coordinator_id),
    reason = COALESCE(sqlc.narg('reason'), reason),
    updated_at = NOW()
WHERE id = $1 AND status = 'pending';

-- name: GetLocationTransferStats :one
SELECT 
    COUNT(*) as total_count,
    COUNT(*) FILTER (WHERE status = 'pending') as pending_count,
    COUNT(*) FILTER (WHERE status = 'approved') as approved_count,
    COUNT(*) FILTER (WHERE status = 'rejected') as rejected_count,
    CASE 
        WHEN COUNT(*) FILTER (WHERE status IN ('approved', 'rejected')) > 0 THEN 
            ROUND((COUNT(*) FILTER (WHERE status = 'approved')::DECIMAL / COUNT(*) FILTER (WHERE status IN ('approved', 'rejected'))::DECIMAL) * 100, 2)
        ELSE 0
    END as approval_rate
FROM client_location_transfers;