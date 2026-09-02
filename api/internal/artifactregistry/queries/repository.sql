-- name: ListGithubReposForTeam :many
SELECT
	sqlc.embed(team_artifact_registry_gh_repos_allow_list),
	COUNT(*) OVER () AS total_count
FROM
	team_artifact_registry_gh_repos_allow_list
WHERE
	team_slug = @team_slug
ORDER BY
	repository_name ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: AddGithubRepositoryToTeam :one
INSERT INTO
	team_artifact_registry_gh_repos_allow_list (team_slug, repository_name)
VALUES
	(@team_slug, @repository_name)
RETURNING
	*
;

-- name: RemoveGithubRepositoryFromTeam :exec
DELETE FROM team_artifact_registry_gh_repos_allow_list
WHERE
	team_slug = @team_slug
	AND repository_name = @repository_name
;

-- name: CreateArtifactRegistryRepository :exec
INSERT INTO
    team_artifact_registry_repositories (team_slug, format, size_bytes)
VALUES
    (@team_slug, @format, 0)
;
