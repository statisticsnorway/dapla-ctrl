-- name: List :many
SELECT
	*,
	COUNT(*) OVER () AS total_count
FROM
	(
		SELECT
			FORMAT('ssb-%s-data-%s-%s', t.slug, tbs.kind, tbs.env) AS name,
			t.slug AS team_slug,
			tbs.kind,
			tbs.env
		FROM
			teams t
			CROSS JOIN team_bucket_specs tbs
		WHERE
			t.is_managed = TRUE
	) X
WHERE
	(
		sqlc.narg('kinds')::TEXT[] IS NULL
		OR (kind) = ANY (sqlc.narg('kinds')::TEXT[])
	)
	AND (
		sqlc.narg('envs')::TEXT[] IS NULL
		OR (env) = ANY (sqlc.narg('envs')::TEXT[])
	)
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
	team_slug ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: ListForTeam :many
SELECT
	*,
	COUNT(*) OVER () AS total_count
FROM
	(
		SELECT
			FORMAT('ssb-%s-data-%s-%s', t.slug, tbs.kind, tbs.env) AS name,
			t.slug AS team_slug,
			tbs.kind,
			tbs.env
		FROM
			teams t
			CROSS JOIN team_bucket_specs tbs
		WHERE
			t.is_managed = TRUE
	) X
WHERE
	team_slug = @team_slug::slug
	AND (
		sqlc.narg('kinds')::TEXT[] IS NULL
		OR (kind) = ANY (sqlc.narg('kinds')::TEXT[])
	)
	AND (
		sqlc.narg('envs')::TEXT[] IS NULL
		OR (env) = ANY (sqlc.narg('envs')::TEXT[])
	)
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
	team_slug ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: GetByName :one
SELECT
	*
FROM
	(
		SELECT
			FORMAT('ssb-%s-data-%s-%s', t.slug, tbs.kind, tbs.env) AS name,
			t.slug AS team_slug,
			tbs.kind,
			tbs.env
		FROM
			teams t
			CROSS JOIN team_bucket_specs tbs
		WHERE
			t.is_managed = TRUE
	) X
WHERE
	name = @name::TEXT
;

-- name: GetByNames :many
SELECT
	*
FROM
	(
		SELECT
			FORMAT('ssb-%s-data-%s-%s', t.slug, tbs.kind, tbs.env) AS name,
			t.slug AS team_slug,
			tbs.kind,
			tbs.env
		FROM
			teams t
			CROSS JOIN team_bucket_specs tbs
		WHERE
			t.is_managed = TRUE
	) X
WHERE
	name = ANY (@names::TEXT[])
ORDER BY
	name
;
