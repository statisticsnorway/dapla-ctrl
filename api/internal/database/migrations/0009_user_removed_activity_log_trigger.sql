-- +goose Up
-- Add trigger to create activity log when a user is deleted (e.g. by usersync)
-- When the user is deleted, we should create an activity log for each
-- group the user was a member of
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION log_user_deletion () RETURNS TRIGGER AS $$
BEGIN
		INSERT INTO activity_log_entries (actor, action, resource_type, resource_name, team_slug, data)
		SELECT DISTINCT
				'system' AS actor,
				'REMOVED' AS action,
				'GROUP' AS resource_type,
				gs.name AS resource_name,
				gs.team_slug AS team_slug,
				jsonb_build_object(
						'userID', OLD.id,
						'userEmail', OLD.email
				)::text::bytea AS data
		FROM group_members gm JOIN groups gs ON gm.group_name = gs.name
		WHERE gm.user_id = OLD.id;

		RETURN OLD;
END;
$$ LANGUAGE plpgsql
;

-- +goose StatementEnd
DROP TRIGGER IF EXISTS user_deletion_trigger ON users
;

CREATE TRIGGER user_deletion_trigger
BEFORE DELETE ON users FOR EACH ROW
EXECUTE FUNCTION log_user_deletion ()
;
