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
