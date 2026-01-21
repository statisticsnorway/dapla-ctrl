-- +goose Up
CREATE TABLE shared_buckets_stopgap (
	name TEXT NOT NULL PRIMARY KEY,
	team_slug slug NOT NULL,
	short_name TEXT NOT NULL,
	kind TEXT NOT NULL,
	env TEXT NOT NULL
)
;

CREATE TABLE shared_buckets_access_stopgap (
	bucket_name TEXT NOT NULL REFERENCES shared_buckets_stopgap (name) ON DELETE CASCADE,
	group_name TEXT NOT NULL REFERENCES groups (name) ON DELETE CASCADE,
	PRIMARY KEY (bucket_name, group_name)
)
;

CREATE TRIGGER shared_buckets_stopgap_notify
AFTER INSERT
OR
UPDATE
OR DELETE ON shared_buckets_stopgap FOR EACH ROW
EXECUTE PROCEDURE api_notify ("name", "team_slug")
;
