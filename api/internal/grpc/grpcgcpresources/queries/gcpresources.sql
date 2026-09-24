-- name: UpsertTeamFolder :exec
INSERT INTO
	gcp_team_folders (team_slug, env, folder_id)
VALUES
	(@team_slug, @env, @folder_id)
ON CONFLICT (team_slug, env) DO UPDATE
SET
	folder_id = EXCLUDED.folder_id
;

-- name: GetTeamFolder :one
SELECT
	sqlc.embed(gcp_team_folders)
FROM
	gcp_team_folders
WHERE
	team_slug = @team_slug
	AND env = @env
;

-- name: ListTeamFolders :many
SELECT
	sqlc.embed(gcp_team_folders)
FROM
	gcp_team_folders
WHERE
	team_slug = @team_slug
ORDER BY
	env
;
