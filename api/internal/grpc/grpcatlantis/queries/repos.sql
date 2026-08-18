-- name: UpsertWebhookSecret :exec
INSERT INTO
    team_atlantis_config (team_slug, webhook_secret)
VALUES
    (@team_slug, @webhook_secret)
ON CONFLICT (team_slug) DO UPDATE
SET
    webhook_secret = EXCLUDED.webhook_secret
;

-- name: Get :one
SELECT
    sqlc.embed(team_atlantis_config)
FROM team_atlantis_config
WHERE
    team_slug = @team_slug::slug
;
