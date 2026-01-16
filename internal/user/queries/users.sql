-- name: List :many
SELECT
	sqlc.embed(users),
	COUNT(*) OVER () AS total_count
FROM
	users
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'name:asc' THEN name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'name:desc' THEN name
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'email:asc' THEN email
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'email:desc' THEN email
	END DESC,
	name,
	email ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: GetByIDs :many
SELECT
	*
FROM
	users
WHERE
	id = ANY (@ids::UUID[])
ORDER BY
	name,
	email ASC
;

-- name: GetByEmail :one
SELECT
	*
FROM
	users
WHERE
	email = LOWER(@email)
;

-- name: GetByExternalId :one
SELECT
	*
FROM
	users
WHERE
	external_id = LOWER(@external_id)
;

-- name: ListTeamMembersForUser :many
SELECT
	sqlc.embed(users),
	COUNT(*) OVER () AS total_count
FROM
	(
		SELECT
			user_id
		FROM
			(
				SELECT
					team_slug AS slug
				FROM
					team_members
				WHERE
					team_members.user_id = @user_id
			) t
			JOIN team_members ON team_members.team_slug = t.slug
		WHERE
			team_members.user_id != @user_id
		GROUP BY
			team_members.user_id
	) t
	JOIN users ON users.id = t.user_id
GROUP BY
	users.id
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'name:asc' THEN users.name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'name:desc' THEN users.name
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
