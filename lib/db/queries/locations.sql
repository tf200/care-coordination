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
    (sqlc.narg('search')::text IS NULL OR
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