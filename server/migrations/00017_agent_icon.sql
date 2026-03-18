-- +goose Up
ALTER TABLE "agents" ADD COLUMN "icon" text;

-- +goose Down
ALTER TABLE "agents" DROP COLUMN IF EXISTS "icon";
