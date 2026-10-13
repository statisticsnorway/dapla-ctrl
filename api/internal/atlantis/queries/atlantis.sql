-- name: Create :exec
INSERT INTO
	team_atlantis_config (team_slug)
VALUES
	(team_slug = @team_slug::slug)
;
