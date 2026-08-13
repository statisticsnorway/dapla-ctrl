-- name: ListGithubReposForTeam :many
SELECT
	sqlc.embed(team_artifact_registry_github_repositories),
	COUNT(*) OVER () AS total_count
FROM
	team_artifact_registry_github_repositories
WHERE
	team_slug = @team_slug
ORDER BY
	github_repository ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: AddGithubRepositoryToTeam :one
INSERT INTO
	team_artifact_registry_github_repositories (team_slug, github_repository)
VALUES
	(@team_slug, @github_repository)
RETURNING
	*
;

-- name: RemoveGithubRepositoryFromTeam :exec
DELETE FROM team_artifact_registry_github_repositories
WHERE
	team_slug = @team_slug
	AND github_repository = @github_repository
;
