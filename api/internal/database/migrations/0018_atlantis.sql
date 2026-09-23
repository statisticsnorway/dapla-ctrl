-- +goose Up
CREATE TABLE team_atlantis_config (
	team_slug slug PRIMARY KEY REFERENCES teams (slug) ON DELETE CASCADE,
	custom_name TEXT,
	webhook_secret TEXT,
	custom_image TEXT,
	resources JSON,
	disk_size TEXT,
	repo_config JSON,
	UNIQUE(custom_name)
)
;
