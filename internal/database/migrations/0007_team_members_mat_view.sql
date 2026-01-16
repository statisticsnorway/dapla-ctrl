-- +goose Up
CREATE MATERIALIZED VIEW team_members AS
SELECT DISTINCT
	ON (teams.slug, group_members.user_id) teams.slug AS team_slug,
	group_members.user_id
FROM
	teams
	JOIN groups ON groups.team_slug = teams.slug
	JOIN group_members ON groups.name = group_members.group_name
;
