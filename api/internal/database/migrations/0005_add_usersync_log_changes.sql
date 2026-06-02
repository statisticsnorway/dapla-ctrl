-- +goose Up
ALTER TABLE usersync_log_entries
ADD COLUMN IF NOT EXISTS changes JSONB
;

UPDATE usersync_log_entries AS ule
SET
	changes = COALESCE(
		(
			SELECT
				JSONB_OBJECT_AGG(key, value)
			FROM
				LATERAL (
					VALUES
						(
							'name',
							CASE
								WHEN ule.old_user_name IS NOT NULL
								AND ule.old_user_name != ule.user_name THEN JSONB_BUILD_OBJECT('old', ule.old_user_name, 'new', ule.user_name)
							END
						),
						(
							'email',
							CASE
								WHEN ule.old_user_email IS NOT NULL
								AND ule.old_user_email != ule.user_email THEN JSONB_BUILD_OBJECT('old', ule.old_user_email, 'new', ule.user_email)
							END
						)
				) AS t (key, value)
			WHERE
				value IS NOT NULL
		),
		'{}'::JSONB
	)
WHERE
	ule.changes IS NULL
	AND ule.action = 'update_user'
;

-- +goose Down
ALTER TABLE usersync_log_entries
DROP COLUMN IF EXISTS changes
;
