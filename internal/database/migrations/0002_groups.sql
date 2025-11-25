-- +goose Up
CREATE TABLE groups (
	name TEXT PRIMARY KEY,
	team_slug slug NOT NULL,
	category TEXT NOT NULL,
	suffix TEXT NOT NULL,
	external_id TEXT,
	UNIQUE (team_slug, category, suffix)
)
;

CREATE TABLE group_members (
	group_name TEXT NOT NULL REFERENCES groups (name) ON DELETE CASCADE,
	user_id UUID NOT NULL,
	PRIMARY KEY (group_name, user_id)
)
;

CREATE INDEX group_members_for_user ON group_members (user_id)
;

CREATE INDEX group_members_for_group ON group_members (group_name)
;

-- foreign keys
ALTER TABLE groups
ADD FOREIGN KEY (team_slug) REFERENCES teams (slug) ON DELETE CASCADE
;

ALTER TABLE group_members
ADD FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
;

-- authorizations
INSERT INTO
	authorizations (name, description)
VALUES
	(
		'teams:groups:create',
		'Permission to create groups.'
	),
	(
		'teams:groups:members:admin',
		'Permission to manage group members'
	)
;
