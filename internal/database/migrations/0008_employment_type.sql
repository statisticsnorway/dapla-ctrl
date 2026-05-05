-- +goose Up
ALTER TABLE users
ADD COLUMN IF NOT EXISTS employment_type TEXT NOT NULL DEFAULT 'Ukjent'
;
