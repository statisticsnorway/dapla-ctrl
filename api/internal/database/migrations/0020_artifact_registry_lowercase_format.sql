-- +goose Up
UPDATE team_artifact_registry_repositories
SET
	format = LOWER(format)
;

ALTER TABLE team_artifact_registry_repositories
ADD CHECK (format = LOWER(format))
;

UPDATE activity_log_entries
SET
	resource_name = LOWER(resource_name)
WHERE
	resource_type = 'ARTIFACT_REGISTRY_REPOSITORY'
;
