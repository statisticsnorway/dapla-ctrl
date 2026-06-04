-- +goose Up
ALTER TABLE teams
DROP COLUMN purpose
;

DROP TRIGGER teams_notify ON teams
;

CREATE TRIGGER teams_notify
AFTER INSERT OR UPDATE OR DELETE ON teams FOR EACH ROW
EXECUTE PROCEDURE api_notify ("slug", "section_code")
;
