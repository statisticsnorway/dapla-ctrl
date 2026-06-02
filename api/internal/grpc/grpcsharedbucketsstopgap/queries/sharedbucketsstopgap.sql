-- name: Create :exec
INSERT INTO
	shared_buckets_stopgap (name, team_slug, short_name, kind, env)
VALUES
	(@name, @team_slug, @short_name, @kind, @env)
;

-- name: Get :one
SELECT
	sqlc.embed(shared_buckets_stopgap)
FROM
	shared_buckets_stopgap
WHERE
	team_slug = @team_slug
	AND short_name = @short_name
	AND kind = @kind
	AND env = @env
;

-- name: ListGroups :many
SELECT
	sqlc.embed(groups)
FROM
	groups
	JOIN shared_buckets_access_stopgap ON groups.name = shared_buckets_access_stopgap.group_name
WHERE
	shared_buckets_access_stopgap.bucket_name = @name
ORDER BY
	bucket_name
;

-- name: AddGroup :exec
INSERT INTO
	shared_buckets_access_stopgap (bucket_name, group_name)
VALUES
	(@bucket_name, @group_name)
;

-- name: RemoveGroup :exec
DELETE FROM shared_buckets_access_stopgap
WHERE
	bucket_name = @bucket_name
	AND group_name = @group_name
;
