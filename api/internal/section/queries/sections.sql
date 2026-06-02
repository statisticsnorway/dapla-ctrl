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
