-- name: Create :one
INSERT INTO
	teams (slug, purpose)
VALUES
	(@slug, @purpose)
RETURNING
	*
;

-- name: Update :one
UPDATE teams
SET
	purpose = COALESCE(sqlc.narg(purpose), purpose)
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
		WHEN @order_by::TEXT = 'slug:asc' THEN slug
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'slug:desc' THEN slug
	END DESC,
	slug ASC
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
	purpose
FROM
	teams
ORDER BY
	slug ASC
;
