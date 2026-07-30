-- +goose Up
CREATE TABLE team_repositories (
	team_slug slug NOT NULL,
	github_repository TEXT NOT NULL,
	PRIMARY KEY (team_slug, github_repo)
)
;

CREATE TABLE team_artifact_registry_repositories (
	team_slug slug NOT NULL REFERENCES teams (slug) ON DELETE CASCADE,
	format string NOT NULL,
	size_bytes BIGINT NOT NULL
)
;
