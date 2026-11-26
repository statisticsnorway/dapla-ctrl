-- name: ListForTeam :many
SELECT
	team_slug,
	env,
	folder_id,
	COUNT(*) OVER () AS total_count
FROM
	gcp_team_folders
WHERE
	team_slug = @team_slug::slug
	AND (
		sqlc.narg('envs')::TEXT[] IS NULL
		OR (env) = ANY (sqlc.narg('envs')::TEXT[])
	)
ORDER BY
	env ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: Get :one
SELECT
	*
FROM
	gcp_team_folders
WHERE
	team_slug = @team_slug::slug
	AND env = @env::TEXT
;
