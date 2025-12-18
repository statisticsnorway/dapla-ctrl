-- +goose Up
ALTER TABLE teams
ADD COLUMN display_name TEXT NOT NULL DEFAULT '0'
;

UPDATE teams
SET
	display_name = slug
;

ALTER TABLE teams
ALTER COLUMN display_name
DROP DEFAULT
;
