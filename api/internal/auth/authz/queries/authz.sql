-- name: IsSectionManager :one
SELECT
	(
		EXISTS (
			SELECT
				1
			FROM
				sections
			WHERE
				manager_id = @user_id
		)
	)
;

-- name: IsGlobalAdmin :one
SELECT
	(
		EXISTS (
			SELECT
				1
			FROM
				users
			WHERE
				id = @user_id
				AND admin = TRUE
		)
	)
;

-- name: IsManagerForTeamSection :one
SELECT
	(
		(
			EXISTS (
				SELECT
					1
				FROM
					sections
					JOIN teams ON teams.section_code = sections.code
				WHERE
					sections.manager_id = @user_id
					AND teams.slug = @team_slug::slug
			)
		)
	)
;
