-- name: Get :one
SELECT
    sqlc.embed(team_artifact_registry_repositories)
FROM
    team_artifact_registry_repositories
WHERE
    team_slug = @team_slug::slug AND
    format = @format
;

-- name: List :many
SELECT
    sqlc.embed(team_artifact_registry_repositories)
FROM
    team_artifact_registry_repositories
WHERE
    team_slug = @team_slug::slug
ORDER BY
    format ASC
LIMIT
    sqlc.arg('limit')
OFFSET
    sqlc.arg('offset')
;

-- name: CountTeamRepos :one
SELECT
    COUNT(*) as total
FROM
    team_artifact_registry_repositories
WHERE
   team_slug = @team_slug::slug
;
