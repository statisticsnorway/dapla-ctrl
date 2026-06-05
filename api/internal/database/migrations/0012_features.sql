-- +goose Up
ALTER TABLE teams
ADD COLUMN has_manual_editing BOOLEAN NOT NULL DEFAULT FALSE
;
