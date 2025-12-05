-- name: Create :one
INSERT INTO
	groups (name, team_slug, category, suffix, external_id)
VALUES
	(
		@name,
		@team_slug,
		@category,
		@suffix,
		@external_id
	)
RETURNING
	*
;

-- name: GroupExists :one
SELECT
	EXISTS (
		SELECT
			team_slug,
			category,
			suffix
		FROM
			groups
		WHERE
			team_slug = @team_slug
			AND category = @category
			AND suffix = @suffix
	)
;

-- name: TeamExists :one
SELECT
	EXISTS (
		SELECT
			slug
		FROM
			teams
		WHERE
			slug = @slug
			AND delete_key_confirmed_at IS NULL
	)
;

-- name: List :many
SELECT
	sqlc.embed(groups),
	COUNT(*) OVER () AS total_count
FROM
	groups
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'slug:asc' THEN team_slug
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'slug:desc' THEN team_slug
	END DESC,
	team_slug ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: Get :one
SELECT
	*
FROM
	groups
WHERE
	name = @name
;

-- name: GetByNames :many
SELECT
	*
FROM
	groups
WHERE
	name = ANY (@ids::TEXT[])
ORDER BY
	team_slug,
	category,
	suffix
;

-- name: ListByTeamSlug :many
SELECT
	*
FROM
	groups
WHERE
	team_slug = @team_slug::slug
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'slug:asc' THEN team_slug
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'slug:desc' THEN team_slug
	END DESC,
	team_slug ASC
;

-- name: ListAllForSearch :many
SELECT
    name,
    team_slug
FROM
    groups
ORDER BY
    name ASC
;
