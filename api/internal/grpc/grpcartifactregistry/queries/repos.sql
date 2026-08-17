-- name: Get :one
SELECT
	sqlc.embed(ar)
FROM
	team_artifact_registry_repositories ar
WHERE
	team_slug = @team_slug::slug
	AND format = @format
;

-- name: GetGithubRepositoriesForTeam :many
SELECT
	repository_name
FROM
	team_artifact_registry_gh_repos_allow_list
WHERE
	team_slug = @team_slug::slug
ORDER BY
	repository_name ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: List :many
SELECT
	sqlc.embed(ar)
FROM
	team_artifact_registry_repositories ar
WHERE
	ar.team_slug = @team_slug::slug
ORDER BY
	format ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: CountTeamRepos :one
SELECT
	COUNT(*) AS total
FROM
	team_artifact_registry_repositories
WHERE
	team_slug = @team_slug::slug
;

-- name: SetSizeBytes :exec
UPDATE team_artifact_registry_repositories
SET
	size_bytes = @size_bytes
WHERE
	team_slug = @team_slug::slug
	AND format = @format
;
