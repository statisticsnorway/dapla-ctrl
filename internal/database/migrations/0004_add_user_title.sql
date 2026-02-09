-- +goose Up
ALTER TABLE users
ADD COLUMN IF NOT EXISTS job_title TEXT
;

-- +goose Down
ALTER TABLE users
DROP COLUMN job_title
;
