-- +goose Up
ALTER TABLE users
ADD COLUMN section_code TEXT REFERENCES sections (code) ON DELETE SET NULL
;
