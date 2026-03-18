-- +goose Up
ALTER TABLE "projects" ADD COLUMN "color" text;
ALTER TABLE "projects" ADD COLUMN "archived_at" timestamp with time zone;

-- +goose Down
ALTER TABLE "projects" DROP COLUMN IF EXISTS "archived_at";
ALTER TABLE "projects" DROP COLUMN IF EXISTS "color";
