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

-- name: ListForUser :many
SELECT
	sqlc.embed(users),
	sqlc.embed(groups),
	COUNT(*) OVER () AS total_count
FROM
	group_members
	JOIN groups ON groups.name = group_members.group_name
	JOIN users ON users.id = group_members.user_id
WHERE
	group_members.user_id = @user_id
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'slug:asc' THEN groups.team_slug
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'slug:desc' THEN groups.team_slug
	END DESC,
	groups.team_slug ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: ListForTeamMember :many
SELECT
	sqlc.embed(groups),
	COUNT(*) OVER () AS total_count
FROM
	group_members
	JOIN groups ON groups.name = group_members.group_name
	JOIN users ON users.id = group_members.user_id
WHERE
	group_members.user_id = @user_id
	AND groups.team_slug = @team_slug
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'name:asc' THEN groups.name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'name:desc' THEN groups.name
	END DESC,
	groups.name ASC
;

-- name: GetMember :one
SELECT
	users.*
FROM
	group_members
	JOIN users ON users.id = group_members.user_id
WHERE
	group_members.group_name = @group_name
	AND group_members.user_id = @user_id
;

-- name: GetMemberByEmail :one
SELECT
	users.*
FROM
	group_members
	JOIN users ON users.id = group_members.user_id
WHERE
	group_members.group_name = @group_name
	AND users.email = @email
;

-- name: AddMember :exec
INSERT INTO
	group_members (group_name, user_id)
VALUES
	(@group_name, @user_id)
ON CONFLICT DO NOTHING
;

-- name: RemoveMember :exec
DELETE FROM group_members
WHERE
	user_id = @user_id
	AND group_name = @group_name
;
