-- +goose Up
CREATE TABLE team_artifact_registry_repositories (
	team_slug slug NOT NULL REFERENCES teams (slug) ON DELETE CASCADE,
	format string NOT NULL,
	size_bytes BIGINT NOT NULL,
	UNIQUE (team_slug, format)
)
;

CREATE TABLE team_artifact_registry_github_repositories (
	team_slug slug NOT NULL REFERENCES teams (slug) ON DELETE CASCADE,
	github_repository TEXT NOT NULL,
	UNIQUE (team_slug, github_repo)
)
;
