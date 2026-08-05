-- +goose Up
CREATE TABLE team_features (
	team_slug slug NOT NULL REFERENCES teams (slug) ON DELETE CASCADE,
	name TEXT NOT NULL,
	env TEXT NOT NULL,
	UNIQUE (team_slug, name, env)
)
;
