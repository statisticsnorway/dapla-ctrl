-- name: ListForTeam :many
SELECT
	sqlc.embed(team_artifact_registry_repositories),
	COUNT(*) OVER () AS total_count
FROM
	team_artifact_registry_repositories
WHERE
	team_slug = @team_slug
ORDER BY
	format ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;
