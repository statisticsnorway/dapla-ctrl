-- name: UpsertTeamFolder :exec
INSERT INTO
	gcp_team_folders (team_slug, env, folder_id)
VALUES
	($1, $2, $3)
ON CONFLICT (team_slug, env) DO UPDATE
SET
	folder_id = EXCLUDED.folder_id
;

-- name: GetTeamFolder :one
SELECT
	team_slug,
	env,
	folder_id
FROM
	gcp_team_folders
WHERE
	team_slug = $1
	AND env = $2
;

-- name: DeleteTeamFolders :exec
DELETE FROM gcp_team_folders
WHERE
	team_slug = $1
;
