-- +goose Up
-- Grant permissions in GCP if the role cloudsqlsuperuser exists
-- +goose StatementBegin
DO $$
BEGIN
   IF EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'cloudsqlsuperuser') THEN
        GRANT ALL ON SCHEMA public TO cloudsqlsuperuser;
   END IF;
END
$$
;

-- +goose StatementEnd
-- extensions
CREATE EXTENSION fuzzystrmatch
;

-- functions
-- +goose StatementBegin
CREATE FUNCTION set_updated_at () RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = NOW(); RETURN NEW; END; $$ LANGUAGE plpgsql
;

CREATE FUNCTION unique_team_slug () RETURNS trigger AS $unique_team_slug$
    BEGIN
        IF (SELECT slug from team_slugs WHERE slug = NEW.slug) IS NOT NULL THEN
            RAISE 'Team slug is not available: %', NEW.slug
            USING ERRCODE = 'unique_violation';
        END IF;
        RETURN NEW;
    END;
$unique_team_slug$ LANGUAGE plpgsql
;

CREATE FUNCTION register_team_slug () RETURNS trigger AS $register_team_slug$
    BEGIN
        INSERT INTO team_slugs (slug) VALUES (NEW.slug);
        RETURN NEW;
    END;
$register_team_slug$ LANGUAGE plpgsql
;

CREATE FUNCTION api_notify () RETURNS trigger AS $$
BEGIN
  -- We accept a number of keys as arguments, and will read the values using NEW if it is set, or OLD if it is not.
  -- We will then send a notification to api_notifiy with a JSON object containing the keys and values, as well as
  -- the table name and operation.
  DECLARE
    values text[];
    i integer := 0;
    key text;
  BEGIN
    IF TG_NARGS > 0 AND TG_OP IN ('CREATE', 'UPDATE', 'DELETE') THEN
      FOREACH key IN ARRAY TG_ARGV LOOP
        IF TG_OP != 'DELETE' THEN
          values := array_append(values, row_to_json(NEW)->>key);
        ELSE
          values := array_append(values, row_to_json(OLD)->>key);
        END IF;
        i := i + 1;
      END LOOP;
    END IF;

    -- Construct the JSON object and send the notification. The JSON object will be of the form:
    -- {
    --   "table": "table_name",
    --   "op": "operation",
    --   "data": {
    --     "key1": "value1",
    --     "key2": "value2",
    --     ...
    --   }
    -- }
    PERFORM pg_notify('api_notify', jsonb_build_object('table', TG_TABLE_NAME, 'op', TG_OP, 'data', jsonb_object(TG_ARGV, values))::text);
    RETURN NULL;
  END;
RETURN NULL;
END;
$$ LANGUAGE plpgsql
;
-- +goose StatementEnd

-- types
CREATE DOMAIN slug AS TEXT CHECK (value ~ '^[a-z][a-z0-9-]{0,15}[a-z]$'::TEXT)
;

CREATE TYPE repository_authorization_enum AS ENUM('deploy')
;

CREATE TYPE usersync_log_entry_action AS ENUM(
       'create_user',
       'update_user',
       'delete_user',
       'assign_role',
       'revoke_role'
)
;

-- tables
CREATE TABLE authorizations (
    name TEXT PRIMARY KEY,
    description TEXT NOT NULL
)
;

CREATE TABLE activity_log_entries (
	id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
	actor TEXT NOT NULL,
	action TEXT NOT NULL,
	resource_type TEXT NOT NULL,
	resource_name TEXT NOT NULL,
	team_slug slug,
	data bytea
)
;

CREATE TABLE reconciler_errors (
	id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
	correlation_id UUID NOT NULL,
	reconciler TEXT NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
	error_message TEXT NOT NULL,
	team_slug slug NOT NULL,
	UNIQUE (team_slug, reconciler)
)
;

CREATE TABLE reconciler_config (
	reconciler TEXT NOT NULL,
	key TEXT NOT NULL,
	display_name TEXT NOT NULL,
	description TEXT NOT NULL,
	value TEXT,
	secret BOOLEAN DEFAULT TRUE NOT NULL,
	PRIMARY KEY (reconciler, key)
)
;

CREATE TABLE reconciler_states (
	id UUID DEFAULT gen_random_uuid () NOT NULL PRIMARY KEY,
	reconciler_name TEXT NOT NULL,
	team_slug slug NOT NULL,
	value bytea NOT NULL,
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE (reconciler_name, team_slug)
)
;

CREATE TABLE reconcilers (
	name TEXT PRIMARY KEY,
	display_name TEXT NOT NULL UNIQUE,
	description TEXT NOT NULL,
	enabled BOOLEAN DEFAULT FALSE NOT NULL,
	member_aware BOOLEAN DEFAULT FALSE NOT NULL
)
;

CREATE TABLE roles (
	name TEXT PRIMARY KEY,
	description TEXT NOT NULL,
	is_only_global BOOLEAN NOT NULL DEFAULT FALSE
)
;

COMMENT ON COLUMN roles.is_only_global IS 'If true, the role can only be assigned globally'
;

CREATE TABLE role_authorizations (
	role_name TEXT NOT NULL REFERENCES roles (name) ON DELETE CASCADE ON UPDATE CASCADE,
	authorization_name TEXT NOT NULL REFERENCES authorizations (name) ON DELETE CASCADE ON UPDATE CASCADE,
	PRIMARY KEY (role_name, authorization_name)
)
;

CREATE TABLE sessions (
	id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
	user_id UUID NOT NULL,
	expires TIMESTAMP WITH TIME ZONE NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CLOCK_TIMESTAMP()
)
;

CREATE TABLE team_delete_keys (
	key UUID DEFAULT gen_random_uuid () PRIMARY KEY,
	team_slug slug NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
	created_by UUID NOT NULL,
	confirmed_at TIMESTAMP WITH TIME ZONE
)
;

CREATE TABLE team_slugs (slug slug PRIMARY KEY)
;

CREATE TABLE teams (
	slug slug PRIMARY KEY,
	purpose TEXT NOT NULL,
	last_successful_sync TIMESTAMP WITHOUT TIME ZONE,
	delete_key_confirmed_at TIMESTAMPTZ,
	CHECK (
		(
			TRIM(
				BOTH
				FROM
					purpose
			) <> ''::TEXT
		)
	)
)
;

CREATE TABLE service_accounts (
	id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP() NOT NULL,
	name TEXT NOT NULL CONSTRAINT name_length CHECK (CHAR_LENGTH(name) <= 80),
	description TEXT NOT NULL,
	team_slug slug REFERENCES teams (slug) ON DELETE CASCADE
)
;

CREATE TABLE service_account_roles (
	id SERIAL PRIMARY KEY,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP() NOT NULL,
	role_name TEXT NOT NULL REFERENCES roles (name) ON DELETE CASCADE ON UPDATE CASCADE,
	service_account_id UUID NOT NULL REFERENCES service_accounts (id) ON DELETE CASCADE
)
;

CREATE TABLE service_account_tokens (
	id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP() NOT NULL,
	last_used_at TIMESTAMP WITH TIME ZONE,
	expires_at DATE,
	name TEXT NOT NULL CONSTRAINT name_length CHECK (CHAR_LENGTH(name) <= 80),
	description TEXT NOT NULL,
	token TEXT NOT NULL UNIQUE,
	service_account_id UUID NOT NULL REFERENCES service_accounts (id) ON DELETE CASCADE
)
;

CREATE TABLE user_roles (
	id SERIAL PRIMARY KEY,
	role_name TEXT NOT NULL,
	user_id UUID NOT NULL,
	target_team_slug slug
)
;

CREATE TABLE users (
	id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
	email TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	external_id TEXT NOT NULL UNIQUE,
	admin BOOLEAN NOT NULL DEFAULT FALSE
)
;

CREATE TABLE usersync_log_entries (
       id UUID DEFAULT gen_random_uuid () PRIMARY KEY,
       created_at TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP() NOT NULL,
       action usersync_log_entry_action NOT NULL,
       user_id UUID NOT NULL,
       user_name TEXT NOT NULL,
       user_email TEXT NOT NULL,
       old_user_name TEXT,
       old_user_email TEXT,
       role_name TEXT
)
;

-- views

-- additional indexes
CREATE INDEX activity_log_entries_team_slug_idx ON activity_log_entries (team_slug)
;

CREATE INDEX activity_log_entries_resource_type_idx ON activity_log_entries (resource_type)
;

CREATE INDEX activity_log_entries_created_at_idx ON activity_log_entries (created_at)
;

CREATE INDEX ON reconciler_errors USING btree (created_at DESC)
;

CREATE UNIQUE INDEX ON service_accounts USING btree (name, team_slug) NULLS NOT DISTINCT
;

CREATE UNIQUE INDEX ON service_account_roles USING btree (service_account_id, role_name)
;

CREATE UNIQUE INDEX ON service_account_tokens USING btree (service_account_id, name)
;


CREATE INDEX ON teams (delete_key_confirmed_at)
;

CREATE UNIQUE INDEX ON user_roles USING btree (user_id, role_name)
WHERE
	(
		(target_team_slug IS NULL)
	)
;

CREATE UNIQUE INDEX ON user_roles USING btree (user_id, role_name, target_team_slug)
WHERE
	(target_team_slug IS NOT NULL)
;

-- foreign keys
ALTER TABLE reconciler_config
ADD FOREIGN KEY (reconciler) REFERENCES reconcilers (name) ON DELETE CASCADE
;

ALTER TABLE reconciler_errors
ADD FOREIGN KEY (reconciler) REFERENCES reconcilers (name) ON DELETE CASCADE,
ADD FOREIGN KEY (team_slug) REFERENCES teams (slug) ON DELETE CASCADE
;

ALTER TABLE reconciler_states
ADD FOREIGN KEY (reconciler_name) REFERENCES reconcilers (name) ON DELETE CASCADE,
ADD FOREIGN KEY (team_slug) REFERENCES teams (slug) ON DELETE CASCADE
;

ALTER TABLE sessions
ADD FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
;

ALTER TABLE team_delete_keys
ADD FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE CASCADE,
ADD FOREIGN KEY (team_slug) REFERENCES teams (slug) ON DELETE CASCADE
;

ALTER TABLE user_roles
ADD FOREIGN KEY (target_team_slug) REFERENCES teams (slug) ON DELETE CASCADE,
ADD FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
ADD FOREIGN KEY (role_name) REFERENCES roles (name) ON DELETE CASCADE ON UPDATE CASCADE
;

-- triggers
CREATE TRIGGER reconciler_states_set_updated BEFORE
UPDATE ON reconciler_states FOR EACH ROW
EXECUTE PROCEDURE set_updated_at ()
;

CREATE TRIGGER unique_team_slug BEFORE INSERT ON teams FOR EACH ROW
EXECUTE PROCEDURE unique_team_slug ()
;

CREATE TRIGGER register_slug
AFTER INSERT ON teams FOR EACH ROW
EXECUTE PROCEDURE register_team_slug ()
;

CREATE TRIGGER service_accounts_set_updated BEFORE
UPDATE ON service_accounts FOR EACH ROW
EXECUTE PROCEDURE set_updated_at ()
;

CREATE TRIGGER service_account_tokens_set_updated BEFORE
UPDATE ON service_account_tokens FOR EACH ROW
EXECUTE PROCEDURE set_updated_at ()
;

CREATE TRIGGER teams_notify
AFTER INSERT
OR
UPDATE
OR DELETE ON teams FOR EACH ROW
EXECUTE PROCEDURE api_notify ("slug", "purpose")
;

-- initial values
INSERT INTO
	authorizations (name, description)
VALUES
	(
		'activity_logs:read',
		'Permission to read activity logs.'
	),
	(
		'service_accounts:create',
		'Permission to create service accounts.'
	),
	(
		'service_accounts:delete',
		'Permission to delete service accounts.'
	),
	(
		'service_accounts:read',
		'Permission to read service accounts.'
	),
	(
		'service_accounts:update',
		'Permission to update service accounts.'
	),
	('teams:create', 'Permission to create teams.'),
	('teams:delete', 'Permission to delete teams.'),
	(
		'teams:metadata:update',
		'Permission to update team metadata.'
	),
	(
		'teams:members:admin',
		'Permission to administer team members.'
	)
;

INSERT INTO
	roles (name, description, is_only_global)
VALUES
	(
		'Service account owner',
		'Permits the actor to manage service accounts.',
		FALSE
	),
	(
		'Team creator',
		'Permits the actor to create teams.',
		TRUE
	),
	(
		'Team member',
		'Permits the actor to do actions on behalf of a team. Also includes managing most team resources except members.',
		FALSE
	),
	(
		'Team owner',
		'Permits the actor to do actions on behalf of a team. Also includes managing all team resources, including members.',
		FALSE
	)
;

INSERT INTO
	role_authorizations (role_name, authorization_name)
VALUES
	(
		'Service account owner',
		'service_accounts:create'
	),
	(
		'Service account owner',
		'service_accounts:delete'
	),
	('Service account owner', 'service_accounts:read'),
	(
		'Service account owner',
		'service_accounts:update'
	),
	('Team creator', 'teams:create'),
	('Team member', 'teams:metadata:update'),
	('Team member', 'service_accounts:create'),
	('Team member', 'service_accounts:delete'),
	('Team member', 'service_accounts:read'),
	('Team member', 'service_accounts:update'),
	('Team owner', 'teams:delete'),
	('Team owner', 'teams:metadata:update'),
	('Team owner', 'teams:members:admin'),
	('Team owner', 'service_accounts:create'),
	('Team owner', 'service_accounts:delete'),
	('Team owner', 'service_accounts:read'),
	('Team owner', 'service_accounts:update')
;
