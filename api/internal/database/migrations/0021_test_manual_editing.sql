-- +goose Up
ALTER TABLE teams
ADD COLUMN has_manual_editing_test BOOLEAN NOT NULL DEFAULT FALSE
;
