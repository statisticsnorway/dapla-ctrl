-- +goose Up
CREATE TABLE team_atlantis_config (
    team_slug slug NOT NULL REFERENCES teams (slug) ON DELETE CASCADE,
    webhook_secret TEXT
);
