-- name: ListMembers :many
SELECT
	sqlc.embed(users),
	sqlc.embed(groups),
	COUNT(*) OVER () AS total_count
FROM
	group_members
	JOIN groups ON groups.name = group_members.group_name
	JOIN users ON users.id = group_members.user_id
WHERE
	group_members.group_name = @group_name
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'name:asc' THEN LOWER(users.name)
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'name:desc' THEN LOWER(users.name)
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'email:asc' THEN users.email
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'email:desc' THEN users.email
	END DESC,
	users.name,
	users.email ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: CountMembers :one
SELECT
	COUNT(*) AS total
FROM
	group_members
WHERE
	group_name = @group_name
;

-- name: Get :one
SELECT
	name,
	external_id,
	team_slug
FROM
	groups
WHERE
	name = @name
;

-- name: UpdateExternalId :exec
UPDATE groups
SET
	external_id = @external_id
WHERE
	name = @name
;
