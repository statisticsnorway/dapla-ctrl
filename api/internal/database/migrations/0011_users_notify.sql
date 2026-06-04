-- +goose Up
CREATE TRIGGER users_notify
AFTER INSERT OR UPDATE OR DELETE ON users FOR EACH ROW
EXECUTE PROCEDURE api_notify ("id", "name", "email")
;
