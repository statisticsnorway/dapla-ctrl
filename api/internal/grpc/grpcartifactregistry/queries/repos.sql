-- name: Get :one
SELECT
	sqlc.embed(ar)
FROM
	team_artifact_registry_repositories ar
WHERE
	team_slug = @team_slug::slug
	AND format = @format
;

-- name: GetGithubRepositoriesForTeam :one
SELECT
	team_slug,
	ARRAY_AGG(gr.github_repository)::TEXT[] AS github_repos
FROM
	team_artifact_registry_github_repositories gr
WHERE
	team_slug = @team_slug::slug
GROUP BY
	team_slug
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
	JOIN team_artifact_registry_github_repositories gr ON ar.team_slug = gr.team_slug
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
