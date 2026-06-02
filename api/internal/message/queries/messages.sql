-- name: Create :one
INSERT INTO
	messages (actor, recipient, status, subject, message)
VALUES
	(@actor, @recipient, 'PENDING', @subject, @message)
RETURNING
	*
;

-- name: UpdateStatus :one
UPDATE messages
SET
	status = @status
WHERE
	messages.id = @id
RETURNING
	*
;

-- name: GetByIDs :many
SELECT
	*
FROM
	messages
WHERE
	id = ANY (@ids::UUID[])
ORDER BY
	created_at DESC
;

-- name: GetByStatus :many
SELECT
	*
FROM
	messages
WHERE
	status = @status
ORDER BY
	created_at DESC
;

-- name: GetUserByID :one
SELECT
	*
FROM
	users
WHERE
	id = @id::UUID
ORDER BY
	name,
	email ASC
;

-- name: GetUserByEmail :one
SELECT
	*
FROM
	users
WHERE
	email = LOWER(@email)
;

-- name: UserExists :one
SELECT
	EXISTS (
		SELECT
			email
		FROM
			users
		WHERE
			email = @email
	)
;

-- name: List :many
SELECT
	sqlc.embed(messages),
	COUNT(*) OVER () AS total_count
FROM
	messages
WHERE
	(
		sqlc.narg(status)::TEXT IS NULL
		OR status = sqlc.narg(status)::TEXT
	)
	AND (
		sqlc.narg(actor)::TEXT IS NULL
		OR actor = sqlc.narg(actor)::TEXT
	)
	AND (
		sqlc.narg(recipient)::UUID IS NULL
		OR recipient = sqlc.narg(recipient)::UUID
	)
ORDER BY
	messages.created_at ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;
