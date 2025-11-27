-- name: SetLastSuccessfulSync :exec
UPDATE teams
SET
	last_successful_sync = NOW()
WHERE
	teams.slug = @slug
;

-- name: Delete :exec
DELETE FROM teams
WHERE
	slug = @slug
	AND delete_key_confirmed_at IS NOT NULL
;

-- name: Get :one
SELECT
	*
FROM
	teams
WHERE
	slug = @slug
;

-- name: List :many
SELECT
	*
FROM
	teams
ORDER BY
	slug ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: Count :one
SELECT
	COUNT(*) AS total
FROM
	teams
;

-- name: ListMembers :many
SELECT
	users.*
FROM
	user_roles
	JOIN teams ON teams.slug = user_roles.target_team_slug
	JOIN users ON users.id = user_roles.user_id
WHERE
	user_roles.target_team_slug = @team_slug::slug
ORDER BY
	users.name ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: CountMembers :one
SELECT
	COUNT(user_roles.*) AS total
FROM
	user_roles
	JOIN teams ON teams.slug = user_roles.target_team_slug
WHERE
	user_roles.target_team_slug = @team_slug::slug
;

-- name: ListGroups :many
SELECT
	name,
	team_slug,
	external_id
FROM
	groups
WHERE
	team_slug = @team_slug::slug
ORDER BY
	name ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: CountGroups :one
SELECT
	COUNT(*) AS total
FROM
	groups
WHERE
	team_slug = @team_slug::slug
;
