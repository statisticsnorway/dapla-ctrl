-- name: ListMembers :many
SELECT
	users.id,
	groups.team_slug,
	ARRAY_AGG(groups.name)::TEXT[] AS groups,
	COUNT(*) OVER () AS total_count
FROM
	group_members
	JOIN groups ON groups.name = group_members.group_name
	JOIN users ON users.id = group_members.user_id
WHERE
	groups.team_slug = @team_slug::slug
GROUP BY
	groups.team_slug,
	users.id
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'name:asc' THEN LOWER(users.name)
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'name:desc' THEN LOWER(users.name)
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'email:asc' THEN users.email
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'email:desc' THEN users.email
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'section_code:asc' THEN users.section_code
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'section_code:desc' THEN users.section_code
	END DESC,
	users.name,
	users.email ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: ListForUser :many
SELECT
	teams.slug,
	@user_id::UUID AS user_id,
	ARRAY_AGG(groups.name)::TEXT[] AS groups,
	COUNT(*) OVER () AS total_count
FROM
	teams
	LEFT JOIN groups ON groups.team_slug = teams.slug
	LEFT JOIN group_members ON group_members.group_name = groups.name
	LEFT JOIN users ON users.id = group_members.user_id
	LEFT JOIN sections ON teams.section_code = sections.code
	LEFT JOIN user_roles ON user_roles.target_team_slug = teams.slug
WHERE
	users.id = @user_id
	OR sections.manager_id = @user_id
	OR user_roles.user_id = @user_id
GROUP BY
	teams.slug
ORDER BY
	CASE
		WHEN @order_by::TEXT = 'slug:asc' THEN teams.slug
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'slug:desc' THEN teams.slug
	END DESC,
	CASE
		WHEN @order_by::TEXT = 'section_code:asc' THEN teams.section_code
	END ASC,
	CASE
		WHEN @order_by::TEXT = 'section_code:desc' THEN teams.section_code
	END DESC,
	teams.slug ASC
LIMIT
	sqlc.arg('limit')
OFFSET
	sqlc.arg('offset')
;

-- name: UserIsMember :one
SELECT
	(
		EXISTS (
			SELECT
				1
			FROM
				group_members
				JOIN groups ON groups.name = group_members.group_name
			WHERE
				user_id = @user_id
				AND groups.team_slug = @team_slug::slug
		)
		OR EXISTS (
			SELECT
				1
			FROM
				sections
				JOIN teams ON teams.section_code = sections.code
			WHERE
				teams.slug = @team_slug::slug
				AND sections.manager_id = @user_id
		)
	)::BOOLEAN
;

-- name: UserIsOwner :one
SELECT
	(
		EXISTS (
			SELECT
				1
			FROM
				group_members
				JOIN groups ON groups.name = group_members.group_name
			WHERE
				group_members.user_id = @user_id
				AND groups.team_slug = @team_slug::slug
				AND groups.category = 'managers'
				AND groups.suffix = ''
		)
		OR EXISTS (
			SELECT
				1
			FROM
				sections
				JOIN teams ON teams.section_code = sections.code
			WHERE
				teams.slug = @team_slug::slug
				AND sections.manager_id = @user_id
		)
	)::BOOLEAN
;
