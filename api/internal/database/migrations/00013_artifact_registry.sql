-- +goose Up
CREATE TABLE team_repositories (
    team_slug slug NOT NULL,
    github_repository TEXT NOT NULL,
    PRIMARY KEY (team_slug, github_repo)
)
;
