-- +goose Up
ALTER TABLE companies ADD COLUMN communication_language VARCHAR(5) DEFAULT NULL;

-- +goose Down
ALTER TABLE companies DROP COLUMN communication_language;
