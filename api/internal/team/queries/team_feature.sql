-- name: EnableTeamFeature :exec
INSERT INTO
	team_features (team_slug, name, env)
VALUES
	(@team_slug, @name, @env)
;

-- name: DisableTeamFeature :exec
DELETE FROM team_features
WHERE
	team_slug = @team_slug
	AND name = @name
	AND env = @env
;

-- name: GetFeaturesForTeam :many
SELECT
	sqlc.embed(team_features)
FROM
	team_features
WHERE
	team_slug = @team_slug
ORDER BY
	name ASC
;
