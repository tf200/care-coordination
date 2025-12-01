-- name: CreateAttachment :exec
INSERT INTO attachments (
    id,
    filekey,
    content_type
) VALUES (
    $1, $2, $3
);