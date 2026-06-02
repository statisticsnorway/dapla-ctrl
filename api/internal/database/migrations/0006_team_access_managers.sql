-- +goose Up
INSERT INTO
	roles (name, description)
VALUES
	(
		'Tilgangsansvarlig',
		'Kan administrere medlemskap av teamets grupper'
	)
;

INSERT INTO
	role_authorizations (role_name, authorization_name)
VALUES
	('Tilgangsansvarlig', 'teams:members:admin')
;
