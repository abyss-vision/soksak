-- +goose Up
ALTER TABLE "issues" ADD COLUMN "hidden_at" timestamp with time zone;

-- +goose Down
ALTER TABLE "issues" DROP COLUMN IF EXISTS "hidden_at";
