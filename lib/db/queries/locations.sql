


-- name: CreateLocation :exec
INSERT INTO locations (
   id,
   name,
   postal_code,
   address,
   capacity,
   occupied
   )
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListLocations :many
SELECT
    l.id,
    l.name,
    l.postal_code,
    l.address,
    l.capacity,
    l.occupied,
    COUNT(*) OVER() as total_count
FROM locations l
WHERE
    l.is_deleted = FALSE
    AND (sqlc.narg('search')::text IS NULL OR
     LOWER(l.name) LIKE LOWER('%' || sqlc.narg('search')::text || '%') OR
     LOWER(l.postal_code) LIKE LOWER('%' || sqlc.narg('search')::text || '%') OR
     LOWER(l.address) LIKE LOWER('%' || sqlc.narg('search')::text || '%'))
ORDER BY l.name
LIMIT $1 OFFSET $2;

-- name: IncrementLocationOccupied :exec
UPDATE locations
SET occupied = occupied + 1, updated_at = NOW()
WHERE id = $1;

-- name: DecrementLocationOccupied :exec
UPDATE locations
SET occupied = occupied - 1, updated_at = NOW()
WHERE id = $1 AND occupied > 0;

-- name: UpdateLocation :exec
UPDATE locations SET
    name = COALESCE(sqlc.narg('name'), name),
    postal_code = COALESCE(sqlc.narg('postal_code'), postal_code),
    address = COALESCE(sqlc.narg('address'), address),
    capacity = COALESCE(sqlc.narg('capacity'), capacity),
    occupied = COALESCE(sqlc.narg('occupied'), occupied),
    updated_at = NOW()
WHERE id = $1;

-- name: SoftDeleteLocation :exec
UPDATE locations SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1;

-- name: GetLocationCapacityStats :one
SELECT 
    COALESCE(SUM(l.capacity), 0) as total_capacity,
    COALESCE(COUNT(c.id) FILTER (WHERE c.status = 'in_care'), 0) as capacity_used,
    COALESCE(SUM(l.capacity), 0) - COALESCE(COUNT(c.id) FILTER (WHERE c.status = 'in_care'), 0) as free_capacity
FROM locations l
LEFT JOIN clients c ON c.assigned_location_id = l.id
WHERE l.is_deleted = FALSE;