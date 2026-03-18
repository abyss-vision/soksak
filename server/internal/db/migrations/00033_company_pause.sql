-- +goose Up
ALTER TABLE "companies" ADD COLUMN "pause_reason" text;
ALTER TABLE "companies" ADD COLUMN "paused_at" timestamp with time zone;

-- +goose Down
ALTER TABLE "companies" DROP COLUMN IF EXISTS "paused_at";
ALTER TABLE "companies" DROP COLUMN IF EXISTS "pause_reason";
