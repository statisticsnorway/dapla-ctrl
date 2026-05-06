-- +goose Up
ALTER TABLE messages
DROP CONSTRAINT IF EXISTS messages_recipient_fkey
;

ALTER TABLE messages
ADD CONSTRAINT messages_recipient_fkey FOREIGN KEY (recipient) REFERENCES users (id) ON DELETE CASCADE
;
