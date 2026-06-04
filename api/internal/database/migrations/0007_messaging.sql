-- +goose Up
CREATE TABLE messages (
	id UUID DEFAULT GEN_RANDOM_UUID() PRIMARY KEY,
	actor TEXT NOT NULL,
	recipient UUID NOT NULL REFERENCES users (id),
	subject TEXT NOT NULL,
	message TEXT NOT NULL,
	status TEXT NOT NULL CHECK (
		status ~ '^PENDING|PUBLISHED|SUCCESSFUL|FAILED$'::TEXT
	),
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP()
)
;

INSERT INTO
	roles (name, description)
VALUES
	('Message sender', 'Can send messages')
;

INSERT INTO
	authorizations (name, description)
VALUES
	('messages:send', 'Permission to send messages')
;

INSERT INTO
	role_authorizations (role_name, authorization_name)
VALUES
	('Message sender', 'messages:send')
;
