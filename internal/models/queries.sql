-- name: GetMessage :one
SELECT * FROM messages
WHERE id = ? LIMIT 1;

-- name: ListMessages :many
SELECT * FROM messages;

-- name: CreateMessage :one
INSERT INTO messages (
  id, text
) VALUES (
  ?, ?
)
RETURNING *;

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE id = ?;

-- name: GetSentMessage :one
SELECT * FROM sent_messages
WHERE id = ? AND message_id = ? LIMIT 1;

-- name: GetSentMessageByID :one
SELECT * FROM sent_messages
WHERE id = ? LIMIT 1;

-- name: CreateSentMessage :one
INSERT INTO sent_messages (
  id, message_id, text, result
) VALUES (
  ?, ?, ?, ?
)
RETURNING *;

-- name: UpdateSentMessage :one
UPDATE sent_messages
set result = ?
WHERE id = ?
RETURNING *;