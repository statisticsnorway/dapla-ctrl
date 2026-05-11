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
	CASE
		WHEN @order_by::TEXT = 'kind:asc' THEN kind
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'kind:desc' THEN kind
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'short_name:asc' THEN short_name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'short_name:desc' THEN short_name
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'env:asc' THEN env
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'env:desc' THEN env
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'team:asc' THEN team_slug
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'team:desc' THEN team_slug
	END DESC,
	short_name ASC
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
	sqlc.embed(groups),
	COUNT(*) OVER () AS total_count
FROM
	groups
	JOIN shared_buckets_access_stopgap ON groups.name = shared_buckets_access_stopgap.group_name
WHERE
	shared_buckets_access_stopgap.bucket_name = @name
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'name:asc' THEN groups.name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'name:desc' THEN groups.name
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

-- name: ListUniqueUsersForBucket :many
SELECT
	users.*,
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

-- name: ListTeamsForBucket :many
SELECT
	teams.*,
	COUNT(*) OVER () AS total_count
FROM
	shared_buckets_access_stopgap
	JOIN groups ON groups.name = shared_buckets_access_stopgap.group_name
	JOIN teams ON groups.team_slug = teams.slug
WHERE
	shared_buckets_access_stopgap.bucket_name = @name
GROUP BY
	teams.slug
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
	CASE
		WHEN @order_by::TEXT = 'kind:asc' THEN kind
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'kind:desc' THEN kind
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'short_name:asc' THEN short_name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'short_name:desc' THEN short_name
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'env:asc' THEN env
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'env:desc' THEN env
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'team:asc' THEN team_slug
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'team:desc' THEN team_slug
	END DESC,
	short_name ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: ListForUser :many
SELECT
	sbs.name,
	groups.team_slug,
	ARRAY_AGG(groups.name)::TEXT[] AS groups,
	COUNT(*) OVER () AS total_count
FROM
	shared_buckets_stopgap sbs
	JOIN shared_buckets_access_stopgap sbas ON sbs.name = sbas.bucket_name
	JOIN group_members gm ON sbas.group_name = gm.group_name
	JOIN groups ON gm.group_name = groups.name
WHERE
	gm.user_id = @user_id
GROUP BY
	sbs.name,
	groups.team_slug
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'name:asc' THEN sbs.name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'name:desc' THEN sbs.name
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'kind:asc' THEN kind
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'kind:desc' THEN kind
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'short_name:asc' THEN short_name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'short_name:desc' THEN short_name
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'env:asc' THEN env
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'env:desc' THEN env
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'team:asc' THEN sbs.team_slug
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'team:desc' THEN sbs.team_slug
	END DESC,
	short_name ASC
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
	CASE
		WHEN @order_by::TEXT = 'kind:asc' THEN kind
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'kind:desc' THEN kind
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'short_name:asc' THEN short_name
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'short_name:desc' THEN short_name
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'env:asc' THEN env
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'env:desc' THEN env
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'team:asc' THEN shared_buckets_stopgap.team_slug
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'team:desc' THEN shared_buckets_stopgap.team_slug
	END DESC,
	short_name ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;
