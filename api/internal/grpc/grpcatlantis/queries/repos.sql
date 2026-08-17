-- name: Set :one
UPDATE team_atlantis_config
SET
    webhook_secret = @webhook_secret
WHERE
    team_slug = @team_slug::slug
RETURNING
    *
;

-- name: Get :one
SELECT
    sqlc.embed(team_atlantis_config)
FROM team_atlantis_config
WHERE
    team_slug = @team_slug::slug
;
