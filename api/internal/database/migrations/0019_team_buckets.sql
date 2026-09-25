-- +goose Up
CREATE TABLE team_bucket_specs (
	env TEXT NOT NULL,
	kind TEXT NOT NULL,
	UNIQUE (env, kind)
)
;

INSERT INTO
	team_bucket_specs (env, kind)
VALUES
	('test', 'produkt'),
	('test', 'kilde'),
	('prod', 'produkt'),
	('prod', 'kilde')
;
