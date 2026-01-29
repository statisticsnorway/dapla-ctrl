-- name: List :many
SELECT
	sqlc.embed(shared_buckets_stopgap),
	COUNT(*) OVER () AS total_count
FROM
	shared_buckets_stopgap
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'name:asc' THEN name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'name:desc' THEN name
	END DESC,
	name ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: ListAllForSearch :many
SELECT
	name,
	team_slug
FROM
	shared_buckets_stopgap
ORDER BY
	name ASC
;

-- name: GetByNames :many
SELECT
	*
FROM
	shared_buckets_stopgap
WHERE
	name = ANY (@names::TEXT[])
ORDER BY
	name
;

-- name: ListGroupsForBucket :many
SELECT
	group_name,
	COUNT(*) OVER () AS total_count
FROM
	shared_buckets_access_stopgap
WHERE
	shared_buckets_access_stopgap.bucket_name = @name
ORDER BY
	group_name ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: ListUsersForBucket :many
SELECT
	groups.team_slug,
	users.id AS user_id,
	COUNT(*) OVER () AS total_count
FROM
	shared_buckets_access_stopgap
	JOIN group_members ON group_members.group_name = shared_buckets_access_stopgap.group_name
	JOIN users ON group_members.user_id = users.id
	JOIN groups ON groups.name = group_members.group_name
WHERE
	shared_buckets_access_stopgap.bucket_name = @name
GROUP BY
	groups.team_slug,
	users.id
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

-- name: ListUniqueUsersForBucket :many
SELECT
	users.id,
	COUNT(*) OVER () AS total_count
FROM
	shared_buckets_access_stopgap
	JOIN group_members ON group_members.group_name = shared_buckets_access_stopgap.group_name
	JOIN users ON group_members.user_id = users.id
WHERE
	shared_buckets_access_stopgap.bucket_name = @name
GROUP BY
	users.id
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

-- name: ListTeamsForBucket :many
SELECT
	team_slug,
	COUNT(*) OVER () AS total_count
FROM
	shared_buckets_access_stopgap
	JOIN groups ON groups.name = shared_buckets_access_stopgap.group_name
WHERE
	shared_buckets_access_stopgap.bucket_name = @name
GROUP BY
	team_slug
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

-- name: ListForTeam :many
SELECT
	sqlc.embed(shared_buckets_stopgap),
	COUNT(*) OVER () AS total_count
FROM
	shared_buckets_stopgap
WHERE
	team_slug = @team_slug
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'name:asc' THEN name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'name:desc' THEN name
	END DESC,
	name ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: ListForUser :many
SELECT
	sqlc.embed(shared_buckets_stopgap),
	COUNT(*) OVER () AS total_count
FROM
	shared_buckets_stopgap
	JOIN shared_buckets_access_stopgap ON shared_buckets_access_stopgap.bucket_name = shared_buckets_stopgap.name
	JOIN group_members ON shared_buckets_access_stopgap.group_name = group_members.group_name
WHERE
	group_members.user_id = @user_id
GROUP BY
	name
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'name:asc' THEN name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'name:desc' THEN name
	END DESC,
	name ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: ListAccessToForTeam :many
SELECT
	sqlc.embed(shared_buckets_stopgap),
	COUNT(*) OVER () AS total_count
FROM
	shared_buckets_stopgap
	JOIN shared_buckets_access_stopgap ON shared_buckets_access_stopgap.bucket_name = shared_buckets_stopgap.name
	JOIN groups ON shared_buckets_access_stopgap.group_name = groups.name
WHERE
	groups.team_slug = @team_slug
GROUP BY
	shared_buckets_stopgap.name,
	groups.team_slug
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'name:asc' THEN shared_buckets_stopgap.name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'name:desc' THEN shared_buckets_stopgap.name
	END DESC,
	shared_buckets_stopgap.name ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;
