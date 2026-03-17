-- name: Create :one
INSERT INTO
	teams (slug, display_name, section_code, is_managed)
VALUES
	(@slug, @display_name, @section_code, @is_managed)
RETURNING
	*
;

-- name: Update :one
UPDATE teams
SET
	display_name = COALESCE(sqlc.narg(display_name), display_name),
	section_code = COALESCE(sqlc.narg(section_code), section_code)
WHERE
	teams.slug = @slug
RETURNING
	*
;

-- name: Exists :one
SELECT
	EXISTS (
		SELECT
			slug
		FROM
			teams
		WHERE
			slug = @slug
	)
;

-- name: SlugAvailable :one
SELECT
	NOT EXISTS (
		SELECT
			slug
		FROM
			team_slugs
		WHERE
			slug = @slug
	)
;

-- name: List :many
SELECT
	sqlc.embed(teams),
	COUNT(*) OVER () AS total_count
FROM
	teams
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'slug:asc' THEN teams.slug
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'slug:desc' THEN teams.slug
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'section_code:asc' THEN teams.section_code
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'section_code:desc' THEN teams.section_code
	END DESC,
	teams.slug ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: Get :one
SELECT
	*
FROM
	teams
WHERE
	slug = @slug
;

-- name: ListBySlugs :many
SELECT
	*
FROM
	teams
WHERE
	slug = ANY (@slugs::slug[])
ORDER BY
	slug ASC
;

-- name: CreateDeleteKey :one
INSERT INTO
	team_delete_keys (team_slug, created_by)
VALUES
	(@team_slug, @created_by)
RETURNING
	*
;

-- name: GetDeleteKey :one
SELECT
	*
FROM
	team_delete_keys
WHERE
	key = @key
	AND team_slug = @slug::slug
;

-- name: ConfirmDeleteKey :exec
UPDATE team_delete_keys
SET
	confirmed_at = NOW()
WHERE
	key = @key
;

-- name: SetDeleteKeyConfirmedAt :exec
UPDATE teams
SET
	delete_key_confirmed_at = NOW()
WHERE
	slug = @slug
;

-- name: ListAllSlugs :one
SELECT
	ARRAY_AGG(slug)::slug[] AS slugs
FROM
	teams
;

-- name: ListAllForSearch :many
SELECT
	slug,
	section_code
FROM
	teams
ORDER BY
	slug ASC
;

-- name: GetAccessManagers :many
SELECT
	user_id
FROM
	user_roles
WHERE
	user_roles.target_team_slug = @team_slug::slug
	AND user_roles.role_name = 'Tilgangsansvarlig'
ORDER BY
	user_id
;

-- name: AddAccessManager :exec
INSERT INTO
	user_roles (user_id, role_name, target_team_slug)
VALUES
	(@user_id, 'Tilgangsansvarlig', @team_slug::slug)
ON CONFLICT DO NOTHING
;

-- name: RemoveAccessManager :exec
DELETE FROM user_roles
WHERE
	user_id = @user_id
	AND role_name = 'Tilgangsansvarlig'
	AND target_team_slug = @team_slug::slug
;
