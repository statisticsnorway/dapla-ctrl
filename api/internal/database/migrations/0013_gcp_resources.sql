-- +goose Up
CREATE TABLE gcp_team_folders (
	team_slug slug NOT NULL REFERENCES teams (slug) ON DELETE CASCADE,
	env TEXT NOT NULL,
	folder_id TEXT NOT NULL,
	UNIQUE (team_slug, env)
)
;
