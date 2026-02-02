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
	CASE
		WHEN @order_by::TEXT = 'section_code:asc' THEN users.section_code
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'section_code:desc' THEN users.section_code
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
	group_name,
	user_id,
	COUNT(*) OVER () AS total_count
FROM
	group_members
	JOIN groups ON group_members.group_name = groups.name
WHERE
	group_members.user_id = @user_id
	AND (
		sqlc.narg('filter')::TEXT[] IS NULL
		OR (category) = ANY (sqlc.narg('filter')::TEXT[])
	)
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'name:asc' THEN name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'name:desc' THEN name
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'team:asc' THEN team_slug
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'team:desc' THEN team_slug
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'category:asc' THEN category
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'category:desc' THEN category
	END DESC,
	name ASC
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

-- name: RefreshTeamMembers :exec
REFRESH MATERIALIZED VIEW team_members
;
