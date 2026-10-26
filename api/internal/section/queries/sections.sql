-- name: List :many
SELECT
	sqlc.embed(sections),
	COUNT(*) OVER () AS total_count
FROM
	sections
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'code:asc' THEN code
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'code:desc' THEN code
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'name:asc' THEN name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'name:desc' THEN name
	END DESC,
	code,
	name ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: GetByCodes :many
SELECT
	*
FROM
	sections
WHERE
	code = ANY (@codes::TEXT[])
ORDER BY
	code
;

-- name: GetByManagerId :one
SELECT
	*
FROM
	sections
WHERE
	manager_id = @user_id
ORDER BY
	code
;

-- name: ListTeamsForSection :many
SELECT
	teams.slug,
	teams.display_name,
	teams.section_code,
	teams.is_managed,
	teams.has_manual_editing,
	teams.last_successful_sync,
	teams.delete_key_confirmed_at,
	count(*) OVER () AS total_count
FROM
	teams
WHERE
	teams.section_code = @section_code
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

-- name: CountTeamsForSection :one
SELECT
	COUNT(*)
FROM
	teams
WHERE
	teams.section_code = @section_code;
