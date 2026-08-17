-- +goose Up
CREATE TABLE team_artifact_registry_repositories (
	team_slug slug NOT NULL REFERENCES teams (slug) ON DELETE CASCADE,
	format TEXT NOT NULL,
	size_bytes BIGINT NOT NULL,
	UNIQUE (team_slug, format)
)
;

CREATE TABLE team_artifact_registry_gh_repos_allow_list (
	team_slug slug NOT NULL REFERENCES teams (slug) ON DELETE CASCADE,
	repository_name TEXT NOT NULL,
	UNIQUE (team_slug, repository_name)
)
;
